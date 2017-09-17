package provisioner

import (
	"gopkg.in/alecthomas/kingpin.v2"
	. "gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func TestCommand(t *testing.T) { TestingT(t) }

type CommandSuite struct{}

func (s *CommandSuite) TestLoadCommand(c *C) {
	var cfg LoaderConfig
	app := kingpin.New("provisioner", "Terraform based provisioners for ops center")

	command := LoadCommands(app, &cfg)
	c.Assert(command.initVars, NotNil)
	c.Assert(command.findInstance, NotNil)
	c.Assert(command.syncFiles, NotNil)
	c.Assert(command.removeS3Key, NotNil)
}
