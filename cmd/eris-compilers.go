package cmd

import (
	"os"

	"github.com/monax/compilers/version"

	"github.com/monax/cli/log"

	"github.com/spf13/cobra"
)

const VERSION = version.VERSION

var (
	Verbose bool
	Debug   bool
)

var CompilersCmd = &cobra.Command{
	Use:   "eris-compilers COMMAND [FLAG ...]",
	Short: "A client/server set up for automatic compilation of smart contracts",
	Long: `A client/server set up for automatic compilation of smart contracts

Made with <3 by Monax Industries.

Complete documentation is available at https://monax.io/docs`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetOutput(os.Stdout)
		log.SetLevel(log.WarnLevel)
		if Verbose {
			log.SetLevel(log.InfoLevel)
		} else if Debug {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func Execute() {
	AddCommands()
	AddGlobalFlags()
	CompilersCmd.Execute()
}

func AddCommands() {
	BuildServerCommand()
	BuildCompileCommand()
	BuildBinaryCommand()
}

func AddGlobalFlags() {
	CompilersCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", SetVerbose(), "verbose output")
	CompilersCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", SetDebug(), "debug level output")
}

func SetVerbose() bool {
	return false
}

func SetDebug() bool {
	return false
}
