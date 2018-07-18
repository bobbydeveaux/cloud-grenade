package main

import (
	"github.com/bobbydeveaux/cloud-grenade/services/ec2"
	"github.com/bobbydeveaux/cloud-grenade/services/vpc"
)

func main() {
	ec2.Nuke()
	vpc.Nuke()
}
