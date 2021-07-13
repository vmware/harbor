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

package dao

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // import pgsql driver for migrator
	_ "github.com/golang-migrate/migrate/v4/source/file"       // import local file driver for migrator
	_ "github.com/lib/pq"                                      // register pgsql driver
)

const defaultMigrationPath = "migrations/postgresql/"

type pgsql struct {
	host         string
	port         string
	usr          string
	pwd          string
	database     string
	sslmode      string
	maxIdleConns int
	maxOpenConns int
}

// Name returns the name of PostgreSQL
func (p *pgsql) Name() string {
	return "PostgreSQL"
}

// String ...
func (p *pgsql) String() string {
	return fmt.Sprintf("type-%s host-%s port-%s database-%s sslmode-%q",
		p.Name(), p.host, p.port, p.database, p.sslmode)
}

// NewPGSQL returns an instance of postgres
func NewPGSQL(host string, port string, usr string, pwd string, database string, sslmode string, maxIdleConns int, maxOpenConns int) Database {
	if len(sslmode) == 0 {
		sslmode = "disable"
	}
	return &pgsql{
		host:         host,
		port:         port,
		usr:          usr,
		pwd:          pwd,
		database:     database,
		sslmode:      sslmode,
		maxIdleConns: maxIdleConns,
		maxOpenConns: maxOpenConns,
	}
}

// Register registers pgSQL to orm with the info wrapped by the instance.
func (p *pgsql) Register(alias ...string) error {
	if err := utils.TestTCPConn(fmt.Sprintf("%s:%s", p.host, p.port), 60, 2); err != nil {
		return err
	}

	if err := orm.RegisterDriver("postgres", orm.DRPostgres); err != nil {
		return err
	}

	an := "default"
	if len(alias) != 0 {
		an = alias[0]
	}
	info := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC",
		p.host, p.port, p.usr, p.pwd, p.database, p.sslmode)

	if err := orm.RegisterDataBase(an, "postgres", info, p.maxIdleConns, p.maxOpenConns); err != nil {
		return err
	}

	// Due to the issues of beego v1.12.1 and v1.12.2, we set the max open conns ourselves.
	// See https://github.com/goharbor/harbor/issues/12403
	// and https://github.com/astaxie/beego/issues/4059 for more info.
	db, _ := orm.GetDB(an)
	db.SetMaxOpenConns(p.maxOpenConns)
	db.SetConnMaxLifetime(5 * time.Minute)

	return nil
}

// UpgradeSchema calls migrate tool to upgrade schema to the latest based on the SQL scripts.
func (p *pgsql) UpgradeSchema() error {
	port, err := strconv.ParseInt(p.port, 10, 64)
	if err != nil {
		return err
	}
	m, err := NewMigrator(&models.PostGreSQL{
		Host:     p.host,
		Port:     int(port),
		Username: p.usr,
		Password: p.pwd,
		Database: p.database,
		SSLMode:  p.sslmode,
	})
	if err != nil {
		return err
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil || dbErr != nil {
			log.Warningf("Failed to close migrator, source error: %v, db error: %v", srcErr, dbErr)
		}
	}()
	log.Infof("Upgrading schema for pgsql ...")
	err = m.Up()
	if err == migrate.ErrNoChange {
		log.Infof("No change in schema, skip.")
	} else if err != nil { // migrate.ErrLockTimeout will be thrown when another process is doing migration and timeout.
		log.Errorf("Failed to upgrade schema, error: %q", err)
		return err
	}
	return nil
}

// NewMigrator creates a migrator base on the information
func NewMigrator(database *models.PostGreSQL) (*migrate.Migrate, error) {
	dbURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(database.Username, database.Password),
		Host:     fmt.Sprintf("%s:%d", database.Host, database.Port),
		Path:     database.Database,
		RawQuery: fmt.Sprintf("sslmode=%s", database.SSLMode),
	}

	// For UT
	path := os.Getenv("POSTGRES_MIGRATION_SCRIPTS_PATH")
	if len(path) == 0 {
		path = defaultMigrationPath
	}
	srcURL := fmt.Sprintf("file://%s", path)
	m, err := migrate.New(srcURL, dbURL.String())
	if err != nil {
		return nil, err
	}
	m.Log = newMigrateLogger()
	return m, nil
}
