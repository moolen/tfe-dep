package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/moolen/tdep/pkg/webhook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// orchestratorTriggerCmd represents the json command
var orchestratorTriggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "trigger the orchestrator for the given organization/workspace with the latest run id",
	Long:  "trigger the orchestrator for the given organization/workspace with the latest run id",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		tfeClient, err := tfe.NewClient(&tfe.Config{
			Token: os.Getenv("TFE_TOKEN"),
		})
		if err != nil {
			log.Fatal(err)
		}

		ws, err := tfeClient.Workspaces.Read(context.Background(), organization, workspace)
		if err != nil {
			log.Fatal(err)
		}

		bt, err := json.Marshal(&webhook.Payload{
			RunID:            ws.CurrentRun.ID,
			WorkspaceID:      ws.ID,
			WorkspaceName:    ws.Name,
			OrganizationName: ws.Organization.Name,
		})
		if err != nil {
			log.Fatal(err)
		}
		buf := bytes.NewBuffer(bt)
		res, err := http.Post("http://localhost:8080/webhook/", "application/json", buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("res: %s", res.StatusCode)

	},
}

func init() {
	orchestratorCmd.AddCommand(orchestratorTriggerCmd)
}
