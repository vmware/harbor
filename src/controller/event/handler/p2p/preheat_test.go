package p2p

import (
	"context"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/p2p/preheat"
	pkg_artifact "github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	test_artifact "github.com/goharbor/harbor/src/testing/controller/artifact"
	test_preheat "github.com/goharbor/harbor/src/testing/controller/p2p/preheat"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// PreheatTestSuite is a test suite of testing preheat handler
type PreheatTestSuite struct {
	suite.Suite
	artifactCtl artifact.Controller
	preheatEnf  preheat.Enforcer

	handler *Handler
}

// TestPreheat is an entry method of running PreheatTestSuite
func TestPreheat(t *testing.T) {
	suite.Run(t, &PreheatTestSuite{})
}

// SetupSuite prepares env for running PreheatTestSuite
func (suite *PreheatTestSuite) SetupSuite() {
	fakeArtifactCtl := &test_artifact.Controller{}
	fakeArtifactCtl.On("GetByReference",
		context.TODO(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("*artifact.Option"),
	).Return(&artifact.Artifact{}, nil)
	fakeArtifactCtl.On("Get",
		context.TODO(),
		mock.AnythingOfType("int64"),
		mock.AnythingOfType("*artifact.Option"),
	).Return(&artifact.Artifact{
		Artifact: pkg_artifact.Artifact{
			Type: "IMAGE",
		},
	}, nil)

	fakeEnforcer := &test_preheat.FakeEnforcer{}
	fakeEnforcer.On("PreheatArtifact",
		context.TODO(),
		mock.AnythingOfType("*artifact.Artifact"),
	).Return(nil, nil)

	suite.artifactCtl = artifact.Ctl
	artifact.Ctl = fakeArtifactCtl
	suite.preheatEnf = preheat.Enf
	preheat.Enf = fakeEnforcer
	suite.handler = &Handler{
		Context: context.TODO,
	}
}

// TearDownSuite cleans the testing env
func (suite *PreheatTestSuite) TearDownSuite() {
	artifact.Ctl = suite.artifactCtl
	preheat.Enf = suite.preheatEnf
}

// TestIsStateful ...
func (suite *PreheatTestSuite) TestIsStateful() {
	b := suite.handler.IsStateful()
	suite.False(b, "handler is stateful")
}

// TestHandle ...
func (suite *PreheatTestSuite) TestHandle() {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "PreheatHandler String Error",
			args: args{
				data: "",
			},
			wantErr: true,
		},
		{
			name: "PreheatHandler 1",
			args: args{
				data: &event.PushArtifactEvent{
					ArtifactEvent: &event.ArtifactEvent{
						Artifact: &pkg_artifact.Artifact{
							Type:         "IMAGE",
							ID:           11,
							RepositoryID: 23,
						},
						Tags: []string{"v1.1", "v1.2"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "PreheatHandler 2",
			args: args{
				data: &event.ScanImageEvent{
					OccurAt: time.Now(),
					Artifact: &v1.Artifact{
						Repository: "library",
						Digest:     "sha256:1359608115b94599e5641638bac5aef1ddfaa79bb96057ebf41ebc8d33acf8a7",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "PreheatHandler 3",
			args: args{
				data: &event.ArtifactLabeledEvent{
					OccurAt:    time.Now(),
					ArtifactID: 1,
					LabelID:    2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := suite.handler.Handle(tt.args.data)
		if tt.wantErr {
			suite.Error(err, tt.name)
		} else {
			suite.NoError(err, tt.name)
		}
	}
}
