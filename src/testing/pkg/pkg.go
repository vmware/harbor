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

//go:generate mockery --case snake --dir ../../pkg/blob --name Manager --output ./blob --outpkg blob
//go:generate mockery --case snake --dir ../../vendor/github.com/docker/distribution --name Manifest --output ./distribution --outpkg distribution
//go:generate mockery --case snake --dir ../../pkg/quota --name Manager --output ./quota --outpkg quota
//go:generate mockery --case snake --dir ../../pkg/quota/driver --name Driver --output ./quota/driver --outpkg driver
//go:generate mockery --case snake --dir ../../pkg/scan/report --name Manager --output ./scan/report --outpkg report
//go:generate mockery --case snake --dir ../../pkg/scan/rest/v1 --all --output ./scan/rest/v1 --outpkg v1
//go:generate mockery --case snake --dir ../../pkg/scan/scanner --all --output ./scan/scanner --outpkg scanner
//go:generate mockery --case snake --dir ../../pkg/scheduler --name Scheduler --output ./scheduler --outpkg scheduler
