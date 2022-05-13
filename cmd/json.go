package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/moolen/tdep/pkg/analysis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// jsonCmd represents the json command
var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "renders dependency graph as json",
	Long:  "renders dependency graph as json",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		tfeClient, err := tfe.NewClient(&tfe.Config{
			Token: os.Getenv("TFE_TOKEN"),
		})
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(io.Discard)
		node, err := analysis.Analyze(tfeClient, organization, workspace)
		if err != nil && errors.Is(err, analysis.ErrCircular) {
			log.Error(err)
		}
		out, err := json.Marshal(node)
		if err != nil {
			log.Fatal(err)
		}
		os.Stdout.Write(out)
	},
}

func init() {
	rootCmd.AddCommand(jsonCmd)
}
