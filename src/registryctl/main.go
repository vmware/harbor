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

package main

import (
	"crypto/tls"
	"flag"
	"net/http"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/registryctl/config"
	"github.com/goharbor/harbor/src/registryctl/handlers"

	_ "github.com/docker/distribution/registry/storage/driver/azure"
	_ "github.com/docker/distribution/registry/storage/driver/filesystem"
	_ "github.com/docker/distribution/registry/storage/driver/gcs"
	_ "github.com/docker/distribution/registry/storage/driver/inmemory"
	_ "github.com/docker/distribution/registry/storage/driver/middleware/cloudfront"
	_ "github.com/docker/distribution/registry/storage/driver/middleware/redirect"
	_ "github.com/docker/distribution/registry/storage/driver/oss"
	_ "github.com/docker/distribution/registry/storage/driver/s3-aws"
	_ "github.com/docker/distribution/registry/storage/driver/swift"
)

// RegistryCtl for registry controller
type RegistryCtl struct {
	ServerConf config.Configuration
	Handler    http.Handler
}

// Start the registry controller
func (s *RegistryCtl) Start() {
	regCtl := &http.Server{
		Addr:      ":" + s.ServerConf.Port,
		Handler:   s.Handler,
		TLSConfig: common_http.NewServerTLSConfig(),
	}

	var err error
	if s.ServerConf.Protocol == "https" {
		if common_http.InternalEnableVerifyClientCert() {
			regCtl.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
		err = regCtl.ListenAndServeTLS(s.ServerConf.HTTPSConfig.Cert, s.ServerConf.HTTPSConfig.Key)
	} else {
		err = regCtl.ListenAndServe()
	}

	if err != nil {
		log.Fatal(err)
	}

	return
}

func main() {
	configPath := flag.String("c", "", "Specify registryCtl configuration file path")
	flag.Parse()

	if configPath == nil || len(*configPath) == 0 {
		flag.Usage()
		log.Fatal("Config file should be specified")
	}
	if err := config.DefaultConfig.Load(*configPath, true); err != nil {
		log.Fatalf("Failed to load configurations with error: %s\n", err)
	}

	regCtl := &RegistryCtl{
		ServerConf: *config.DefaultConfig,
		Handler:    handlers.NewHandlerChain(*config.DefaultConfig),
	}
	regCtl.Start()
}
