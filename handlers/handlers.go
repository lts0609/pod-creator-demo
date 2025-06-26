package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"net/http"
	"pod-creator-demo/model"
	"strconv"
)

func CreateRequestHandler(client clientset.Interface) gin.HandlerFunc {
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
		deploymentTemplate, err := GenerateDeploymentTemplate(req)
		if err != nil {
			HandleError(c, "GenerateDeploymentTemplate Error", err, http.StatusBadRequest)
			return
		}
		deployment, err := client.AppsV1().Deployments(deploymentTemplate.Namespace).Create(c.Request.Context(), deploymentTemplate, metav1.CreateOptions{})
		if err != nil {
			HandleError(c, "Create Deployment Error", err, http.StatusBadRequest)
			return
		}
		klog.Infof("Create Deployment %s in Namespace %s Successfully", deployment.Name, deployment.Namespace)

		// Create Secret
		secretTemplate, err := GenerateSecretTemplate(req, deployment)
		if err != nil {
			HandleError(c, "GenerateSecretTemplate Error", err, http.StatusBadRequest)
			return
		}
		secret, err := client.CoreV1().Secrets(secretTemplate.Namespace).Create(c.Request.Context(), secretTemplate, metav1.CreateOptions{})
		if err != nil {
			HandleError(c, "Create Secret Error", err, http.StatusBadRequest)
			return
		}
		klog.Infof("Create Secret %s in Namespace %s Successfully", secret.Name, secret.Namespace)

		// Create Service
		serviceTemplate, err := GenerateServiceTemplate(req, deployment)
		if err != nil {
			HandleError(c, "GenerateServiceTemplate Error", err, http.StatusBadRequest)
			return
		}
		service, err := client.CoreV1().Services(serviceTemplate.Namespace).Create(c.Request.Context(), serviceTemplate, metav1.CreateOptions{})
		if err != nil {
			HandleError(c, "Create Service Error", err, http.StatusBadRequest)
			return
		}
		klog.Infof("Create Service %s in Namespace %s Successfully", service.Name, service.Namespace)

		// TODO: Use Informer watch pods active, and patch Ingress with pod's env($NB_PREFIX)
		selector := metav1.FormatLabelSelector(deployment.Spec.Selector)
		pod, err := client.CoreV1().Pods(req.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		var jupyterPath string
		jupyterPath = ("notebook/" + pod.Items[0].Name)
		fmt.Println("jupyterPath", jupyterPath)

		var sshPort, jupyterPort string
		ports := service.Spec.Ports
		for _, port := range ports {
			if port.Name == "ssh" {
				sshPort = strconv.Itoa(int(port.NodePort))
			}
			if port.Name == "jupyter" {
				jupyterPort = strconv.Itoa(int(port.NodePort))
			}
		}

		var ssh_domain, jupyter_domain string
		ssh_domain = DeafultDomain
		jupyter_domain = "http://" + DeafultDomain + ":" + jupyterPort + "/" + jupyterPath
		c.JSON(http.StatusCreated, gin.H{
			"Message":       "Deployment and Service Created Successfully",
			"Deployment":    deployment.Name,
			"SSHDomain":     ssh_domain,
			"SSHPort":       sshPort,
			"SSHUser":       "jovyan",
			"JupyterDomain": jupyter_domain,
			"InitPassword":  secret.Data["SSH_PASSWORD"],
		})
	}
}

func HandleError(c *gin.Context, resaon string, err error, code int) {
	errMsg := fmt.Sprintf("Reason: %v, Error: %v", resaon, err)
	klog.Errorf(errMsg)
	c.JSON(code, gin.H{"error": errMsg})
}
