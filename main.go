package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/porter-dev/porter-node-drainer/pkg/k8s"

	porteraws "github.com/porter-dev/porter-node-drainer/pkg/aws"
)

func HandleRequest(ctx context.Context, event events.AutoScalingEvent) error {
	fmt.Println("starting node drain for autoscaling group:", fmt.Sprintf("%v", event.Detail["AutoScalingGroupName"]))

	fmt.Println("getting ec2 instance information from AWS")

	hostname, err := porteraws.GetEC2InstanceHostname(event)

	if err != nil {
		fmt.Println("failed to get credentials:", err)
		return err
	}

	fmt.Println("got ec2 instance hostname:", hostname)

	fmt.Println("getting credentials from environment")

	eksCreds, err := porteraws.NewEKSCredentialsFromEnv()

	if err != nil {
		fmt.Println("failed to get credentials:", err)
		return err
	}

	// get Kubernetes clientset from eks credentials
	clientset, err := eksCreds.GetClientSet()

	if err != nil {
		fmt.Println("could not get kubernetes clientset:", err)
		return err
	}

	err = k8s.DrainNode(clientset, hostname)

	if err != nil {
		fmt.Println("could not drain node:", err)
		return err
	}

	svc := autoscaling.New(session.New())

	asgInput := &autoscaling.CompleteLifecycleActionInput{
		AutoScalingGroupName:  aws.String(fmt.Sprintf("%v", event.Detail["AutoScalingGroupName"])),
		LifecycleActionResult: aws.String("CONTINUE"),
		LifecycleActionToken:  aws.String(fmt.Sprintf("%v", event.Detail["LifecycleActionToken"])),
		LifecycleHookName:     aws.String(fmt.Sprintf("%v", event.Detail["LifecycleHookName"])),
	}

	_, err = svc.CompleteLifecycleAction(asgInput)

	if err == nil {
		fmt.Println("Successfully drained node!")
	}

	return err
}

func main() {
	lambda.Start(HandleRequest)
}
