package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/font-util/cmd/font-util/cmds"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "font-util",
	Short:   "A general-purpose font manipulation tool",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return logging.InitLoggerFromCobra(cmd)
	},
}

func main() {
	err := logging.AddLoggingSectionToRootCommand(rootCmd, "font-util")
	cobra.CheckErr(err)

	helpSystem := help.NewHelpSystem()
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	ttc2ttfCmd, err := cmds.NewTtc2TtfCommand()
	cobra.CheckErr(err)
	cobraCmd, err := cli.BuildCobraCommand(ttc2ttfCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: "font-util"}),
	)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
