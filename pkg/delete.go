package pkg

import (
	"k8s.io/client-go/kubernetes"
)

func Delete(kubeClient *kubernetes.Clientset) error {
	dss := kubeClient.AppsV1().DaemonSets("default")
	return dss.Delete("towel", nil)
}
