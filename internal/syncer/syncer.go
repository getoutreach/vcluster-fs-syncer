// Package syncer implements the syncer for synchronization
// of /var/lib/kubelet/pods.
package syncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/getoutreach/devenv/pkg/kube"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	corev1 "k8s.io/api/core/v1"

	// needed for auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type event struct {
	pod *corev1.Pod

	// added or deleted
	event string
}

type vclusterPod struct {
	corev1.Pod

	VCPodInfo *vclusterPodInfo
}

type vclusterPodInfo struct {
	ClusterName string
	Name        string
	Namespace   string
	UID         string

	Deleted   bool
	ExpiresAt time.Time
}

type Syncer struct {
	fromPath string
	toPath   string

	log   logrus.FieldLogger
	k     kubernetes.Interface
	rconf *rest.Config

	queue       workqueue.RateLimitingInterface
	threadiness int
}

// NewSyncer creates bind mounts from to -> from based on changes
// in the from directory.
func NewSyncer(from, to string, log logrus.FieldLogger) *Syncer {
	k, conf, err := kube.GetKubeClientWithConfig()
	if err != nil {
		panic(err)
	}

	return &Syncer{
		fromPath:    from,
		toPath:      to,
		log:         log,
		k:           k,
		rconf:       conf,
		queue:       workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(time.Second*1, time.Minute*1)),
		threadiness: 1,
	}
}

func (s *Syncer) onAdded(vpo *vclusterPod) error {
	hostPodPath := filepath.Join(s.fromPath, string(vpo.UID))
	vclusterPodPath := filepath.Join(s.toPath, vpo.VCPodInfo.ClusterName,
		"kubelet", "pods", vpo.VCPodInfo.UID)

	if inf, err := os.Stat(hostPodPath); err != nil || !inf.IsDir() {
		return fmt.Errorf("failed to read pod dir, or isn't a directory")
	}

	if _, err := os.Stat(vclusterPodPath); err == nil {
		// we've already processed this pod before
		s.log.Info("skipping pod mount, already mounted")
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(vclusterPodPath), 0755); err != nil {
		return errors.Wrap(err, "failed to create pod directory")
	}

	s.log.WithField("from", hostPodPath).WithField("to", vclusterPodPath).
		Info("mounting vcluster pod")
	return bindMount(hostPodPath, vclusterPodPath)
}

func (s *Syncer) onRemoved(vpo *vclusterPod) error {
	toPath := filepath.Join(s.toPath, vpo.VCPodInfo.ClusterName,
		"kubelet", "pods", vpo.VCPodInfo.UID)

	s.log.WithField("pod.path", toPath).
		Info("unmounting vcluster pod")

	if _, err := os.Stat(toPath); os.IsNotExist(err) {
		s.log.WithField("pod.path", toPath).Warn("not cleaning up directory, didn't exist")
		return nil
	}

	return unmountBind(toPath)
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

// getVClusterPod returns pod with information about the associated
// vcluster attached to it
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
	inf := informers.NewSharedInformerFactoryWithOptions(s.k, 5*time.Minute).
		Core().V1().Pods().Informer()
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			po, ok := obj.(*corev1.Pod)
			if !ok {
				s.log.WithField("event.type", reflect.TypeOf(po).String()).Warn("skipping event")
				return
			}

			s.queue.Add(&event{
				pod:   po,
				event: "added",
			})
		},
		UpdateFunc: func(oldObj, obj interface{}) {
			po, ok := obj.(*corev1.Pod)
			if !ok {
				s.log.WithField("event.type", reflect.TypeOf(po).String()).Warn("skipping event")
				return
			}

			s.queue.Add(&event{
				pod:   po,
				event: "updated",
			})
		},
		DeleteFunc: func(obj interface{}) {
			po, ok := obj.(*corev1.Pod)
			if !ok {
				s.log.WithField("event.type", reflect.TypeOf(po).String()).Warn("skipping event")
				return
			}

			s.queue.Add(&event{
				pod:   po,
				event: "deleted",
			})
		},
	})

	// start the informer
	go inf.Run(ctx.Done())

	cctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()
	if !cache.WaitForCacheSync(cctx.Done(), inf.HasSynced) {
		return fmt.Errorf("failed to sync cache")
	}

	s.log.Info("Informer started and synced")
	return nil
}

// Start starts the syncer.
func (s *Syncer) Start(ctx context.Context) error { //nolint:funlen
	s.log.Infof("Starting %d proxier worker(s)", s.threadiness)
	for i := 0; i < s.threadiness; i++ {
		go wait.Until(s.runWorker, time.Second, ctx.Done())
	}

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

	s.startInformer(ctx) //nolint:errcheck // Why: uneeded

	<-ctx.Done()

	return nil
}

func (s *Syncer) reconcile(e *event) error {
	if e.pod.Spec.NodeName == "" {
		// pod wasn't scheduled, ignore it for now.
		return nil
	}

	if e.pod.Spec.NodeName != os.Getenv("MY_NODE_NAME") {
		// pod wasn't scheduled onto our node, skip for now.
		return nil
	}

	vpo, err := s.getVClusterPod(e.pod)
	if err != nil {
		// pod wasn't a vcluster pod, ignore it
		return nil
	}

	fields := logrus.Fields{
		"pod.uid":    e.pod.ObjectMeta.UID,
		"pod.key":    s.getPodKey(e.pod),
		"vpod.uid":   vpo.VCPodInfo.UID,
		"vpod.key":   s.getPodKey(vpo),
		"event.type": e.event,
	}

	s.log.WithFields(fields).Info("observed vcluster pod event")

	switch e.event {
	case "added", "updated":
		err = s.onAdded(vpo)
	case "deleted":
		err = s.onRemoved(vpo)
	default:
		err = fmt.Errorf("unknown event %s", e.event)
	}
	if err != nil {
		s.log.WithError(err).WithFields(fields).Error("failed to process vcluster pod event")
	}
	return err
}

func (s *Syncer) Close() error {
	s.log.Info("Shutting down syncer")

	s.queue.ShutDown()

	return errors.Wrap(filepath.Walk(s.toPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non directories
		if !info.IsDir() {
			return nil
		}

		_, err = uuid.Parse(filepath.Base(path))
		if err != nil {
			// skip files that aren't UUID
			return nil
		}

		s.log.WithField("pod.path", path).Info("cleaning up mount")

		err = unmountBind(path)
		if err != nil {
			s.log.WithError(err).WithField("pod.path", s.toPath).Warn("failed to remove bind mount")
		}

		// we just removed the directory, so do not attempt to walk it
		return filepath.SkipDir
	}), "failed to cleanup mounts")
}
