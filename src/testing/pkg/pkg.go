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

package pkg

//go:generate mockery --case snake --dir ../../pkg/artifact --name Manager --output ./artifact --outpkg artifact
//go:generate mockery --case snake --dir ../../pkg/blob --name Manager --output ./blob --outpkg blob
//go:generate mockery --case snake --dir ../../vendor/github.com/distribution/distribution --name Manifest --output ./distribution --outpkg distribution
//go:generate mockery --case snake --dir ../../pkg/project --name Manager --output ./project --outpkg project
//go:generate mockery --case snake --dir ../../pkg/project/metadata --name Manager --output ./project/metadata --outpkg metadata
//go:generate mockery --case snake --dir ../../pkg/quota --name Manager --output ./quota --outpkg quota
//go:generate mockery --case snake --dir ../../pkg/quota/driver --name Driver --output ./quota/driver --outpkg driver
//go:generate mockery --case snake --dir ../../pkg/scan/report --name Manager --output ./scan/report --outpkg report
//go:generate mockery --case snake --dir ../../pkg/scan/rest/v1 --all --output ./scan/rest/v1 --outpkg v1
//go:generate mockery --case snake --dir ../../pkg/scan/scanner --all --output ./scan/scanner --outpkg scanner
//go:generate mockery --case snake --dir ../../pkg/scheduler --name Scheduler --output ./scheduler --outpkg scheduler
//go:generate mockery --case snake --dir ../../pkg/task --name Manager --output ./task --outpkg task
//go:generate mockery --case snake --dir ../../pkg/task --name ExecutionManager --output ./task --outpkg task
//go:generate mockery --case snake --dir ../../pkg/user --name Manager --output ./user --outpkg user
//go:generate mockery --case snake --dir ../../pkg/rbac --name Manager --output ./rbac --outpkg rbac
//go:generate mockery --case snake --dir ../../pkg/rbac/dao --name DAO --output ./rbac/dao --outpkg dao
//go:generate mockery --case snake --dir ../../pkg/robot --name Manager --output ./robot --outpkg robot
//go:generate mockery --case snake --dir ../../pkg/robot/dao --name DAO --output ./robot/dao --outpkg dao
//go:generate mockery --case snake --dir ../../pkg/repository/dao --name DAO --output ./repository/dao --outpkg dao
//go:generate mockery --case snake --dir ../../pkg/notification/job/dao --name DAO --output ./notification/job/dao --outpkg dao
//go:generate mockery --case snake --dir ../../pkg/notification/policy/dao --name DAO --output ./notification/policy/dao --outpkg dao
//go:generate mockery --case snake --dir ../../pkg/notification/policy --name Manager --output ./notification/policy --outpkg notification
//go:generate mockery --case snake --dir ../../pkg/immutable/dao --name DAO --output ./immutable/dao --outpkg dao
//go:generate mockery --case snake --dir ../../pkg/ldap --name Manager --output ./ldap --outpkg ldap
//go:generate mockery --case snake --dir ../../pkg/allowlist --name Manager --output ./allowlist --outpkg robot
//go:generate mockery --case snake --dir ../../pkg/allowlist/dao --name DAO --output ./allowlist/dao --outpkg dao
//go:generate mockery --case snake --dir ../../pkg/reg/dao --name DAO --output ./reg/dao --outpkg dao
//go:generate mockery --case snake --dir ../../pkg/reg --name Manager --output ./reg --outpkg manager
//go:generate mockery --case snake --dir ../../pkg/replication/dao --name DAO --output ./replication/dao --outpkg dao
//go:generate mockery --case snake --dir ../../pkg/replication --name Manager --output ./replication --outpkg manager
