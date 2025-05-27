package handlers

import (
	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"net/http"
	"pod-creator-demo/model"
)

func SetupRoutes(r *gin.Engine, client clientset.Interface) {
	r.POST("/create-pod", CreatePodHandler(client))
}

func CreatePodHandler(client clientset.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.PodCreateRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			klog.Fatalf("Error binding request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := req.Validae(); err != nil {
			klog.Fatalf("Validae Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		pod, err := GeneratePodTemplate(req)
		if err != nil {
			klog.Fatalf("Error generating pod template: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = client.CoreV1().Pods(pod.Namespace).Create(c.Request.Context(), pod, metav1.CreateOptions{})
		if err != nil {
			klog.Fatalf("Error creating pod: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Pod created successfully",
			"pod":     pod})
	}
}

func GeneratePodTemplate(req model.PodCreateRequest) (*v1.Pod, error) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: req.Name,
			Namespace:    req.Namespace,
			Labels: map[string]string{
				"app":       "myapp",
				"create-by": "pod-creator",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  req.Name,
					Image: req.Image,
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU: resource.MustParse("1"),
						},
						Limits: v1.ResourceList{
							v1.ResourceCPU: resource.MustParse("1"),
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyAlways,
		},
	}
	return pod, nil
}
