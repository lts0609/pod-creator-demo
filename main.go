package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"os/signal"
	"pod-creator-demo/clientbuilder"
	"pod-creator-demo/handlers"
	"syscall"
	"time"
)

func main() {
	klog.InitFlags(nil)
	klog.Errorf("Pod Creator Starting")

	clientBuilder, err := clientbuilder.NewClientBuilder()
	if err != nil {
		klog.Errorf("Failed to create client builder: %v", err)
	}
	client, err := clientBuilder.Client()
	if err != nil {
		klog.Errorf("Failed to create client: %v", err)
	}
	config, err := clientBuilder.Config()
	if err != nil {
		klog.Errorf("Failed to get rest config: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://8.156.65.148:9528"}, // 允许的前端源
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // 预检请求缓存时间
	}))
	handlers.SetupRoutes(router, client, config)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	err = srv.ListenAndServe()
	if err != nil {
		klog.Errorf("Failed to start server: %v", err)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		klog.Errorf("Server Shutdown Failed: %v", err)
	}
	klog.Infof("Shutting down gracefully")
}
