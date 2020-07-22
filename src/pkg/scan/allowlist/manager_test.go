package allowlist

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	// databases := []string{"mysql", "sqlite"}
	databases := []string{"postgresql"}
	for _, database := range databases {
		log.Infof("run test cases for database: %s", database)

		result := 1
		switch database {
		case "postgresql":
			dao.PrepareTestForPostgresSQL()
		default:
			log.Fatalf("invalid database: %s", database)
		}

		result = m.Run()

		if result != 0 {
			os.Exit(result)
		}
	}
}

func TestDefaultManager_CreateEmpty(t *testing.T) {
	dm := NewDefaultManager()
	assert.NoError(t, dm.CreateEmpty(99))
	assert.Error(t, dm.CreateEmpty(99))
}

func TestDefaultManager_Get(t *testing.T) {
	dm := NewDefaultManager()
	// return empty list
	l, err := dm.Get(1234)
	assert.Nil(t, err)
	assert.Empty(t, l.Items)
}
