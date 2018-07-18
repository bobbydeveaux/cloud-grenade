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

	defaultSgroups := []*string{}
	failedSgroups := []*string{}

	for _, sgroup := range resp.SecurityGroups {
		if *sgroup.GroupName == "default" {
			//nukeSecurityGroupIngressRules(*sgroup.GroupId)
			defaultSgroups = append(defaultSgroups, sgroup.GroupId)
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
			failedSgroups = append(failedSgroups, sgroup.GroupName)
		}
		fmt.Println(result)
	}

	if len(failedSgroups) == 0 {
		return
	}

	fmt.Println("Due to the above failures - now iterating through to remove dependencies, and cleaning up.")

	for _, dsgroup := range defaultSgroups {
		for _, fsgroup := range failedSgroups {
			nukeSecurityGroupIngressRules(*dsgroup, *fsgroup)
		}
	}

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
