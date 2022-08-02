package aws

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetEC2InstanceHostname(event events.AutoScalingEvent) (string, error) {
	ec2Svc := ec2.New(session.New())

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(fmt.Sprintf("%v", event.Detail["EC2InstanceId"])),
		},
	}

	result, err := ec2Svc.DescribeInstances(input)
	if err != nil {
		fmt.Println("could not describe ec2 instances:", err)
		return "", err
	}

	hostname := *result.Reservations[0].Instances[0].PrivateDnsName

	return hostname, nil
}
