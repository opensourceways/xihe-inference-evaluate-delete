package route

import (
	"container_manager/controller"
	"github.com/gin-gonic/gin"
)

func Route(r *gin.Engine) {
	gi := r.Group("/inference")
	{
		gi.POST("create", controller.NewInferControl().Create)
		gi.POST("extend_expiry", controller.NewInferControl().ExtendExpiry)
		gi.GET("get", controller.NewInferControl().Create)
	}

}
