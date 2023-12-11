package k8s

import (
	"fmt"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func CreateClientFromClusterConfig() (*k8sclient.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Errorf("failed to create the Kubernetes config: %v", err)
		return nil, err
	}
	clientset, err := k8sclient.NewForConfig(config)
	if err != nil {
		fmt.Errorf("failed to create the dce client: %v", err)
		return nil, err
	}
	return clientset, nil
}
