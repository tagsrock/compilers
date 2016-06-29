package cmd

import (
	"os"

	"github.com/eris-ltd/eris-compilers/version"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
	"github.com/spf13/cobra"
)

const VERSION = version.VERSION

var (
	Verbose bool
	Debug   bool
)

var RootCmd = &cobra.Command{
	Use:   "eris-compilers COMMAND [FLAG ...]",
	Short: "A client/server set up for automatic compilation of smart contracts",
	Long: `A client/server set up for automatic compilation of smart contracts

Made with <3 by Eris Industries.

Complete documentation is available at https://docs.erisindustries.com` + "\nVersion:\n " + VERSION,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetOutput(os.Stdout)
		log.SetLevel(log.WarnLevel)
		if Verbose {
			log.SetLevel(log.InfoLevel)
		} else if Debug {
			log.SetLevel(log.DebugLevel)
		}
		common.InitErisDir()
	},
}

func Execute() {
	AddCommands()
	AddGlobalFlags()
	RootCmd.Execute()
}

func AddCommands() {
	BuildServerCommand()
	BuildCompileCommand()
}

func AddGlobalFlags() {
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", SetVerbose(), "verbose output")
	RootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", SetDebug(), "debug level output")
}

func SetVerbose() bool {
	return false
}

func SetDebug() bool {
	return false
}
