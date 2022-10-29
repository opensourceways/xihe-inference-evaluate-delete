package controller

import (
	"container_manager/service"
	"container_manager/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

const metaNameInference = "inference"

type Inference struct {
	Info *InferenceInfo
}

type InferenceInfo struct {
	Id           string
	ProjectId    string
	LastCommit   string
	ProjectName  string
	ProjectOwner string
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
	i.initParams(c)
	data, err := service.NewK8sService().Create(i)
	tools.Response(c, data, err)
}

func (i *Inference) ExtendExpiry(c *gin.Context) {
	i.initParams(c)
	expiry := c.PostForm("expiry")
	expiryInt, _ := strconv.Atoi(expiry)
	data, err := service.NewK8sService().Update(i, expiryInt)
	tools.Response(c, data, err)
}

func (i *Inference) initParams(c *gin.Context) {

	// todo 入参方式待定
	i.Info.Id = c.PostForm("id")
	i.Info.ProjectId = c.PostForm("project_id")
	i.Info.LastCommit = c.PostForm("last_commit")
	i.Info.ProjectName = c.PostForm("project_name")
	i.Info.ProjectOwner = c.PostForm("project_owner")
}

func (i *Inference) GeneMetaName() string {
	return fmt.Sprintf("%s-%s", metaNameInference, i.Info.LastCommit)
}

func (i *Inference) GeneLabels() map[string]string {
	m := make(map[string]string)
	m["id"] = i.Info.Id
	m["project_id"] = i.Info.ProjectId
	m["last_commit"] = i.Info.LastCommit
	m["project_name"] = i.Info.ProjectName
	m["project_owner"] = i.Info.ProjectOwner
	m["type"] = metaNameInference
	return m
}
