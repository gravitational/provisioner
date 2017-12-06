package provisioner

import (
	"fmt"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// initVarsCmd groups the top-level command for initializing the variables and its arguments
type initVarsCmd struct {
	*kingpin.CmdClause
	varsKey *string
}

func (cmd *initVarsCmd) perform(cfg LoaderConfig) error {
	loader, err := NewLoader(cfg)
	if err != nil {
		return trace.Wrap(err)
	}

	if err := loader.initVars(*cmd.varsKey); err != nil {
		return trace.Wrap(err)
	}

	return nil
}

// syncFilesCmd groups the top-level command for syncing file and its arguments
type syncFilesCmd struct {
	*kingpin.CmdClause
	paths     []string
	targetDir string
}

func (cmd *syncFilesCmd) perform(cfg LoaderConfig) error {
	loader, err := NewLoader(cfg)
	if err != nil {
		return trace.Wrap(err)
	}

	return loader.sync(cmd.paths, cmd.targetDir)
}

// findInstanceCmd groups the top-level command for finding instance by private ip and its arguments
type findInstanceCmd struct {
	*kingpin.CmdClause
	findPrivateIP string
}

func (cmd *findInstanceCmd) perform(cfg LoaderConfig) error {
	resource, err := findInstance(cmd.findPrivateIP, nil)
	if err != nil {
		return trace.Wrap(err)
	}

	fmt.Println(resource)
	return nil
}

// removeS3KeyCmd groups the top-level command for removing S3 keys and its arguments
type removeS3KeyCmd struct {
	*kingpin.CmdClause
	rmKey string
}

func (cmd *removeS3KeyCmd) perform(cfg LoaderConfig) error {
	loader, err := NewLoader(cfg)
	if err != nil {
		return trace.Wrap(err)
	}

	return loader.rm(cmd.rmKey)
}

// CommandRunner is interface to our main entrypoint to the cli
type CommandRunner interface {
	Run(args []string) error
}

// Command wraps config kingpin, cfg and all command in same struct
type Command struct {
	App          *kingpin.Application
	cfg          *LoaderConfig
	initVars     *initVarsCmd
	syncFiles    *syncFilesCmd
	findInstance *findInstanceCmd
	removeS3Key  *removeS3KeyCmd
}

// registerFindInstance define command and flags
func (c *Command) registerFindInstance() {
	// find-instance
	findInstance := findInstanceCmd{}
	findInstance.CmdClause = c.App.Command("find-instance", "Finds instance in terraform show output by private ip")
	findInstance.Flag("private-ip", "Private IP").Required().StringVar(&findInstance.findPrivateIP)

	c.findInstance = &findInstance
}

// registerSyncFile define command and flags
func (c *Command) registerSyncFile() {
	// sync files syncs files from s3 to the local bucket
	csync := syncFilesCmd{}
	csync.CmdClause = c.App.Command("sync-files", "Syncs files from S3 bucket to local storage")
	csync.Flag("region", "AWS region to inspect").Required().StringVar(&c.cfg.Region)
	csync.Flag("cluster-bucket", "Check bucket key for pre-stored value").Required().StringVar(&c.cfg.ClusterBucket)
	csync.Flag("prefix", "Path prefix").Required().StringsVar(&csync.paths)
	csync.Flag("target", "Target dir").Required().StringVar(&csync.targetDir)

	c.syncFiles = &csync
}

// registerInitVars define command and flags
func (c *Command) registerInitVars() {
	// init-vars inits cluster specific variables
	cinitVars := initVarsCmd{}
	cinitVars.CmdClause = c.App.Command("init-vars", "Initalize or load variables of the cluster. This command is idempotent.")

	cinitVars.Flag("vpc-id", "AWS VPC to inspect").Required().StringVar(&c.cfg.VPCID)
	cinitVars.Flag("region", "AWS region to inspect").Required().StringVar(&c.cfg.Region)
	cinitVars.Flag("cluster-bucket", "Check bucket key for pre-stored value").Required().StringVar(&c.cfg.ClusterBucket)
	cinitVars.Flag("template", "Path to vars template").Required().StringVar(&c.cfg.TemplatePath)
	cinitVars.varsKey = cinitVars.CmdClause.Flag("key", "Key with cluster specific variables").String()

	c.initVars = &cinitVars
}

// registerRemoveS3Key define command and flags
func (c *Command) registerRemoveS3Key() {
	crm := removeS3KeyCmd{}

	crm.CmdClause = c.App.Command("rm", "Remove key from bucket")
	crm.Flag("region", "AWS region to inspect").Required().StringVar(&c.cfg.Region)
	crm.Flag("cluster-bucket", "Check bucket key for pre-stored value").StringVar(&c.cfg.ClusterBucket)
	crm.Flag("key", "Bucket key to remove").Required().StringVar(&crm.rmKey)

	c.removeS3Key = &crm
}

// LoadCommands initializes main CommandRunner
func LoadCommands(app *kingpin.Application, cfg *LoaderConfig) *Command {
	c := Command{
		App: app,
		cfg: cfg,
	}

	c.registerInitVars()
	c.registerFindInstance()
	c.registerSyncFile()
	c.registerRemoveS3Key()

	return &c
}

// Run parses CLI argument and execute sub-command
func (c *Command) Run(args []string) error {
	var err error
	invokedCommad, err := c.App.Parse(args)

	if err != nil {
		return trace.Wrap(err)
	}

	switch invokedCommad {
	case c.initVars.FullCommand():
		err = c.initVars.perform(*c.cfg)
	case c.syncFiles.FullCommand():
		err = c.syncFiles.perform(*c.cfg)
	case c.findInstance.FullCommand():
		err = c.findInstance.perform(*c.cfg)
	case c.removeS3Key.FullCommand():
		err = c.removeS3Key.perform(*c.cfg)
	}

	if err != nil {
		log.Error("failed to run command: ", trace.DebugReport(err))
	}

	return err
}
