package main

import (
	"github.com/gravitational/provisioner"
	. "gopkg.in/check.v1"
	"os"
	"testing"
)

// mockRunner implements a Runner and stores the parameters it was invoked with
type mockRunner struct {
	Args []string
}

// Run is our main mock function
// We simply store state of argument into the struct itself
// so we can assert if it's called or not
func (m *mockRunner) Run(args []string) error {
	m.Args = args
	return nil
}

// Hook up gocheck into the "go test" runner.
func TestMain(t *testing.T) { TestingT(t) }

type MainSuite struct{}

var _ = Suite(&MainSuite{})

// Test parse correct parameters
func (s *MainSuite) TestParseCliArgument(c *C) {
	m := &mockRunner{}
	getCommand = func() provisioner.CommandRunner {
		return m
	}

	os.Args = []string{"first", "second", "third"}
	main()

	c.Assert(m.Args, DeepEquals, []string{"second", "third"})
}

// Test building command struct
func (s *MainSuite) TestCreateCommand(c *C) {
	command := getCommand().(*provisioner.Command)
	c.Assert(command.App.Name, Equals, "provisioner", Commentf("Command name is incorrect"))
	c.Assert(command.App.Help, Equals, "Terraform based provisioners for ops center", Commentf("Command help is incorrect"))
}
