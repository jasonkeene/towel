package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type flags struct {
	fs *pflag.FlagSet

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
	return f, f.fs.Parse(args)
}

func Exec(kubeClient *kubernetes.Clientset, args ...string) error {
	f, err := parseFlags(args)
	if err != nil {
		return err
	}

	pl, err := kubeClient.CoreV1().Pods(f.namespace).List(metav1.ListOptions{
		LabelSelector: f.labelSelector,
		FieldSelector: f.fieldSelector,
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
	nodeName := pl.Items[0].Spec.NodeName
	containerID := strings.TrimPrefix(
		pl.Items[0].Status.ContainerStatuses[0].ContainerID,
		"docker://",
	)

	pl, err = kubeClient.CoreV1().Pods("default").List(metav1.ListOptions{
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
