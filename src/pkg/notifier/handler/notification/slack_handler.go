package notification

import (
	"errors"
	"fmt"

	"bytes"
	"encoding/json"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"strings"
	"text/template"
)

const (
	// SlackBodyTemplate defines Slack request body template
	SlackBodyTemplate = `{
	"blocks": [
		{
            "type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*Harbor webhook events*"
			}
        },
        {
            "type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*event_type:* {{.Type}}"
			}
        },
        {
            "type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*occur_at:* <!date^{{.OccurAt}}^{date} at {time}|February 18th, 2014 at 6:39 AM PST>"
			}
        },
        {	"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*operator:* {{.Operator}}"
			}
		},
        {	"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*event_data:*"
			}
		},
		{	"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "{{.EventData}}"
			}
		}
    ]}`
)

// SlackHandler preprocess event data to slack and start the hook processing
type SlackHandler struct {
}

// Handle handles event to slack
func (s *SlackHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("SlackHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return errors.New("invalid notification slack event")
	}

	return s.process(event)
}

// IsStateful ...
func (s *SlackHandler) IsStateful() bool {
	return false
}

func (s *SlackHandler) process(event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	// Create a slackJob to send message to slack
	j.Name = job.SlackJob

	// Convert payload to slack format
	payload, err := s.convert(event.Payload)
	if err != nil {
		return fmt.Errorf("convert payload to slack body failed: %v", err)
	}

	j.Parameters = map[string]interface{}{
		"payload":          payload,
		"address":          event.Target.Address,
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(event, j)
}

func (s *SlackHandler) convert(payLoad *model.Payload) (string, error) {
	data := make(map[string]interface{})
	data["Type"] = payLoad.Type
	data["OccurAt"] = payLoad.OccurAt
	data["Operator"] = payLoad.Operator
	eventData, err := json.MarshalIndent(payLoad.EventData, "", "\t")
	if err != nil {
		return "", fmt.Errorf("marshal from eventData %v failed: %v", payLoad.EventData, err)
	}
	data["EventData"] = "```" + strings.Replace(string(eventData), `"`, `\"`, -1) + "```"

	st, _ := template.New("slack").Parse(SlackBodyTemplate)
	var slackBuf bytes.Buffer
	if err := st.Execute(&slackBuf, data); err != nil {
		return "", fmt.Errorf("%v", err)
	}
	return slackBuf.String(), nil
}
