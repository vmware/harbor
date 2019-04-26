package job

import (
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/job/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	testClient Client
)

const ID = "u-1234-5678-9012"

func TestMain(m *testing.M) {
	mockServer := test.NewJobServiceServer()
	defer mockServer.Close()
	testClient = NewDefaultClient(mockServer.URL, "")
	rc := m.Run()
	if rc != 0 {
		os.Exit(rc)
	}
}

func TestSubmitJob(t *testing.T) {
	assert := assert.New(t)
	d := &models.JobData{
		Name:     "replication",
		Metadata: nil,
	}
	uuid, err := testClient.SubmitJob(d)
	assert.Nil(err)
	assert.Equal(ID, uuid)

}

func TestGetJobLog(t *testing.T) {
	assert := assert.New(t)
	_, err1 := testClient.GetJobLog("non")
	assert.NotNil(err1)

	b2, err2 := testClient.GetJobLog(ID)
	assert.Nil(err2)
	text := string(b2)
	assert.Contains(text, "The content in this file is for mocking the get log api.")
}

func TestGetExecutions(t *testing.T) {
	assert := assert.New(t)
	exes, err := testClient.GetExecutions(ID)
	assert.Nil(err)
	stat := exes[0]
	assert.Equal(ID, stat.Info.JobID)
}

func TestPostAction(t *testing.T) {
	assert := assert.New(t)
	err := testClient.PostAction(ID, "fff")
	assert.NotNil(err)
	err2 := testClient.PostAction(ID, "stop")
	assert.Nil(err2)
}
