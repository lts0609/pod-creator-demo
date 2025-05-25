package main

import (
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"pod-creator-demo/pkg/clientbuilder"
)

func main() {
	klog.InitFlags(nil)
	kubeconfigPath := clientcmd.RecommendedHomeFile
	clientbuilder, err := clientbuilder.NewClientBuilder(kubeconfigPath)
	if err != nil {
		klog.Fatalf("Failed to create client builder: %v", err)
	}
	client, err := clientbuilder.Client()
	if err != nil {
		klog.Fatalf("Failed to create client: %v", err)
	}
	router := gin.Default()
	router.POST("/create-pod", func(c *gin.Context) {
	})

}
