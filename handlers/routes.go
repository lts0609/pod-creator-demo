package handlers

import (
	"github.com/gin-gonic/gin"
	clientset "k8s.io/client-go/kubernetes"
)

func SetupRoutes(r *gin.Engine, client clientset.Interface) {
	r.POST("/gpu-container/instance", CreateRequestHandler(client))
}
