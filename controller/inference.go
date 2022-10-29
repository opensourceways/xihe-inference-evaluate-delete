package controller

import (
	"container_manager/service"
	"container_manager/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
)

const MetaNameInference = "inference"

type Inference struct {
	Info *InferenceInfo
}

type InferenceInfo struct {
	Id           string `json:"id"`
	ProjectId    string `json:"project_id"`
	LastCommit   string `json:"last_commit"`
	ProjectName  string `json:"project_name"`
	ProjectOwner string `json:"project_owner"`
	Expiry       int    `json:"expiry,omitempty"`
}

func NewInferControl() *Inference {
	return &Inference{
		Info: new(InferenceInfo),
	}
}

func (i *Inference) Get(c *gin.Context) {
	name := c.Query("name")
	data, err := service.NewK8sService().Get(name)
	tools.Response(c, data, err)
}

func (i *Inference) Create(c *gin.Context) {
	var t InferenceInfo
	if err := c.ShouldBindJSON(&t); err != nil {
		tools.Failure(c, err)
	}
	i.Info = &t
	data, err := service.NewK8sService().Create(i)
	tools.Response(c, data, err)
}

func (i *Inference) ExtendExpiry(c *gin.Context) {
	var t InferenceInfo
	if err := c.ShouldBindJSON(&t); err != nil {
		tools.Failure(c, err)
	}
	i.Info = &t

	data, err := service.NewK8sService().Update(i, i.Info.Expiry)
	tools.Response(c, data, err)
}

func (i *Inference) GeneMetaName() string {
	return fmt.Sprintf("%s-%s", MetaNameInference, i.Info.LastCommit)
}

func (i *Inference) GeneLabels() map[string]string {
	m := make(map[string]string)
	b, _ := json.Marshal(i.Info)
	_ = json.Unmarshal(b, &m)
	m["type"] = MetaNameInference
	log.Println(m)
	return m
}
