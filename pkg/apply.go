package pkg

import (
	"log"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

var ds = func() *appsv1.DaemonSet {
	const dss = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: towel
  namespace: default
spec:
  selector:
    matchLabels:
      app: towel
  template:
    metadata:
      labels:
        app: towel
    spec:
      hostIPC: true
      hostPID: true
      hostNetwork: true
      containers:
      - name: sleep
        image: jasonkeene/towel
        command:
        - /bin/bash
        - -c
        - "while true; do sleep 3600; done"
        env:
        - name: BPFTRACE_KERNEL_SOURCE
          value: /kernel/kernel
        - name: BCC_KERNEL_SOURCE
          value: /kernel/kernel
        securityContext:
          privileged: true
        volumeMounts:
        - name: sys
          mountPath: /sys
        - name: libmodules
          mountPath: /lib/modules
        - name: varlibdocker
          mountPath: /var/lib/docker
        - name: varrun
          mountPath: /var/run
      volumes:
      - name: sys
        hostPath:
          path: /sys
      - name: libmodules
        hostPath:
          path: /lib/modules
      - name: varlibdocker
        hostPath:
          path: /var/lib/docker
      - name: varrun
        hostPath:
          path: /var/run
`
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(dss), nil, nil)
	if err != nil {
		log.Fatalf("error decoding daemonset: %s", err)
	}

	return obj.(*appsv1.DaemonSet)
}()

func Apply(kubeClient *kubernetes.Clientset) error {
	dss := kubeClient.AppsV1().DaemonSets("default")
	_, err := dss.Get("towel", metav1.GetOptions{})
	if err == nil {
		return nil
	}

	_, err = dss.Create(ds)
	return err
}
