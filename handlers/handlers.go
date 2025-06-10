package handlers

import (
	"fmt"
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
			HandleError(c, "Error binding request", err, http.StatusBadRequest)
			return
		}
		if err := req.Validae(); err != nil {
			HandleError(c, "Validae request para,:", err, http.StatusBadRequest)
			return
		}

		// Check if the Deployment already exists
		_, err := client.AppsV1().Deployments(req.Namespace).Get(c.Request.Context(), req.Name, metav1.GetOptions{})
		if err == nil {
			err = fmt.Errorf("Deployment %s already exists", req.Name)
			HandleError(c, "Object already exists", err, http.StatusBadRequest)
			return
		}

		// Create Deployment
		deployment, err := GenerateDeploymentTemplate(req)
		if err != nil {
			HandleError(c, "GenerateDeploymentTemplate Error", err, http.StatusBadRequest)
			return
		}
		klog.Infof("Creating Deployment %s in Namespace %s", deployment.Name, deployment.Namespace)
		_, err = client.AppsV1().Deployments(deployment.Namespace).Create(c.Request.Context(), deployment, metav1.CreateOptions{})
		if err != nil {
			HandleError(c, "Create Deployment Error", err, http.StatusBadRequest)
			return
		}
		klog.Infof("Create Deployment %s in Namespace %s Successfully", deployment.Name, deployment.Namespace)

		// Create Service
		service, err := GenerateServiceTemplate(req)
		if err != nil {
			HandleError(c, "GenerateServiceTemplate Error", err, http.StatusBadRequest)
			return
		}
		klog.Infof("Creating Service %s in Namespace %s", service.Name, service.Namespace)
		service, err = client.CoreV1().Services(service.Namespace).Create(c.Request.Context(), service, metav1.CreateOptions{})
		if err != nil {
			HandleError(c, "Create Service Error", err, http.StatusBadRequest)
			return
		}
		klog.Infof("Create Service %s in Namespace %s Successfully", service.Name, service.Namespace)

		c.JSON(http.StatusCreated, gin.H{
			"Message":    "Deployment and Service Created Successfully",
			"Deployment": deployment.Name,
			"Service":    service.Name,
			"NodePort":   service.Spec.Ports[0].NodePort,
		})
	}
}

func TestHandler(client clientset.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.DeployCreateRequest
		klog.Errorf("Handle deploy create request")
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleError(c, "Error binding request", err, http.StatusBadRequest)
			return
		}
		if err := req.Validae(); err != nil {
			HandleError(c, "Validae request para,:", err, http.StatusBadRequest)
			return
		}

		// Create Deployment
		deployment, err := GenerateDeploymentTemplateWithEnv(req)
		if err != nil {
			HandleError(c, "GenerateDeploymentTemplate Error", err, http.StatusBadRequest)
			return
		}
		klog.Infof("Creating Deployment %s in Namespace %s", deployment.Name, deployment.Namespace)
		_, err = client.AppsV1().Deployments(deployment.Namespace).Create(c.Request.Context(), deployment, metav1.CreateOptions{})
		if err != nil {
			HandleError(c, "Create Deployment Error", err, http.StatusBadRequest)
			return
		}
		klog.Infof("Create Deployment %s in Namespace %s Successfully", deployment.Name, deployment.Namespace)

		// Create Service
		service, err := GenerateServiceTemplate(req)
		if err != nil {
			HandleError(c, "GenerateServiceTemplate Error", err, http.StatusBadRequest)
			return
		}
		klog.Infof("Creating Service %s in Namespace %s", service.Name, service.Namespace)
		service, err = client.CoreV1().Services(service.Namespace).Create(c.Request.Context(), service, metav1.CreateOptions{})
		if err != nil {
			HandleError(c, "Create Service Error", err, http.StatusBadRequest)
			return
		}
		klog.Infof("Create Service %s in Namespace %s Successfully", service.Name, service.Namespace)

		c.JSON(http.StatusCreated, gin.H{
			"Message":    "Deployment and Service Created Successfully",
			"Deployment": deployment.Name,
			"Service":    service.Name,
			"NodePort":   service.Spec.Ports[0].NodePort,
		})
	}
}

func HandleError(c *gin.Context, resaon string, err error, code int) {
	errMsg := fmt.Sprintf("Reason: %v, Error: %v", resaon, err)
	klog.Errorf(errMsg)
	c.JSON(code, gin.H{"error": errMsg})
}
