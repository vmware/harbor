package metadata

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

// QuotaMetaData defines quota related event data
type QuotaMetaData struct {
	Project  *models.Project
	RepoName string
	Tag      string
	Digest   string
	// used to define the event topic
	Level int
	// the msg contains the limitation and current usage of quota
	Msg     string
	OccurAt time.Time
}

// Resolve quota exceed into common image event
func (q *QuotaMetaData) Resolve(evt *event.Event) error {
	var topic string
	switch q.Level {
	case 1:
		topic = event2.TopicQuotaExceed
	case 2:
		topic = event2.TopicQuotaWarning
	default:
		return errors.New("not supported quota status")
	}

	data := &event2.QuotaEvent{
		EventType: topic,
		Project:   q.Project,
		OccurAt:   q.OccurAt,
		RepoName:  q.RepoName,
		Msg:       q.Msg,
	}
	if q.Tag != "" || q.Digest != "" {
		data.Resource = &event2.ImgResource{
			Tag:    q.Tag,
			Digest: q.Digest,
		}
	}

	evt.Topic = topic
	evt.Data = data
	return nil
}
