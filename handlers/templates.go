package handlers

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"pod-creator-demo/model"
	"pod-creator-demo/utils"
	"strconv"
	"strings"
)

var SSHPort int32 = 22
var JupyterPort int32 = 8888
var TestNodePort int32 = 30000
var GPUContainerLabel = "mfy.com/gpu-container"
var GPUtypeLabel = "mfy.com/gpu-type"

// var InitContainerImage = "m.daocloud.io/docker.io/alpine:3.18"
var InitContainerImage = "containercloud-mirror.xaidc.com/library/alpine:3.20"

func GenerateDeploymentTemplate(req model.DeployCreateRequest) (*appsv1.Deployment, error) {
	// 增加判空
	replicas, err := ParseReplicas(req.Replicas)
	if err != nil {
		return nil, fmt.Errorf("ParseReplicas Error: %v", err)
	}

	labels, err := ParseLabels(req.Labels)
	if err != nil {
		return nil, fmt.Errorf("ParseLabels Error: %v", err)
	}

	podTemplate, err := GeneratePodTemplate(req)
	if err != nil {
		return nil, fmt.Errorf("GeneratePodTemplate Error: %v", err)
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: req.Name + "-",
			Namespace:    req.Namespace,
			Labels:       labels,
			Annotations:  map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": req.Name,
				},
			},
			Template: podTemplate,
		},
	}

	return deployment, nil
}

func GeneratePodTemplate(req model.DeployCreateRequest) (v1.PodTemplateSpec, error) {
	labels, err := ParseLabels(req.Labels)
	if err != nil {
		return v1.PodTemplateSpec{}, fmt.Errorf("ParseLabels Error: %v", err)
	}
	labels["app"] = req.Name

	env := []v1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name:  "NB_PREFIX",
			Value: "/notebook/$(POD_NAME)",
		},
	}

	mainContainer := v1.Container{
		Name:  req.Name,
		Image: req.Image,
		Ports: []v1.ContainerPort{
			{
				Name:          "ssh",
				ContainerPort: SSHPort,
			},
			{
				Name:          "http",
				ContainerPort: JupyterPort,
			},
		},
		Resources: ParseResources(req.Resources),
		Env:       env,
		EnvFrom: []v1.EnvFromSource{
			{
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: req.Name + "-secret",
					},
				},
			},
		},
	}

	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{mainContainer},
			NodeSelector: map[string]string{
				GPUContainerLabel: "true",
				GPUtypeLabel:      req.Resources.GPUType,
			},
		},
	}, nil
}

func GenerateSecretTemplate(req model.DeployCreateRequest, deployment *appsv1.Deployment) (*v1.Secret, error) {
	password, hashedPassword, err := utils.GenerateJupyterPassword()
	if err != nil {
		return nil, fmt.Errorf("GenerateJupyterPassword Error: %v", err)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name + "-secret",
			Namespace: req.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: appsv1.SchemeGroupVersion.String(),
					Kind:       "Deployment",
					Name:       deployment.Name,
					UID:        deployment.UID,
					Controller: new(bool),
				},
			},
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"SSH_PASSWORD":     password,
			"NB_PASSWD":        password,
			"NB_HASHED_PASSWD": hashedPassword,
		},
	}

	return secret, nil
}

func GenerateServiceTemplate(req model.DeployCreateRequest, deployment *appsv1.Deployment) (*v1.Service, error) {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name + "-service",
			Namespace: req.Namespace,
			Labels: map[string]string{
				"app":       req.Name,
				"create-by": "mck",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: appsv1.SchemeGroupVersion.String(),
					Kind:       "Deployment",
					Name:       deployment.Name,
					UID:        deployment.UID,
					Controller: new(bool),
				},
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": req.Name,
			},
			Ports: []v1.ServicePort{
				{
					Name: "ssh",
					Port: SSHPort,
					TargetPort: intstr.IntOrString{
						IntVal: SSHPort,
					},
				},
				{
					Name: "http",
					Port: JupyterPort,
					TargetPort: intstr.IntOrString{
						IntVal: JupyterPort,
					},
					// NodePort: TestNodePort,
				},
			},
		},
	}

	return service, nil
}

func ParseReplicas(replicasStr string) (int32, error) {
	if replicasStr == "" {
		return 1, fmt.Errorf("ReplicasStr is empty")
	}
	replicas, err := strconv.Atoi(replicasStr)
	if err != nil || replicas <= 0 {
		return 1, fmt.Errorf("ReplicasStr is invalid")
	}
	return int32(replicas), nil
}

func ParseResources(res model.Resources) v1.ResourceRequirements {
	requirements := v1.ResourceRequirements{
		Requests: make(v1.ResourceList),
		Limits:   make(v1.ResourceList),
	}

	if res.CPU != "" {
		if cpu, err := resource.ParseQuantity(res.CPU); err == nil {
			requirements.Requests[v1.ResourceCPU] = cpu
			requirements.Limits[v1.ResourceCPU] = cpu
		}
	}

	if res.Memory != "" {
		if mem, err := resource.ParseQuantity(res.Memory); err == nil {
			requirements.Requests[v1.ResourceMemory] = mem
			requirements.Limits[v1.ResourceMemory] = mem
		}
	}

	if res.GPU != "" {
		if gpu, err := resource.ParseQuantity(res.GPU); err == nil {
			requirements.Requests[v1.ResourceName("nvidia.com/gpu")] = gpu
			requirements.Limits[v1.ResourceName("nvidia.com/gpu")] = gpu
		}
	}

	return requirements
}

func ParseLabels(labelSpec string) (map[string]string, error) {
	if len(labelSpec) == 0 {
		return nil, fmt.Errorf("no label spec passed")
	}
	labels := map[string]string{}
	labelSpecs := strings.Split(labelSpec, ",")
	for ix := range labelSpecs {
		labelSpec := strings.Split(labelSpecs[ix], "=")
		if len(labelSpec) != 2 {
			return nil, fmt.Errorf("unexpected label spec: %s", labelSpecs[ix])
		}
		if len(labelSpec[0]) == 0 {
			return nil, fmt.Errorf("unexpected empty label key")
		}
		labels[labelSpec[0]] = labelSpec[1]
	}
	return labels, nil
}
