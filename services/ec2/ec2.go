package ec2

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
)

func Nuke() {
	fmt.Println("Nuking all EC2 instances")
	ec2svc := ec2.New(session.New())
	params := &ec2.DescribeInstancesInput{}
	resp, err := ec2svc.DescribeInstances(params)
	if err != nil {
		fmt.Println("there was an error listing instances in", err.Error())
		log.Fatal(err.Error())
	}

	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String("i-1234567890abcdef0"),
		},
	}
	for idx, res := range resp.Reservations {
		fmt.Println("  > Reservation Id", *res.ReservationId, " Num Instances: ", len(res.Instances))
		for _, inst := range resp.Reservations[idx].Instances {
			fmt.Println("    - Instance ID: ", *inst.InstanceId)
			input = &ec2.TerminateInstancesInput{
				InstanceIds: []*string{
					aws.String("i-1234567890abcdef0"),
				},
			}

		}
	}

	result, err := ec2svc.TerminateInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return
	}
	fmt.Println(result)
}
