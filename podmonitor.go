package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"podmonitor/lib"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typev1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/util/retry"
)

type args struct {
	namespace  string
	image      string
	deployment string
}

const (
	numberOfPoll = 200
	pollInterval = 3
)

func parseArgs() *args {
	namespace := flag.String("n", "", "namespace")
	deployment := flag.String("deploy", "", "deployment name")
	image := flag.String("image", "", "image for update")
	flag.Parse()
	var _args args
	if *namespace == "" {
		fmt.Fprintln(os.Stderr, "namespace must be specified")
		os.Exit(1)
	}
	_args.namespace = *namespace
	if *deployment == "" {
		fmt.Fprintln(os.Stderr, "deployment must be specified")
		os.Exit(1)
	}
	_args.deployment = *deployment
	if *image == "" {
		fmt.Fprintln(os.Stderr, "image must be specified")
		os.Exit(1)
	}
	_args.image = *image
	return &_args
}

func main() {
	_args := parseArgs()
	// creates the in-cluster config

	deploymentsClient := lib.K8sClient.AppsV1().Deployments(_args.namespace)
	ctx := context.Background()

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := deploymentsClient.Get(ctx, _args.deployment, metav1.GetOptions{})
		if getErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to get latest version of Deployment %s: %v", _args.deployment, getErr)
			os.Exit(1)
		}
		result.Spec.Template.Spec.Containers[0].Image = _args.image
		_, updateErr := deploymentsClient.Update(ctx, result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to update image version of %s/%s to %s: %v", _args.namespace,
			_args.deployment, _args.image, retryErr)
		os.Exit(1)
	}
	_args.pollDeploy(deploymentsClient)
	fmt.Println("Updated deployment")
}

// watch 太浪费资源了，而且时间太长，还是轮询吧
func (p *args) pollDeploy(deploymentsClient typev1.DeploymentInterface) {
	ctx := context.Background()
	for i := 0; i <= numberOfPoll; i++ {
		time.Sleep(pollInterval * time.Second)
		result, getErr := deploymentsClient.Get(ctx, p.deployment, metav1.GetOptions{})
		if getErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to get latest version of Deployment %s: %v", p.deployment, getErr)
			os.Exit(1)
		}
		resourceStatus := result.Status
		fmt.Printf("%s -> replicas: %d, ReadyReplicas: %d, AvailableReplicas: %d, UpdatedReplicas: %d, UnavailableReplicas: %d\n",
			result.Name,
			resourceStatus.Replicas,
			resourceStatus.ReadyReplicas,
			resourceStatus.AvailableReplicas,
			resourceStatus.UpdatedReplicas,
			resourceStatus.UnavailableReplicas)
		if resourceStatus.Replicas == resourceStatus.ReadyReplicas &&
			resourceStatus.ReadyReplicas == resourceStatus.AvailableReplicas &&
			resourceStatus.AvailableReplicas == resourceStatus.UpdatedReplicas {
			return
		}
	}
}
