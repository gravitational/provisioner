package awsutil

import (
	. "gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func TestVPC(t *testing.T) { TestingT(t) }

type VPCSuite struct{}

var _ = Suite(&VPCSuite{})

func (s *VPCSuite) TestNewVPC(c *C) {
	vpc, err := NewVPC(&MockEC2Service{}, "us-west-2")
	c.Assert(err, IsNil)

	c.Assert(vpc.CIDR, Equals, defaultCIDR)
	c.Assert(vpc.Region, Equals, "us-west-2")
}

func (s *VPCSuite) TestFindAZ(c *C) {
	vpc, _ := NewVPC(&MockEC2Service{}, "us-west-2")

	c.Assert(vpc.AvailabilityZones, DeepEquals, testZones, Commentf("AvailabilityZones is incorrect"))
}

func (s *VPCSuite) TestGenSubnets(c *C) {
	vpc, _ := NewVPC(&MockEC2Service{}, "us-west-2")

	privateSubnets := []string{"10.1.0.0/24", "10.1.2.0/24", "10.1.4.0/24"}
	c.Assert(vpc.Subnets, DeepEquals, privateSubnets, Commentf("Private subnet is incorrect"))

	publicSubnets := []string{"10.1.1.0/24", "10.1.3.0/24", "10.1.5.0/24"}
	c.Assert(vpc.PublicSubnets, DeepEquals, publicSubnets, Commentf("Public subnet is incorrect"))
}

func (s *VPCSuite) TestGenVars(c *C) {
	vpc, _ := NewVPC(&MockEC2Service{}, "us-west-2")

	genVariables := vpc.GenVars()
	expected := map[string]interface{}{
		"variables": map[string]interface{}{
			"aws": map[string]interface{}{
				"region":         "us-west-2",
				"vpc_cidr":       vpc.CIDR,
				"azs":            vpc.AvailabilityZones,
				"subnets":        vpc.Subnets,
				"public_subnets": vpc.PublicSubnets,
			},
		},
	}

	c.Assert(genVariables, DeepEquals, expected, Commentf("Generated variables is in correct"))
}
