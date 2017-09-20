package awsutil

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gravitational/trace"
)

const (
	// the default CIDR for generated VPC.
	// This can be overridden with TF_VAR_vpc_cidr
	// Please refer to ../scripts/terraform/templates/vars.tf.template
	defaultCIDR = "10.1.0.0/16"
)

// EC2Service defines an interface to EC2
type EC2Service interface {
	DescribeAvailabilityZones(*ec2.DescribeAvailabilityZonesInput) (*ec2.DescribeAvailabilityZonesOutput, error)
	DescribeAvailabilityZonesRequest(*ec2.DescribeAvailabilityZonesInput) (*request.Request, *ec2.DescribeAvailabilityZonesOutput)
	DescribeSubnetsRequest(*ec2.DescribeSubnetsInput) (*request.Request, *ec2.DescribeSubnetsOutput)
	DescribeNatGatewaysRequest(*ec2.DescribeNatGatewaysInput) (*request.Request, *ec2.DescribeNatGatewaysOutput)
	DescribeInternetGatewaysRequest(*ec2.DescribeInternetGatewaysInput) (*request.Request, *ec2.DescribeInternetGatewaysOutput)
	DescribeVpcsRequest(*ec2.DescribeVpcsInput) (*request.Request, *ec2.DescribeVpcsOutput)
}

// VPC struct holds VPC configuration
type VPC struct {
	AvailabilityZones []string
	CIDR              string
	EC2               EC2Service
	PublicSubnets     []string
	Region            string
	Subnets           []string
}

//NewVPC initializes a VPC configuration(cidr, subnet, availability zones)
func NewVPC(svc EC2Service, region string) (*VPC, error) {
	vpc := VPC{
		CIDR:   defaultCIDR,
		EC2:    svc,
		Region: region,
	}

	if err := vpc.findAZ(); err != nil {
		return nil, trace.Wrap(err)
	}

	vpc.genSubnets()

	return &vpc, nil
}

// findAZ finds availability zone from the selected region
func (vpc *VPC) findAZ() error {
	// build availability zones deterministic array
	resultAvailZones, err := vpc.EC2.DescribeAvailabilityZones(nil)
	if err != nil {
		return trace.Wrap(err)
	}

	var azNames []string
	for _, az := range resultAvailZones.AvailabilityZones {
		azNames = append(azNames, *az.ZoneName)
	}
	vpc.AvailabilityZones = azNames

	return nil
}

// genSubnets generates public/private subnet per availabity zones
func (vpc *VPC) genSubnets() error {
	var allocatedSubnets, privateSubnets, publicSubnets []string

	for i := 0; i < len(vpc.AvailabilityZones); i++ {
		if subnet, err := SelectVPCSubnet(vpc.CIDR, allocatedSubnets); err == nil {
			privateSubnets = append(privateSubnets, subnet)
			allocatedSubnets = append(allocatedSubnets, subnet)
		} else {
			return trace.Wrap(err)
		}

		if subnet, err := SelectVPCSubnet(vpc.CIDR, allocatedSubnets); err == nil {
			publicSubnets = append(publicSubnets, subnet)
			allocatedSubnets = append(allocatedSubnets, subnet)
		} else {
			return trace.Wrap(err)
		}
	}

	vpc.Subnets = privateSubnets
	vpc.PublicSubnets = publicSubnets

	return nil
}

// genVars generates necessary vars for terraform
func (vpc *VPC) GenVars() map[string]interface{} {
	return map[string]interface{}{
		"variables": map[string]interface{}{
			"aws": map[string]interface{}{
				"region":         vpc.Region,
				"vpc_cidr":       vpc.CIDR,
				"azs":            vpc.AvailabilityZones,
				"subnets":        vpc.Subnets,
				"public_subnets": vpc.PublicSubnets,
			},
		},
	}
}
