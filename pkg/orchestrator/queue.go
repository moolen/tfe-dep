package orchestrator

var queue chan Entry

type Entry struct {
	RunID            string
	WorkspaceName    string
	WorkspaceID      string
	OrganizationName string
}

func init() {
	queue = make(chan Entry, 1000)
}

func Enqueue(entry Entry) error {
	queue <- entry
	return nil
}
