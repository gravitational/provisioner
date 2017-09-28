package awsutil

import (
	. "gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func TestSubnet(t *testing.T) { TestingT(t) }

type SubnetSuite struct{}

var _ = Suite(&SubnetSuite{})

func (s *SubnetSuite) TestSelectVPCSubnet(c *C) {
	subnet, err := SelectVPCSubnet("10.0.0.0/16", []string{})
	c.Assert(err, IsNil)
	c.Assert(subnet, Equals, "10.0.0.0/24")

	subnet, err = SelectVPCSubnet("10.0.0.0/16", []string{"10.0.0.0/24", "10.0.2.0/24"})
	c.Assert(err, IsNil)
	c.Assert(subnet, Equals, "10.0.1.0/24")
}
