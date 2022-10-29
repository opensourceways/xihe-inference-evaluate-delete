package service

import (
	"container_manager/client"
	"context"
	"github.com/qinsheng99/crdcode/api/v1"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/json"
)

type MetaNameInter interface {
	GeneMetaName() string
}

type LabelsInter interface {
	GeneLabels() map[string]string
}

type CreateInter interface {
	MetaNameInter
	LabelsInter
}

type UpdateInter interface {
	MetaNameInter
}

type K8sService struct {
}

func NewK8sService() *K8sService {
	return new(K8sService)
}

func (s *K8sService) Get(name string) (interface{}, error) {
	cli := client.GetDyna()
	resource, err, _ := s.GetResource()
	if err != nil {
		return nil, err
	}

	get, err := cli.Resource(resource).Namespace("default").Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	marshal, err := json.Marshal(get.Object)
	if err != nil {
		return nil, err
	}

	var res v1.CodeServer
	err = json.Unmarshal(marshal, &res)
	if err != nil {
		return nil, err

	}
	return res, nil
}

func (s *K8sService) Create(cd CreateInter) (interface{}, error) {
	cli := client.GetDyna()
	resource, err, res := s.GetResource()
	if err != nil {
		return nil, err
	}

	res.Object["metadata"] = map[string]interface{}{
		"name":   cd.GeneMetaName(),
		"labels": cd.GeneLabels(),
	}

	dr := cli.Resource(resource).Namespace("default")
	create, err := dr.Create(context.TODO(), res, metav1.CreateOptions{})
	return create, err
}

func (s *K8sService) Update(ud UpdateInter, expiry int) (interface{}, error) {
	cli := client.GetDyna()
	resource, err, _ := s.GetResource()
	if err != nil {
		return nil, err
	}

	get, err := cli.Resource(resource).Namespace("default").Get(context.TODO(), ud.GeneMetaName(), metav1.GetOptions{})
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
	return nil, err
}

func (s *K8sService) GetResource() (schema.GroupVersionResource, error, *unstructured.Unstructured) {
	k, err, res := s.resource()
	if err != nil {
		return schema.GroupVersionResource{}, err, nil
	}

	mapping, err := client.GetrestMapper().RESTMapping(k.GroupKind(), k.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err, nil
	}

	return mapping.Resource, nil, res
}

func (s *K8sService) resource() (kind *schema.GroupVersionKind, err error, _ *unstructured.Unstructured) {
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
