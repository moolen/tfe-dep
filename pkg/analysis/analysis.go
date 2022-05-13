package analysis

import (
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/moolen/tdep/pkg/state"
	log "github.com/sirupsen/logrus"
)

const (
	TfeStateType = "tfe_outputs"
)

var ErrCircular = errors.New("circular dependency")
var seen map[string]bool

func init() {
	seen = make(map[string]bool)
}

func Analyze(tfeClient *tfe.Client, organization, workspace string) (*Node, error) {
	tfstate, err := state.Read(tfeClient, organization, workspace)
	if err != nil {
		return nil, err
	}
	node := &Node{
		Organization: organization,
		Workspace:    workspace,
		TFState:      tfstate,
	}
	seen[node.Key()] = true
	err = recurse(node, tfeClient)
	return node, err
}

func recurse(currentNode *Node, tfeClient *tfe.Client) error {
	log.Infof("working on %s:%s", currentNode.Key(), currentNode.TFState.Lineage)
	for _, resource := range currentNode.TFState.Resources {
		if resource.Type != TfeStateType {
			continue
		}
		for _, instance := range resource.Instances {
			log.Infof("pointing to remote state at %s/%s", instance.Attributes.Organization, instance.Attributes.Workspace)
			newState, err := state.Read(tfeClient, instance.Attributes.Organization, instance.Attributes.Workspace)
			if err != nil {
				return err
			}
			targetNode := &Node{
				Organization: instance.Attributes.Organization,
				Workspace:    instance.Attributes.Workspace,
				TFState:      newState,
			}
			currentNode.Dependencies = append(currentNode.Dependencies, targetNode)
			if _, ok := seen[targetNode.Key()]; ok {
				return fmt.Errorf("%w: %s -> %s", ErrCircular, currentNode.Key(), targetNode.Key())
			}
			seen[targetNode.Key()] = true
			err = recurse(targetNode, tfeClient)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
