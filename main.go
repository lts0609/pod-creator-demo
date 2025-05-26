package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"os/signal"
	"pod-creator-demo/clientbuilder"
	"pod-creator-demo/handlers"
	"syscall"
	"time"
)

var kubeconfigPath string = clientcmd.RecommendedHomeFile

func main() {
	klog.InitFlags(nil)

	clientbuilder, err := clientbuilder.NewClientBuilder(kubeconfigPath)
	if err != nil {
		klog.Fatalf("Failed to create client builder: %v", err)
	}
	client, err := clientbuilder.Client()
	if err != nil {
		klog.Fatalf("Failed to create client: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	handlers.SetupRoutes(router, client)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	srv.ListenAndServe()

	stopchan := make(chan os.Signal, 1)
	signal.Notify(stopchan, syscall.SIGINT, syscall.SIGTERM)
	<-stopchan
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		klog.Fatalf("Server Shutdown Failed: %v", err)
	}
	klog.Infof("Shutting down gracefully")
}
