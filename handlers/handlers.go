package handlers

import (
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"net/http"
	"pod-creator-demo/model"
)

func CreateDeployInstanceHandler(client clientset.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.DeployCreateRequest
		klog.Errorf("Handle deploy create request")
		if err := c.ShouldBindJSON(&req); err != nil {
			klog.Errorf("Error binding request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := req.Validae(); err != nil {
			klog.Errorf("Validae Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if the Deployment already exist
		exist, err := client.AppsV1().Deployments(req.Namespace).Get(c.Request.Context(), req.Name, metav1.GetOptions{})
		if exist != nil {
			klog.Errorf("Deployment %s already exists", req.Name)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create Deployment
		deployment, err := GenerateDeploymentTemplate(req)
		if err != nil {
			klog.Errorf("GenerateDeploymentTemplate Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		klog.Infof("Creating deployment %s in namespace %s", deployment.Name, deployment.Namespace)
		_, err = client.AppsV1().Deployments(deployment.Namespace).Create(c.Request.Context(), deployment, metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("Create Deployment Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create Service
		service, err := GenerateServiceTemplate(req)
		if err != nil {
			klog.Errorf("GenerateServiceTemplate Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		klog.Infof("Creating service %s in namespace %s", service.Name, service.Namespace)
		_, err = client.CoreV1().Services(service.Namespace).Create(c.Request.Context(), service, metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("Create Service Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":    "Deployment and Service Created Successfully",
			"deployment": deployment,
			"service":    service,
		})
	}
}
