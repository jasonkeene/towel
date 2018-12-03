package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/jasonkeene/towel/pkg"
)

var errBadCommand = errors.New("please specify a subcommand [apply, exec, delete]")

func command() (string, error) {
	if len(os.Args) < 2 {
		return "", errBadCommand
	}
	return os.Args[1], nil
}

func execArgs() []string {
	return os.Args[2:]
}

func dispatch(cmd string) error {
	kubeConfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return fmt.Errorf("error building api config: %s", err)
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes client: %s", err)
	}

	switch cmd {
	case "apply":
		return pkg.Apply(kubeClient)
	case "exec":
		return pkg.Exec(kubeClient, execArgs()...)
	case "delete":
		return pkg.Delete(kubeClient)
	default:
		return errBadCommand
	}
}

func fatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	cmd, err := command()
	fatalErr(err)
	err = dispatch(cmd)
	fatalErr(err)
}
