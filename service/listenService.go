package service

import (
	"container_manager/controller"
	"context"
	v1 "github.com/qinsheng99/crdcode/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	"sync"
	"time"
)

var serverUnusable = map[v1.ServerConditionType]struct{}{
	v1.ServerRecycled: {},
	v1.ServerInactive: {},
	v1.ServerErrored:  {},
}

var serverUsable = map[v1.ServerConditionType]struct{}{
	v1.ServerCreated: {},
	v1.ServerReady:   {},
	v1.ServerBound:   {},
}

type ListenInter interface {
	ListenResource()
}

type Listen struct {
	wg       *sync.WaitGroup
	res      *kubernetes.Clientset
	resync   time.Duration
	mux      *sync.Mutex
	config   *rest.Config
	dym      dynamic.Interface
	resource schema.GroupVersionResource
}

type StatusDetail struct {
	IsUsable  bool   `json:"is_usable"`
	AccessUrl string `json:"access_url,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

type InferenceRequest struct {
	controller.InferenceInfo
	Status StatusDetail
}

func NewListen(res *kubernetes.Clientset, c *rest.Config, dym dynamic.Interface, resource schema.GroupVersionResource) ListenInter {
	return &Listen{res: res, wg: &sync.WaitGroup{}, mux: &sync.Mutex{}, config: c, dym: dym, resource: resource}
}

func (l *Listen) ListenResource() {
	log.Println("listen k8s resource for crd")
	infor := l.crdConfig()
	infor.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: l.Update,
		DeleteFunc: l.Delete,
		AddFunc:    l.Add,
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	infor.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, infor.HasSynced) {
		log.Println("cache sync err")
		return
	}

	<-stopCh
}

func (l *Listen) Update(oldObj, newObj interface{}) {
	var res v1.CodeServer

	bys, err := json.Marshal(newObj)
	if err != nil {
		log.Println("marshal error:", err.Error())
		return
	}
	err = json.Unmarshal(bys, &res)
	if err != nil {
		log.Println("unmarshal error:", err.Error())
		return
	}

	go l.dispatcher(res)
}

func (l *Listen) dispatcher(res v1.CodeServer) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("dispatcher panic:", err)
		}
	}()

	status := l.transferStatus(res)
	switch res.Labels["type"] {
	case controller.MetaNameInference:

	}

	log.Println("data:", status)
}

func (l *Listen) transferStatus(res v1.CodeServer) (req StatusDetail) {
	var endPoint string
	for _, condition := range res.Status.Conditions {
		if _, ok := serverUnusable[condition.Type]; ok {
			if condition.Status == corev1.ConditionTrue {
				req.ErrorMsg = condition.Reason
				return
			}
		}

		if _, ok := serverUsable[condition.Type]; ok {
			if condition.Status == corev1.ConditionFalse {
				req.ErrorMsg = condition.Reason
				return
			}
		}
		if endPoint == "" {
			endPoint = condition.Message["instanceEndpoint"]
		}
	}
	req.IsUsable = true
	req.AccessUrl = endPoint
	return
}

func (l *Listen) Delete(obj interface{}) {

}

func (l *Listen) Add(obj interface{}) {

}

func (l *Listen) crdConfig() cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return l.dym.Resource(l.resource).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return l.dym.Resource(l.resource).Watch(context.TODO(), options)
			},
		},
		&unstructured.Unstructured{},
		0,
		cache.Indexers{},
	)
}
