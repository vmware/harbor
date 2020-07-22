// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event

import (
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/audit/model"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

// the event consumers can refer to this file to find all topics and the corresponding event structures

// const definition
const (
	TopicCreateProject     = "CREATE_PROJECT"
	TopicDeleteProject     = "DELETE_PROJECT"
	TopicPushArtifact      = "PUSH_ARTIFACT"
	TopicPullArtifact      = "PULL_ARTIFACT"
	TopicDeleteArtifact    = "DELETE_ARTIFACT"
	TopicDeleteRepository  = "DELETE_REPOSITORY"
	TopicCreateTag         = "CREATE_TAG"
	TopicDeleteTag         = "DELETE_TAG"
	TopicScanningFailed    = "SCANNING_FAILED"
	TopicScanningCompleted = "SCANNING_COMPLETED"
	// QuotaExceedTopic is topic for quota warning event, the usage reaches the warning bar of limitation, like 85%
	TopicQuotaWarning    = "QUOTA_WARNING"
	TopicQuotaExceed     = "QUOTA_EXCEED"
	TopicUploadChart     = "UPLOAD_CHART"
	TopicDownloadChart   = "DOWNLOAD_CHART"
	TopicDeleteChart     = "DELETE_CHART"
	TopicReplication     = "REPLICATION"
	TopicArtifactLabeled = "ARTIFACT_LABELED"
)

// CreateProjectEvent is the creating project event
type CreateProjectEvent struct {
	EventType string
	ProjectID int64
	Project   string
	Operator  string
	OccurAt   time.Time
}

// ResolveToAuditLog ...
func (c *CreateProjectEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    c.ProjectID,
		OpTime:       c.OccurAt,
		Operation:    "create",
		Username:     c.Operator,
		ResourceType: "project",
		Resource:     c.Project}
	return auditLog, nil
}

// DeleteProjectEvent is the deleting project event
type DeleteProjectEvent struct {
	EventType string
	ProjectID int64
	Project   string
	Operator  string
	OccurAt   time.Time
}

// ResolveToAuditLog ...
func (d *DeleteProjectEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    d.ProjectID,
		OpTime:       d.OccurAt,
		Operation:    "delete",
		Username:     d.Operator,
		ResourceType: "project",
		Resource:     d.Project}
	return auditLog, nil
}

// DeleteRepositoryEvent is the deleting repository event
type DeleteRepositoryEvent struct {
	EventType  string
	ProjectID  int64
	Repository string
	Operator   string
	OccurAt    time.Time
}

// ResolveToAuditLog ...
func (d *DeleteRepositoryEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    d.ProjectID,
		OpTime:       d.OccurAt,
		Operation:    "delete",
		Username:     d.Operator,
		ResourceType: "repository",
		Resource:     d.Repository,
	}
	return auditLog, nil
}

// ArtifactEvent is the pushing/pulling artifact event
type ArtifactEvent struct {
	EventType  string
	Repository string
	Artifact   *artifact.Artifact
	Tags       []string // when the artifact is pushed by digest, the tag here will be null
	Operator   string
	OccurAt    time.Time
}

// PushArtifactEvent is the pushing artifact event
type PushArtifactEvent struct {
	*ArtifactEvent
}

// ResolveToAuditLog ...
func (p *PushArtifactEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    p.Artifact.ProjectID,
		OpTime:       p.OccurAt,
		Operation:    "create",
		Username:     p.Operator,
		ResourceType: "artifact"}

	if len(p.Tags) == 0 {
		auditLog.Resource = fmt.Sprintf("%s:%s",
			p.Artifact.RepositoryName, p.Artifact.Digest)
	} else {
		auditLog.Resource = fmt.Sprintf("%s:%s",
			p.Artifact.RepositoryName, p.Tags[0])
	}

	return auditLog, nil
}

// PullArtifactEvent is the pulling artifact event
type PullArtifactEvent struct {
	*ArtifactEvent
}

// ResolveToAuditLog ...
func (p *PullArtifactEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    p.Artifact.ProjectID,
		OpTime:       p.OccurAt,
		Operation:    "pull",
		Username:     p.Operator,
		ResourceType: "artifact"}

	if len(p.Tags) == 0 {
		auditLog.Resource = fmt.Sprintf("%s:%s",
			p.Artifact.RepositoryName, p.Artifact.Digest)
	} else {
		auditLog.Resource = fmt.Sprintf("%s:%s",
			p.Artifact.RepositoryName, p.Tags[0])
	}

	// for pull public resource
	if p.Operator == "" {
		auditLog.Username = "anonymous"
	} else {
		auditLog.Username = p.Operator
	}

	return auditLog, nil
}

// DeleteArtifactEvent is the deleting artifact event
type DeleteArtifactEvent struct {
	*ArtifactEvent
}

// ResolveToAuditLog ...
func (p *DeleteArtifactEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    p.Artifact.ProjectID,
		OpTime:       p.OccurAt,
		Operation:    "delete",
		Username:     p.Operator,
		ResourceType: "artifact",
		Resource:     fmt.Sprintf("%s:%s", p.Artifact.RepositoryName, p.Artifact.Digest)}
	return auditLog, nil
}

// CreateTagEvent is the creating tag event
type CreateTagEvent struct {
	EventType        string
	Repository       string
	Tag              string
	AttachedArtifact *artifact.Artifact
	Operator         string
	OccurAt          time.Time
}

// ResolveToAuditLog ...
func (c *CreateTagEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    c.AttachedArtifact.ProjectID,
		OpTime:       c.OccurAt,
		Operation:    "create",
		Username:     c.Operator,
		ResourceType: "tag",
		Resource:     fmt.Sprintf("%s:%s", c.Repository, c.Tag)}
	return auditLog, nil
}

// DeleteTagEvent is the deleting tag event
type DeleteTagEvent struct {
	EventType        string
	Repository       string
	Tag              string
	AttachedArtifact *artifact.Artifact
	Operator         string
	OccurAt          time.Time
}

// ResolveToAuditLog ...
func (d *DeleteTagEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    d.AttachedArtifact.ProjectID,
		OpTime:       d.OccurAt,
		Operation:    "delete",
		Username:     d.Operator,
		ResourceType: "tag",
		Resource:     fmt.Sprintf("%s:%s", d.Repository, d.Tag)}
	return auditLog, nil
}

// ScanImageEvent is scanning image related event data to publish
type ScanImageEvent struct {
	EventType string
	Artifact  *v1.Artifact
	OccurAt   time.Time
	Operator  string
}

// ChartEvent is chart related event data to publish
type ChartEvent struct {
	EventType   string
	ProjectName string
	ChartName   string
	Versions    []string
	OccurAt     time.Time
	Operator    string
}

// QuotaEvent is project quota related event data to publish
type QuotaEvent struct {
	EventType string
	Project   *models.Project
	Resource  *ImgResource
	OccurAt   time.Time
	RepoName  string
	Msg       string
}

// ImgResource include image digest and tag
type ImgResource struct {
	Digest string
	Tag    string
}

// ReplicationEvent is replication related event data to publish
type ReplicationEvent struct {
	EventType         string
	ReplicationTaskID int64
	OccurAt           time.Time
	Status            string
}

// ArtifactLabeledEvent is event data of artifact labeled
type ArtifactLabeledEvent struct {
	ArtifactID int64
	LabelID    int64
	OccurAt    time.Time
	Operator   string
}
