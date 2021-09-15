// Package syncer implements the syncer for synchronization
// of /var/lib/kubelet/pods.
package syncer

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

type vclusterPod struct {
	corev1.Pod

	VCPodInfo *vclusterPodInfo
}

type vclusterPodInfo struct {
	ClusterName string
	Name        string
	Namespace   string
	UID         string
}

type Syncer struct {
	fromPath string
	toPath   string

	log   logrus.FieldLogger
	k     kubernetes.Interface
	rconf *rest.Config

	podCache   map[string]vclusterPod
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
		podCache: make(map[string]vclusterPod),
	}
}

func (s *Syncer) onAdded(file string) error {
	id, err := uuid.Parse(filepath.Base(file))
	if err != nil {
		return errors.Wrap(err, "unlikely to be pod, name wasn't a UUID")
	}

	if inf, err := os.Stat(filepath.Join(s.fromPath, file)); err != nil || !inf.IsDir() {
		return fmt.Errorf("failed to read pod dir, or isn't a directory")
	}

	po, ok := s.podCache[id.String()]
	if !ok {
		return fmt.Errorf("pod wasn't found in cache")
	}

	// only process pods in a vcluster
	if !strings.HasPrefix(po.Namespace, "vcluster-") {
		return fmt.Errorf("pod wasn't in a vcluster")
	}

	s.log.WithField("pod.key", s.getPodKey(po)).Info("found new pod directory")

	s.log.
		WithField("pod.key", s.getPodKey(po.VCPodInfo)).
		WithField("pod.uid", po.VCPodInfo.UID).
		WithField("vcluster.name", po.VCPodInfo.ClusterName).
		Info("retrieved vcluster pod information")

	fromPath := filepath.Join(s.fromPath, file)
	toPath := filepath.Join(s.toPath, po.VCPodInfo.ClusterName,
		"kubelet", "pods", po.VCPodInfo.UID)

	//nolint:errcheck // Why: Will fix tomorrow
	os.MkdirAll(filepath.Dir(toPath), 0755)

	s.log.WithField("from", fromPath).WithField("to", toPath).
		Info("mounting vcluster pod")
	return bindMount(fromPath, toPath)
}

func (s *Syncer) getPodKey(inf interface{}) string {
	name := ""
	namespace := ""

	infValue := reflect.ValueOf(inf)
	if infValue.Kind() == reflect.Ptr {
		if !infValue.IsNil() {
			inf = infValue.Elem().Interface()
		}
	}

	switch po := inf.(type) {
	case corev1.Pod:
		name = po.Name
		namespace = po.Namespace
	case vclusterPod:
		name = po.Name
		namespace = po.Namespace
	case vclusterPodInfo:
		name = po.Name
		namespace = po.Namespace
	}

	return namespace + "/" + name
}

func (s *Syncer) onRemoved(file string) error {
	return nil
}

func (s *Syncer) getVClusterPod(po *corev1.Pod) (*vclusterPod, error) {
	vcPodName, ok := po.ObjectMeta.Annotations["vcluster.loft.sh/name"]
	if !ok {
		return nil, fmt.Errorf("missing name")
	}

	vcPodNamespace, ok := po.ObjectMeta.Annotations["vcluster.loft.sh/namespace"]
	if !ok {
		return nil, fmt.Errorf("missing namespace")
	}

	uid, ok := po.ObjectMeta.Annotations["vcluster.loft.sh/uid"]
	if !ok {
		return nil, fmt.Errorf("missing uid")
	}

	vcName, ok := po.ObjectMeta.Labels["vcluster.loft.sh/managed-by"]
	if !ok {
		return nil, fmt.Errorf("missing managed-by")
	}

	return &vclusterPod{*po, &vclusterPodInfo{
		ClusterName: vcName,
		Name:        vcPodName,
		Namespace:   vcPodNamespace,
		UID:         uid,
	}}, nil
}

// startInformer starts an informer that updates a pod uid -> key cache
func (s *Syncer) startInformer(ctx context.Context) error {
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
				Debug("observed pod creation")

			vpo, err := s.getVClusterPod(po)
			if err != nil {
				return
			}

			s.log.WithField("pod.uid", po.ObjectMeta.UID).
				WithField("pod.key", s.getPodKey(vpo)).
				Info("observed vcluster pod creation")

			s.podCacheMu.Lock()
			s.podCache[string(po.ObjectMeta.UID)] = *vpo
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
				Debug("observed pod deletion")

			vpo, err := s.getVClusterPod(po)
			if err != nil {
				return
			}

			s.log.WithField("pod.uid", po.ObjectMeta.UID).
				WithField("pod.key", s.getPodKey(vpo)).
				Info("observed vcluster pod deletion")

			s.podCacheMu.Lock()
			delete(s.podCache, string(po.ObjectMeta.UID))
			s.podCacheMu.Unlock()
		},
	})

	// start the informer
	go inf.Run(ctx.Done())

	cctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()
	if !cache.WaitForCacheSync(cctx.Done(), inf.HasSynced) {
		return fmt.Errorf("failed to sync cache")
	}

	s.log.Info("started informer")
	return nil
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

	// TODO: need to close this now that it isn't blocking
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "failed to create watcher")
	}

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
					s.log.WithError(err).WithField("file.name", event.Name).
						Warn("failed to process file change event")
				}
			case err := <-w.Errors: //nolint:govet // Why: We're OK shadowing err
				s.log.WithError(err).Warn("failed to watch file change")
			}
		}
	}()

	dirs, err := ioutil.ReadDir(s.fromPath)
	if err != nil {
		return errors.Wrap(err, "failed to read initial pods")
	}

	for _, fileName := range dirs {
		err := s.onAdded(fileName.Name()) //nolint:govet // Why: we're OK shadowing err
		if err != nil {
			s.log.WithError(err).WithField("file.name", fileName.Name()).
				Warn("failed to process initial pod")
		}
	}

	err = w.Add(s.fromPath)
	if err != nil {
		return errors.Wrapf(err, "failed to start watching '%s'", s.fromPath)
	}

	s.log.WithField("file.path", s.fromPath).Info("started filesystem watcher")

	return nil
}
