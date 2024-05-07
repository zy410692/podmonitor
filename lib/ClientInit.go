package lib

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var K8sClient *kubernetes.Clientset

func init() {
	config := &rest.Config{
		Host:        "http://x.x.x.x:9090",
		BearerToken: "kubeconfig-user-psg66.c-24jxf:jgq5k46rkjjmt6n8d9kzwhgzjcxfrdgqqtmcbvtb8xrj4whtb5fxxh",
	}
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	K8sClient = c
}
