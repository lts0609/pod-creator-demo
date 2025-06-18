package handlers

import (
	"github.com/gin-gonic/gin"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func SetupRoutes(r *gin.Engine, client clientset.Interface, config *rest.Config) {
	r.POST("/gpu-container/instance", CreateRequestHandler(client))
	// 注册 WebSocket 相关路由
	r.GET("/terminal", TerminalSessionHandler(client, config))
}
