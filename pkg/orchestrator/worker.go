package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/moolen/tdep/pkg/analysis"
	"github.com/moolen/tdep/pkg/state"
	log "github.com/sirupsen/logrus"
	"k8s.io/utils/pointer"
)

func work(client *tfe.Client, ctx context.Context, wg *sync.WaitGroup) {
	log.Infof("starting worker")
	defer wg.Done()
worker:
	for {
		select {
		case <-ctx.Done():
			log.Infof("stopping worker")
			return
		case entry := <-queue:
			log.Infof("working on: %s/%s", entry.OrganizationName, entry.WorkspaceName)
			// get state versions
			ls, err := client.StateVersions.List(ctx, &tfe.StateVersionListOptions{
				Organization: entry.OrganizationName,
				Workspace:    entry.WorkspaceName,
			})
			if err != nil {
				log.Error(err)
				continue
			}
			// see if we find matching state for this run id
			var sv *tfe.StateVersion
			for _, item := range ls.Items {
				if item.Run.ID == entry.RunID {
					sv = item
					break
				}
			}
			// todo: go through pagination if we didn't find sv
			if sv == nil {
				log.Errorf("could not find state version for %s/%s of run %s", entry.OrganizationName, entry.WorkspaceName, entry.RunID)
				continue
			}
			stateBytes, err := client.StateVersions.Download(ctx, sv.DownloadURL)
			if err != nil {
				log.Error(err)
				continue
			}
			var tfstate state.TerraformState
			err = json.Unmarshal(stateBytes, &tfstate)
			if err != nil {
				log.Error(err)
				continue
			}

			// TODO: this is completely wrong!
			//       it triggers the dependencies, not the dependents 😅😅.
			//       (1) create a map that contains the dependencies across all workspaces
			//       (2) lookup all dependents and trigger run
			node, err := analysis.AnalyzeDependencies(&tfstate)
			if err != nil {
				log.Error(err)
				continue
			}
			for _, dep := range node.Dependencies {
				log.Infof("found dep %s/%s", dep.Organization, dep.Workspace)

				// fetch id
				workspace, err := client.Workspaces.Read(ctx, dep.Organization, dep.Workspace)
				if err != nil {
					log.Error(err)
					continue worker
				}

				run, err := client.Runs.Create(ctx, tfe.RunCreateOptions{
					Workspace: workspace,
					Message:   pointer.StringPtr(fmt.Sprintf("created by tdep / triggered via %s/%s", entry.OrganizationName, entry.WorkspaceName)),
					// inherited by workspace default setting
					// TODO: consider adding an option to override this
					// TODO: consider adding an webhook of auto-apply is disabled
					//       for this workspace
					// TODO: how can we tell the user when issues with run failures arise?
					// AutoApply: <inherited>,
				})
				if err != nil {
					log.Error(err)
					continue
				}
				log.Infof("created run %s", run.ID)
			}
		}
	}
}
