package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type flags struct {
	fs *pflag.FlagSet

	podName       string
	labelSelector string
	fieldSelector string
	namespace     string
}

func parseFlags(args []string) (*flags, error) {
	f := &flags{
		fs: pflag.NewFlagSet("exec", pflag.ContinueOnError),
	}
	f.fs.StringVarP(&f.labelSelector, "selector", "l", "", "")
	f.fs.StringVar(&f.fieldSelector, "field-selector", "", "")
	f.fs.StringVarP(&f.namespace, "namespace", "n", "default", "")
	err := f.fs.Parse(args)
	if err != nil {
		return nil, err
	}
	f.podName = f.fs.Arg(0)

	return f, nil
}

func podInfo(kubeClient *kubernetes.Clientset, f *flags) (string, string, error) {
	var pod *corev1.Pod
	if f.podName != "" {
		var err error
		pod, err = kubeClient.CoreV1().Pods(f.namespace).Get(f.podName, metav1.GetOptions{})
		if err != nil {
			return "", "", err
		}
	} else {
		pl, err := kubeClient.CoreV1().Pods(f.namespace).List(metav1.ListOptions{
			LabelSelector: f.labelSelector,
			FieldSelector: f.fieldSelector,
		})
		if err != nil {
			return "", "", err
		}
		if len(pl.Items) < 1 {
			return "", "", fmt.Errorf("there are no pods with that search criteria: %#v", f)
		}
		if len(pl.Items) > 1 {
			return "", "", fmt.Errorf("there are too many pods with that search criteria: %#v", f)
		}
		pod = &pl.Items[0]
	}
	nodeName := pod.Spec.NodeName
	containerID := strings.TrimPrefix(
		pod.Status.ContainerStatuses[0].ContainerID,
		"docker://",
	)
	return nodeName, containerID, nil
}

func Exec(kubeClient *kubernetes.Clientset, args ...string) error {
	f, err := parseFlags(args)
	if err != nil {
		return err
	}

	nodeName, containerID, err := podInfo(kubeClient, f)
	if err != nil {
		return err
	}

	pl, err := kubeClient.CoreV1().Pods("default").List(metav1.ListOptions{
		LabelSelector: "app=towel",
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		return err
	}
	if len(pl.Items) < 1 {
		return fmt.Errorf("there are no pods with that search criteria: %#v", f)
	}
	if len(pl.Items) > 1 {
		return fmt.Errorf("there are too many pods with that search criteria: %#v", f)
	}
	towelPodName := pl.Items[0].Name

	cmd := exec.Command(
		"kubectl",
		"exec",
		"-it",
		"--namespace", "default",
		towelPodName,
		"/bin/start",
		containerID,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
