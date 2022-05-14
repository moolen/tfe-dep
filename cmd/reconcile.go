package cmd

import (
	"context"
	"encoding/json"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/moolen/tdep/pkg/analysis"
	"github.com/moolen/tdep/pkg/state"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// reconcileCmd represents the reconcile command
var reconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "reconciles the tf run triggers for all workspaces in a given organization.",
	Long:  "reconciles the tf run triggers for all workspaces in a given organization.",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		tfeClient, err := tfe.NewClient(&tfe.Config{
			Token: os.Getenv("TFE_TOKEN"),
		})
		if err != nil {
			log.Fatal(err)
		}
		wsList, err := tfeClient.Workspaces.List(context.Background(), organization, &tfe.WorkspaceListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, workspace := range wsList.Items {
			reconcileWorkspace(tfeClient, organization, workspace)
		}
	},
}

func reconcileWorkspace(tfeClient *tfe.Client, organization string, sourceWorkspace *tfe.Workspace) {
	log.Infof("fetching workspace state: %s/%s", organization, sourceWorkspace.Name)
	sv, err := tfeClient.StateVersions.ReadCurrent(context.Background(), sourceWorkspace.ID)
	if err != nil {
		log.Fatal(err)
	}
	stateBytes, err := tfeClient.StateVersions.Download(context.Background(), sv.DownloadURL)
	if err != nil {
		log.Fatal(err)
	}
	var tfstate state.TerraformState
	err = json.Unmarshal(stateBytes, &tfstate)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("analyzing dependencies")
	node, err := analysis.AnalyzeDependencies(&tfstate)
	if err != nil {
		log.Fatal(err)
	}
	for _, dep := range node.Dependencies {
		log.Infof("working on %s/%s", dep.Organization, dep.Workspace)
		dependatWorkspace, err := tfeClient.Workspaces.Read(context.Background(), dep.Organization, dep.Workspace)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("listing triggers in source workspace: %s", sourceWorkspace.Name)
		triggers, err := tfeClient.RunTriggers.List(context.Background(), sourceWorkspace.ID, &tfe.RunTriggerListOptions{
			RunTriggerType: tfe.RunTriggerOutbound,
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("found %d triggers", len(triggers.Items))
		var found bool
		for _, rt := range triggers.Items {
			log.Debugf("trigger: %#v", rt)
			if rt.WorkspaceName == dependatWorkspace.Name && rt.SourceableName == sourceWorkspace.Name {
				found = true
			}
		}
		if found {
			log.Infof("trigger already exists in dependant %s for source=%s", dependatWorkspace.Name, sourceWorkspace.Name)
			continue
		}
		if !found {
			log.Infof("creating trigger in dependant %s for source=%s", dependatWorkspace.Name, sourceWorkspace.Name)
			_, err := tfeClient.RunTriggers.Create(context.Background(), dependatWorkspace.ID, tfe.RunTriggerCreateOptions{
				Type:       "workspaces",
				Sourceable: sourceWorkspace,
			})
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(reconcileCmd)
}
