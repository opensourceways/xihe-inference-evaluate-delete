package controller

import (
	"container_manager/service"
	"container_manager/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

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
	data, err := service.NewK8sService(i).Get(name)
	tools.Response(c, data, err)
}

func (i *Inference) Create(c *gin.Context) {
	i.initParams(c)
	data, err := service.NewK8sService(i).Create()
	tools.Response(c, data, err)
}

func (i *Inference) ExtendExpiry(c *gin.Context) {
	i.initParams(c)
	expiry := c.PostForm("expiry")
	expiryInt, _ := strconv.Atoi(expiry)
	data, err := service.NewK8sService(i).Update(expiryInt)
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
	return fmt.Sprintf("inference-%s-%s-%s-%s", i.Info.Id, i.Info.ProjectId, i.Info.ProjectName, i.Info.LastCommit)
}
