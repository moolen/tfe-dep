package state

import (
	"context"
	"encoding/json"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
)

type TerraformState struct {
	Version          int        `json:"version"`
	TerraformVersion string     `json:"terraform_version"`
	Resources        []Resource `json:"resources"`
	Lineage          string     `json:"lineage"`
}

type Resource struct {
	Mode      string             `json:"mode"`
	Type      string             `json:"type"`
	Provider  string             `json:"provider"`
	Instances []ResourceInstance `json:"instances"`
}

type ResourceInstance struct {
	SchemaVersion int                `json:"schema_version"`
	Attributes    InstanceAttributes `json:"attributes"`
}

type InstanceAttributes struct {
	Workspace    string `json:"workspace"`
	Organization string `json:"organization"`
}

type InstanceAttributeConfig struct {
	Value map[string]string `json:"value"`
}

var stateCache map[string]*TerraformState

func init() {
	stateCache = make(map[string]*TerraformState)
}

func Read(tfeClient *tfe.Client, organization, workspace string) (*TerraformState, error) {
	cacheKey := fmt.Sprintf("%s/%s", organization, workspace)
	if val, ok := stateCache[cacheKey]; ok {
		return val, nil
	}

	ws, err := tfeClient.Workspaces.Read(context.Background(), organization, workspace)
	if err != nil {
		return nil, err
	}
	sv, err := tfeClient.StateVersions.ReadCurrent(context.Background(), ws.ID)
	if err != nil {
		return nil, err
	}
	bdy, err := tfeClient.StateVersions.Download(context.Background(), sv.DownloadURL)
	if err != nil {
		return nil, err
	}
	var state TerraformState
	err = json.Unmarshal(bdy, &state)
	if err != nil {
		return nil, err
	}
	stateCache[cacheKey] = &state
	return &state, nil
}
