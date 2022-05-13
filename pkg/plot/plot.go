package plot

import (
	"bytes"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/moolen/tdep/pkg/analysis"
	log "github.com/sirupsen/logrus"
)

const (
	TfeStateType = "tfe_outputs"
)

func Plot(tfeClient *tfe.Client, organization, workspace string) error {
	var err error
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		return err
	}
	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
		}
		g.Close()
	}()
	node, err := analysis.Analyze(tfeClient, organization, workspace)
	if err != nil {
		log.Error(err)
	}
	if node == nil {
		return err
	}
	recurse(node, graph)
	log.Info("rendering graph")
	var buf bytes.Buffer
	if err := g.Render(graph, graphviz.PNG, &buf); err != nil {
		return err
	}
	return g.RenderFilename(graph, graphviz.PNG, "./graph.png")
}

func recurse(node *analysis.Node, graph *cgraph.Graph) {
	currentGraphNode, err := graph.CreateNode(node.Key())
	if err != nil {
		log.Fatal(err)
	}
	for _, target := range node.Dependencies {
		log.Infof("pointing to remote state at %s", target.Key())
		targetGraphNode, err := graph.CreateNode(target.Key())
		if err != nil {
			log.Error(err)
		}
		_, err = graph.CreateEdge("", currentGraphNode, targetGraphNode)
		if err != nil {
			log.Fatal(err)
		}
		recurse(target, graph)
	}
}
