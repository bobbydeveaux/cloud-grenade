package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
)

func Nuke() {

	nukeInstances()
	nukeSecurityGroups()

}

func nukeInstances() {
	fmt.Println("Nuking all EC2 instances")
	ec2svc := ec2.New(session.New())
	params := &ec2.DescribeInstancesInput{}
	resp, err := ec2svc.DescribeInstances(params)
	if err != nil {
		fmt.Println("there was an error listing instances in", err.Error())
		log.Fatal(err.Error())
	}

	instances := []*string{}

	for idx, res := range resp.Reservations {
		fmt.Println("  > Reservation Id", *res.ReservationId, " Num Instances: ", len(res.Instances))
		for _, inst := range resp.Reservations[idx].Instances {
			fmt.Println("    - Instance ID: ", *inst.InstanceId)
			instances = append(instances, inst.InstanceId)
		}
	}

	if len(instances) == 0 {
		return
	}

	input := &ec2.TerminateInstancesInput{
		InstanceIds: instances,
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

func nukeSecurityGroups() {
	fmt.Println("Nuking all EC2 Security Groups")
	ec2svc := ec2.New(session.New())
	params := &ec2.DescribeSecurityGroupsInput{}
	resp, err := ec2svc.DescribeSecurityGroups(params)
	if err != nil {
		fmt.Println("there was an error listing secruity groups in", err.Error())
		log.Fatal(err.Error())
	}

	failedSgroups := make(map[string]string)

	for _, sgroup := range resp.SecurityGroups {
		if *sgroup.GroupName == "default" {
			failedSgroups[*sgroup.GroupId] = *sgroup.GroupName
			continue
		}
		fmt.Println("  > Security Group Id", *sgroup.GroupName, *sgroup.GroupId)

		input := &ec2.DeleteSecurityGroupInput{
			GroupId: sgroup.GroupId,
		}

		result, err := ec2svc.DeleteSecurityGroup(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				fmt.Println(err.Error())
			}
			failedSgroups[*sgroup.GroupId] = *sgroup.GroupName
		}
		fmt.Println(result)
	}

	if len(failedSgroups) == 0 {
		return
	}

	fmt.Println("Due to the above failures - now iterating through to remove dependencies, and cleaning up.")

	for groupId, _ := range failedSgroups {
		for _, groupName := range failedSgroups {
			nukeSecurityGroupIngressRules(groupId, groupName)
		}
	}

	// keep calling this until failedSgroups is no longer reducing (erroring)
	// nukeSecurityGroups()

}

func nukeSecurityGroupIngressRules(securityGroupId string, source string) {
	fmt.Println("Nuking all EC2 Security Group Ingress Rules for ", securityGroupId, " and ", source)
	ec2svc := ec2.New(session.New())

	input := &ec2.RevokeSecurityGroupIngressInput{
		GroupId:                 &securityGroupId,
		SourceSecurityGroupName: &source,
	}

	result, err := ec2svc.RevokeSecurityGroupIngress(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
	fmt.Println(result)

}
