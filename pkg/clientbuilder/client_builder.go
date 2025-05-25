package clientbuilder

import (
	"k8s.io/client-go/discovery"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type ClientBuilder interface {
	Config() *restclient.Config
	Client() (clientset.Interface, error)
}

type ClientBuilderImpl struct {
	ClientConfig *restclient.Config
}

// Get the rest config
func (c ClientBuilderImpl) Config() (*restclient.Config, error) {
	config := c.ClientConfig
	return config.AddUserAgent(&config, name), nil
}

// Get the root client
func (c ClientBuilderImpl) Client() (clientset.Interface, error) {
	config, err := c.Config()
	if err != nil {
		return nil, err
	}
	return clientset.NewForConfig(config)
}
