// Package syncer implements the syncer for synchronization
// of /var/lib/kubelet/pods.
package syncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/getoutreach/devenv/pkg/kube"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	corev1 "k8s.io/api/core/v1"

	// needed for auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type lightPod struct {
	Name      string
	Namespace string
}

type Syncer struct {
	fromPath string
	toPath   string

	log   logrus.FieldLogger
	k     kubernetes.Interface
	rconf *rest.Config

	podCache   map[string]lightPod
	podCacheMu sync.RWMutex
}

// NewSyncer creates bind mounts from to -> from based on changes
// in the from directory.
func NewSyncer(from, to string, log logrus.FieldLogger) *Syncer {
	k, conf, err := kube.GetKubeClientWithConfig()
	if err != nil {
		panic(err)
	}

	return &Syncer{
		fromPath: from,
		toPath:   to,
		log:      log,
		k:        k,
		rconf:    conf,
		podCache: make(map[string]lightPod),
	}
}

func (s *Syncer) onAdded(file string) error {
	id, err := uuid.Parse(filepath.Base(file))
	if err != nil {
		return errors.Wrap(err, "unlikely to be pod, name wasn't a UUID")
	}

	po, ok := s.podCache[id.String()]
	if !ok {
		return fmt.Errorf("pod wasn't found in cache")
	}

	s.log.WithField("pod.key", s.getPodKey(po)).Info("found new pod directory")

	return nil
}

func (s *Syncer) getPodKey(inf interface{}) string {
	name := ""
	namespace := ""
	switch po := inf.(type) {
	case *corev1.Pod:
		name = po.Name
		namespace = po.Namespace
	case lightPod:
		name = po.Name
		namespace = po.Namespace
	}

	return namespace + "/" + name
}

func (s *Syncer) onRemoved(file string) error {
	return nil
}

// startInformer starts an informer that updates a pod uid -> key cache
func (s *Syncer) startInformer(ctx context.Context) {
	// TODO: if this gets real, use a worker queue here
	inf := informers.NewSharedInformerFactoryWithOptions(s.k, 5*time.Minute).
		Core().V1().Pods().Informer()
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			po, ok := obj.(*corev1.Pod)
			if !ok {
				s.log.WithField("event.type", reflect.TypeOf(po).String()).Warn("skipping event")
				return
			}

			s.log.WithField("pod.uid", po.ObjectMeta.UID).
				WithField("pod.key", s.getPodKey(po)).
				Info("observed pod creation")

			s.podCacheMu.Lock()
			s.podCache[string(po.ObjectMeta.UID)] = lightPod{
				Name:      po.Name,
				Namespace: po.Namespace,
			}
			s.podCacheMu.Unlock()
		},
		DeleteFunc: func(obj interface{}) {
			po, ok := obj.(*corev1.Pod)
			if !ok {
				s.log.WithField("event.type", reflect.TypeOf(po).String()).Warn("skipping event")
				return
			}

			s.log.WithField("pod.uid", po.ObjectMeta.UID).
				WithField("pod.key", s.getPodKey(po)).
				Info("observed pod deletion")

			s.podCacheMu.Lock()
			delete(s.podCache, string(po.ObjectMeta.UID))
			s.podCacheMu.Unlock()
		},
	})

	// start the informer
	go inf.Run(ctx.Done())

	s.log.Info("started informer")
}

// Start starts the syncer.
func (s *Syncer) Start(ctx context.Context) error { //nolint:funlen
	if _, err := os.Stat(s.fromPath); err != nil {
		return errors.Wrapf(err, "failed to access source path '%s'", s.fromPath)
	}

	if _, err := os.Stat(s.toPath); os.IsNotExist(err) {
		s.log.WithField("destination", s.toPath).Info("creating destination path")
		err = os.MkdirAll(s.toPath, 0755)
		if err != nil {
			return errors.Wrapf(err, "failed to create destination path '%s'", s.toPath)
		}
	}

	s.startInformer(ctx)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "failed to create watcher")
	}
	defer w.Close()

	// TODO: queue these, with a retry
	go func() {
		defer func() {
			s.log.WithError(ctx.Err()).Warn("watcher stopped")
		}()

		for ctx.Err() == nil {
			select {
			case event := <-w.Events:
				var err error     //nolint:govet // Why: we're OK shadowing err
				switch event.Op { //nolint:exhaustive
				case fsnotify.Create:
					err = s.onAdded(event.Name)
				case fsnotify.Remove:
					err = s.onRemoved(event.Name)
				}
				if err != nil {
					s.log.WithError(err).Warn("failed to process file change event")
				}
			case err := <-w.Errors: //nolint:govet // Why: We're OK shadowing err
				s.log.WithError(err).Warn("failed to watch file change")
			}
		}
	}()

	err = w.Add(s.fromPath)
	if err != nil {
		return errors.Wrapf(err, "failed to start watching '%s'", s.fromPath)
	}

	s.log.WithField("file.path", s.fromPath).Info("started filesystem watcher")

	return nil
}
