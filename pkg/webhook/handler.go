package webhook

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/moolen/tdep/pkg/orchestrator"
)

// Payload is the notification payload sent from tf cloud
// see: https://www.terraform.io/cloud-docs/api-docs/notification-configurations#sample-payload
type Payload struct {
	PayloadVersion              int                   `json:"payload_version"`
	NotificationConfigurationID string                `json:"notification_configuration_id"`
	RunURL                      string                `json:"run_url"`
	RunID                       string                `json:"run_id"`
	RunMessage                  string                `json:"run_message"`
	WorkspaceID                 string                `json:"workspace_id"`
	WorkspaceName               string                `json:"workspace_name"`
	OrganizationName            string                `json:"organization_name"`
	Notifications               []NotificationPayload `json:"notifications"`
}

type NotificationPayload struct {
	Message      string `json:"message"`
	Trigger      string `json:"trigger"`
	RunStatus    string `json:"run_status"`
	RunUpdatedAt string `json:"run_updated_at"`
	RunUpdatedBy string `json:"run_updated_by"`
}

const MaxFileSize = 1024 * 1024

func (s *ReceiverServer) handlePayload() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		rd := http.MaxBytesReader(w, r.Body, MaxFileSize)
		defer rd.Close()
		bt, err := io.ReadAll(rd)
		if err != nil {
			s.logger.Errorf("error reading http body: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var payload Payload
		err = json.Unmarshal(bt, &payload)
		if err != nil {
			s.logger.Errorf("unable to unmarshal payload: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = orchestrator.Enqueue(orchestrator.Entry{
			RunID:            payload.RunID,
			WorkspaceName:    payload.WorkspaceName,
			WorkspaceID:      payload.WorkspaceID,
			OrganizationName: payload.OrganizationName,
		})
		if err != nil {
			s.logger.Errorf("unable to enqueue entry: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
