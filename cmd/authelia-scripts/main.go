//usr/bin/env go run "$0" "$@"; exit

package main

import (
	"github.com/authelia/authelia/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logLevel string

// AutheliaCommandDefinition is the definition of one authelia-scripts command.
type AutheliaCommandDefinition struct {
	Name        string
	Short       string
	Long        string
	CommandLine string
	Args        cobra.PositionalArgs
	Func        func(cmd *cobra.Command, args []string)
	SubCommands []*cobra.Command
}

// CobraCommands list of cobra commands
type CobraCommands = []*cobra.Command

// Commands is the list of commands of authelia-scripts
var Commands = []AutheliaCommandDefinition{
	AutheliaCommandDefinition{
		Name:  "bootstrap",
		Short: "Prepare environment for development and testing.",
		Long: `Prepare environment for development and testing. This command prepares docker
		images and download tools like Kind for Kubernetes testing.`,
		Func: Bootstrap,
	},
	AutheliaCommandDefinition{
		Name:  "build",
		Short: "Build Authelia binary and static assets",
		Func:  Build,
	},
	AutheliaCommandDefinition{
		Name:  "clean",
		Short: "Clean build artifacts",
		Func:  Clean,
	},
	AutheliaCommandDefinition{
		Name:        "docker",
		Short:       "Commands related to building and publishing docker image",
		SubCommands: CobraCommands{DockerBuildCmd, DockerPushCmd, DockerManifestCmd},
	},
	AutheliaCommandDefinition{
		Name:  "hash-password [password]",
		Short: "Compute hash of a password for creating a file-based users database",
		Func:  HashPassword,
		Args:  cobra.MinimumNArgs(1),
	},
	AutheliaCommandDefinition{
		Name:  "serve [config]",
		Short: "Serve compiled version of Authelia",
		Func:  ServeCmd,
		Args:  cobra.MinimumNArgs(1),
	},
	AutheliaCommandDefinition{
		Name:  "suites",
		Short: "Compute hash of a password for creating a file-based users database",
		SubCommands: CobraCommands{
			SuitesTestCmd,
			SuitesListCmd,
			SuitesSetupCmd,
			SuitesTeardownCmd,
		},
	},
	AutheliaCommandDefinition{
		Name:  "ci",
		Short: "Run continuous integration script",
		Func:  RunCI,
	},
	AutheliaCommandDefinition{
		Name:  "unittest",
		Short: "Run unit tests",
		Func:  RunUnitTest,
	},
	AutheliaCommandDefinition{
		Name:  "migrate",
		Short: "Migrate data from v3 to v4",
		SubCommands: CobraCommands{
			MigrateLocalCmd,
			MigrateMongoCmd,
		},
	},
}

func levelStringToLevel(level string) log.Level {
	if level == "debug" {
		return log.DebugLevel
	} else if level == "warning" {
		return log.WarnLevel
	}
	return log.InfoLevel
}

func main() {
	var rootCmd = &cobra.Command{Use: "authelia-scripts"}
	cobraCommands := make([]*cobra.Command, 0)

	for _, autheliaCommand := range Commands {
		var fn func(cobraCmd *cobra.Command, args []string)

		if autheliaCommand.CommandLine != "" {
			cmdline := autheliaCommand.CommandLine
			fn = func(cobraCmd *cobra.Command, args []string) {
				cmd := utils.CommandWithStdout(cmdline, args...)
				err := cmd.Run()
				if err != nil {
					panic(err)
				}
			}
		} else if autheliaCommand.Func != nil {
			fn = autheliaCommand.Func
		}

		command := &cobra.Command{
			Use:   autheliaCommand.Name,
			Short: autheliaCommand.Short,
		}

		if autheliaCommand.Long != "" {
			command.Long = autheliaCommand.Long
		}

		if fn != nil {
			command.Run = fn
		}

		if autheliaCommand.Args != nil {
			command.Args = autheliaCommand.Args
		}

		if autheliaCommand.SubCommands != nil {
			command.AddCommand(autheliaCommand.SubCommands...)
		}

		cobraCommands = append(cobraCommands, command)
	}

	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the log level for the command")
	rootCmd.AddCommand(cobraCommands...)
	cobra.OnInitialize(initConfig)

	err := rootCmd.Execute()

	if err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	log.SetLevel(levelStringToLevel(logLevel))
}
