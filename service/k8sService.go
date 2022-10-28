package service

import (
	"bytes"
	"container_manager/client"
	"container_manager/tools"
	"context"
	"errors"
	"github.com/qinsheng99/crdcode/api/v1"
	"io"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/dynamic"
)

type ResListStatus struct {
	ServerCreatedFlag  bool
	ServerReadyFlag    bool
	ServerInactiveFlag bool
	ServerRecycledFlag bool
	ServerErroredFlag  bool
	ServerBoundFlag    bool
	ServerCreatedTime  string
	ServerReadyTime    string
	ServerBoundTime    string
	ServerInactiveTime string
	ServerRecycledTime string
	ServerErroredTime  string
	InstanceEndpoint   string
}

type ParamInter interface {
	GeneMetaName() string
}

type K8sService struct {
	p ParamInter
}

func NewK8sService(p ParamInter) *K8sService {
	return &K8sService{p: p}
}

func (s K8sService) Create() (interface{}, error) {
	cli := client.GetDyna()
	resource, err, res := s.getResource()
	if err != nil {
		return nil, err
	}

	dr := cli.Resource(resource).Namespace("default")

	create, err := dr.Create(context.TODO(), res, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	rls := s.validation(create, dr, res)
	if rls.ServerReadyFlag && !rls.ServerRecycledFlag && rls.ServerBoundFlag && !rls.ServerInactiveFlag {
		return rls.InstanceEndpoint, nil
	}

	if rls.ServerRecycledFlag {
		return nil, errors.New("overdue")
	}

	if rls.ServerInactiveFlag || rls.ServerErroredFlag {
		label := "app=" + create.GetName()
		podList, err := client.GetClient().CoreV1().Pods(create.GetNamespace()).List(context.TODO(), metav1.ListOptions{LabelSelector: label})
		if err != nil {
			return nil, err
		}
		var a = new(int64)
		*a = 3
		pod := podList.Items[0]
		logs := client.GetClient().CoreV1().Pods(pod.GetNamespace()).GetLogs(pod.GetName(), &corev1.PodLogOptions{TailLines: a})
		stream, err := logs.Stream(context.TODO())
		if err != nil {
			return nil, err
		}

		var buf = new(bytes.Buffer)
		_, err = io.Copy(buf, stream)
		if err != nil {
			return nil, err
		}

		return buf.String(), nil

	}
	return create, nil

}

func (s K8sService) Update(expiry int) (interface{}, error) {
	cli := client.GetDyna()
	resource, err, _ := s.getResource()
	if err != nil {
		return nil, err
	}

	get, err := cli.Resource(resource).Namespace("default").Get(context.TODO(), s.p.GeneMetaName(), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if sp, ok := get.Object["spec"]; ok {
		if spc, ok := sp.(map[string]interface{}); ok {
			spc["add"] = true
			spc["recycleAfterSeconds"] = expiry
		}
	}

	_, err = cli.Resource(resource).Namespace("default").Update(context.TODO(), get, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s K8sService) getResource() (schema.GroupVersionResource, error, *unstructured.Unstructured) {
	k, err, res := s.resource()
	if err != nil {
		return schema.GroupVersionResource{}, err, nil
	}

	metaData := map[string]interface{}{
		"name": s.p.GeneMetaName(),
	}
	res.Object["metadata"] = metaData

	mapping, err := client.GetrestMapper().RESTMapping(k.GroupKind(), k.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err, nil
	}

	return mapping.Resource, nil, res
}

func (s K8sService) resource() (kind *schema.GroupVersionKind, err error, _ *unstructured.Unstructured) {
	var yamldata []byte
	yamldata, err = ioutil.ReadFile("crd-resource.yaml")
	if err != nil {
		return nil, err, nil
	}
	obj := &unstructured.Unstructured{}
	_, kind, err = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(yamldata, nil, obj)
	if err != nil {
		return nil, err, nil
	}
	return kind, nil, obj
}

func (s *K8sService) validation(code *unstructured.Unstructured, dr dynamic.ResourceInterface, object *unstructured.Unstructured) ResListStatus {
	var err error
	var num int
try:
	rls := ResListStatus{}
	code, err = dr.Get(context.TODO(), code.GetName(), metav1.GetOptions{})
	if err != nil {
		num++
		if num >= 10 {
			err = dr.Delete(context.TODO(), code.GetName(), metav1.DeleteOptions{})
			rls.ServerErroredFlag = true
		}
		goto try
	} else {
		if object.GetAPIVersion() == code.GetAPIVersion() {
			metadata, ok := tools.ParsingMap(code.Object, "metadata")
			if !ok {
				rls.ServerErroredFlag = true
				return rls
			}

			name, nameok := tools.ParsingMapStr(metadata, "name")
			if !nameok {
				rls.ServerErroredFlag = true
				return rls
			}

			if name != object.GetName() {
				rls.ServerErroredFlag = true
				return rls
			}

			status, statusok := tools.ParsingMap(code.Object, "status")
			if !statusok {
				return rls
			}

			conditions, conditionok := tools.ParsingMapSlice(status, "conditions")
			if !conditionok {
				return rls
			}

			for _, condition := range conditions {
				cond := condition.(map[string]interface{})
				typ, typeok := tools.ParsingMapStr(cond, "type")
				if !typeok {
					continue
				}

				switch typ {
				case "ServerCreated": //means the code server has been accepted by the system.
					create, createok := tools.ParsingMapStr(cond, "status")
					if createok && create == "True" {
						rls.ServerCreatedFlag = true
					}
					lastTransitionTime, createtimeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if createtimeok {
						rls.ServerCreatedTime = lastTransitionTime
					}
				case "ServerReady": //means the code server has been ready for usage.
					ready, readyok := tools.ParsingMapStr(cond, "status")
					if readyok && ready == "True" {
						rls.ServerReadyFlag = true
					}
					lastTransitionTime, timeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if timeok {
						rls.ServerReadyTime = lastTransitionTime
					}
					message, messageok := tools.ParsingMap(cond, "message")
					if messageok {
						instanceEndpoint, pointok := tools.ParsingMapStr(message, "instanceEndpoint")
						if pointok {
							rls.InstanceEndpoint = instanceEndpoint
						}
					}
				case "ServerBound": //means the code server has been bound to user.
					bound, boundok := tools.ParsingMapStr(cond, "status")
					if boundok && bound == "True" {
						rls.ServerBoundFlag = true
					}
					lastTransitionTime, boundtimeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if boundtimeok {
						rls.ServerBoundTime = lastTransitionTime
					}
				case "ServerRecycled": //means the code server has been recycled totally.
					recycled, recycleok := tools.ParsingMapStr(cond, "status")
					if recycleok && recycled == "True" {
						rls.ServerRecycledFlag = true
					}
					lastTransitionTime, recycletimeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if recycletimeok {
						rls.ServerRecycledTime = lastTransitionTime
					}
				case "ServerInactive": //means the code server will be marked inactive if `InactiveAfterSeconds` elapsed
					inactive, inactiveok := tools.ParsingMapStr(cond, "status")
					if inactiveok && inactive == "True" {
						rls.ServerInactiveFlag = true
					}
					lastTransitionTime, inactivetimeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if inactivetimeok {
						rls.ServerInactiveTime = lastTransitionTime
					}
				case "ServerErrored": //means failed to reconcile code server.
					bound, errorok := tools.ParsingMapStr(cond, "status")
					if errorok && bound == "True" {
						rls.ServerErroredFlag = true
					}
					lastTransitionTime, errortimeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if errortimeok {
						rls.ServerErroredTime = lastTransitionTime
					}
				}
			}
		}
	}
	return rls
}

func (s K8sService) newValidation(code *unstructured.Unstructured, dr dynamic.ResourceInterface, object *unstructured.Unstructured) ResListStatus {
	var (
		err error
		num int
		bys []byte
	)

try:
	rls := ResListStatus{}
	code, err = dr.Get(context.TODO(), code.GetName(), metav1.GetOptions{})
	if err != nil {
		num++
		if num >= 10 {
			err = dr.Delete(context.TODO(), code.GetName(), metav1.DeleteOptions{})
			rls.ServerErroredFlag = true
		}
		goto try
	} else {
		if object.GetAPIVersion() == code.GetAPIVersion() {
			bys, err = json.Marshal(code.Object)
			if err != nil {
				goto Error
			}

			var res v1.CodeServer
			err = json.Unmarshal(bys, &res)
			if err != nil {
				goto Error
			}
			if res.ObjectMeta.Name != object.GetName() {
				goto Error
			}

			if len(res.Status.Conditions) > 0 {
				for _, condition := range res.Status.Conditions {
					switch condition.Type {
					case v1.ServerCreated: //means the code server has been accepted by the system.
						if condition.Status == corev1.ConditionTrue {
							rls.ServerCreatedFlag = true
						}
						rls.ServerCreatedTime = condition.LastTransitionTime.String()
					case v1.ServerReady: //means the code server has been ready for usage.
						if condition.Status == corev1.ConditionTrue {
							rls.ServerReadyFlag = true
						}
						rls.ServerReadyTime = condition.LastTransitionTime.String()
						rls.InstanceEndpoint = condition.Message["instanceEndpoint"]
					case v1.ServerBound: //means the code server has been bound to user.
						if condition.Status == corev1.ConditionTrue {
							rls.ServerBoundFlag = true
						}
						rls.ServerBoundTime = condition.LastTransitionTime.String()
					case v1.ServerRecycled: //means the code server has been recycled totally.
						if condition.Status == corev1.ConditionTrue {
							rls.ServerRecycledFlag = true
						}
						rls.ServerRecycledTime = condition.LastTransitionTime.String()
					case v1.ServerInactive: //means the code server will be marked inactive if `InactiveAfterSeconds` elapsed
						if condition.Status == corev1.ConditionTrue {
							rls.ServerInactiveFlag = true
						}
						rls.ServerInactiveTime = condition.LastTransitionTime.String()
					case v1.ServerErrored: //means failed to reconcile code server.
						if condition.Status == corev1.ConditionTrue {
							rls.ServerErroredFlag = true
						}
						rls.ServerErroredTime = condition.LastTransitionTime.String()
					}
				}
			}
			goto True
		}
	}
Error:
	rls.ServerErroredFlag = true
	return rls
True:
	return rls
}
