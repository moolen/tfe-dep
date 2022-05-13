package cmd

import (
	"log"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/moolen/tdep/pkg/plot"
	"github.com/spf13/cobra"
)

// plotCmd represents the plot command
var plotCmd = &cobra.Command{
	Use:   "plot",
	Short: "renders an graph svg",
	Long:  "renders an graph svg",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		tfeClient, err := tfe.NewClient(&tfe.Config{
			Token: os.Getenv("TFE_TOKEN"),
		})
		if err != nil {
			log.Fatal(err)
		}
		err = plot.Plot(tfeClient, organization, workspace)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(plotCmd)
}
