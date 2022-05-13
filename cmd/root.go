package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "tfe-dep",
	Short:        "Reads terraform cloud workspace dependencies",
	Long:         "Reads terraform cloud workspace dependencies",
	SilenceUsage: true,
}

var (
	workspace    string
	organization string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&workspace, "workspace", "", "workspace to analyze")
	rootCmd.PersistentFlags().StringVar(&organization, "organization", "", "organization to analyze")
	rootCmd.Flags().StringP("token", "t", viper.GetString("TFE_TOKEN"), "set tfe token; use TFE_TOKEN env var")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
