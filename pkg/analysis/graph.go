package analysis

import (
	"fmt"

	"github.com/moolen/tdep/pkg/state"
)

type Node struct {
	Workspace    string                `json:"workspace"`
	Organization string                `json:"organization"`
	TFState      *state.TerraformState `json:"-"`
	Dependencies []*Node               `json:"dependencies"`
}

func (n *Node) Key() string {
	return fmt.Sprintf("%s/%s", n.Organization, n.Workspace)
}
