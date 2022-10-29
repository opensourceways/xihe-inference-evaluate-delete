package controller

import (
	"container_manager/service"
	"container_manager/tools"
	"fmt"
	"github.com/gin-gonic/gin"
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
	if err := c.ShouldBindJSON(i.Info); err != nil {
		tools.Failure(c, err)
	}
	data, err := service.NewK8sService().Create(i)
	tools.Response(c, data, err)
}

func (i *Inference) ExtendExpiry(c *gin.Context) {
	if err := c.ShouldBindJSON(i.Info); err != nil {
		tools.Failure(c, err)
	}

	data, err := service.NewK8sService().Update(i, i.Info.Expiry)
	tools.Response(c, data, err)
}

func (i *Inference) GeneMetaName() string {
	return fmt.Sprintf("%s-%s", MetaNameInference, i.Info.LastCommit)
}

func (i *Inference) GeneLabels() map[string]string {
	m := make(map[string]string)
	m["id"] = i.Info.Id
	m["project_id"] = i.Info.ProjectId
	m["last_commit"] = i.Info.LastCommit
	m["project_name"] = i.Info.ProjectName
	m["project_owner"] = i.Info.ProjectOwner
	m["type"] = MetaNameInference
	return m
}
