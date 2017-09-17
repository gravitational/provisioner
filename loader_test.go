package provisioner

import (
	"github.com/gravitational/provisioner/provider/awsutil"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"os"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func TestMain(t *testing.T) { TestingT(t) }

type WithVPCSuite struct{}
type WithoutVPCSuite struct{}
type LoaderSuite struct{}

var _ = Suite(&WithVPCSuite{})

// Test generating terraform script with an existing VPC as input
func (s *WithVPCSuite) TestLoad(c *C) {
	cfg := LoaderConfig{
		TemplatePath:    "./fixture/vars.tf.template",
		ClusterTemplate: "./fixture/cluster.tf.template",
		VPCID:           "vpc-1",
	}

	loader, err := NewLoader(cfg)
	loader.EC2 = &awsutil.MockEC2Service{}

	c.Assert(err, IsNil)
	loader.EC2 = &awsutil.MockEC2Service{}

	data, err := loader.load()
	c.Assert(err, IsNil)

	stubTemplate := loadStubTemplate(c, "./fixture/output/final_terraform_with_vpc.tf")
	// Assert with string so that it's easier to read if test failed
	c.Assert(string(data), DeepEquals, string(stubTemplate))
}

var _ = Suite(&WithoutVPCSuite{})

// Test generating terraform script without an input VPC
func (s *WithoutVPCSuite) TestLoad(c *C) {
	cfg := LoaderConfig{
		Region:          "us-west-1",
		TemplatePath:    "./fixture/vars.tf.template",
		ClusterTemplate: "./fixture/cluster.tf.template",
	}

	loader, err := NewLoader(cfg)
	loader.EC2 = &awsutil.MockEC2Service{}

	c.Assert(err, IsNil)
	loader.EC2 = &awsutil.MockEC2Service{}

	data, err := loader.templateForNewVPC()
	c.Assert(err, IsNil)

	stubTemplate := loadStubTemplate(c, "./fixture/output/final_terraform_without_vpc.tf")
	c.Assert(string(data), DeepEquals, string(stubTemplate))
}

func loadStubTemplate(c *C, path string) (out []byte) {
	out, err := ioutil.ReadFile(path)
	c.Assert(err, IsNil)
	return out
}

var _ = Suite(&LoaderSuite{})

func (s *LoaderSuite) TestFindPrivateIp(c *C) {
	file, _ := os.Open("./fixture/terraform.show")
	r, e := findInstance("1.2.3.4", file)

	c.Assert(r, Equals, "aws_instance.foo[1]")
	c.Assert(e, IsNil)
}
