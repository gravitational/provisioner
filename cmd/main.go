// Package main provides main binary to generate Terraform script or inspect
// and intereacts with EC2/S3 state.
package main

import (
	"log"
	"os"

	"github.com/gravitational/provisioner"
	"github.com/gravitational/trace"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Define this in a variable to make it testable
var getCommand = func() provisioner.CommandRunner {
	var cfg provisioner.LoaderConfig
	app := kingpin.New("provisioner", "Terraform based provisioners for ops center")
	return provisioner.LoadCommands(app, &cfg)
}

// main is very minimal to make it testable. By doing this, we turn all logic
// into a separate package and make it **importable** to other project
func main() {
	command := getCommand()
	if err := command.Run(os.Args[1:]); err != nil {
		log.Fatalf("[ERROR]: %v", trace.DebugReport(err))
	}
}
