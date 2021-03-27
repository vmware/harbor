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

package notary

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/distribution/distribution/registry/auth/token"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/core/config"
	tokenutil "github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/lib/log"
	model2 "github.com/goharbor/harbor/src/pkg/signature/notary/model"
	"github.com/theupdateframework/notary"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/trustpinning"
	"github.com/theupdateframework/notary/tuf/data"

	digest "github.com/opencontainers/go-digest"
)

var (
	notaryCachePath = "/tmp/notary-cache"
	trustPin        trustpinning.TrustPinConfig
	mockRetriever   notary.PassRetriever
)

func init() {
	mockRetriever = func(keyName, alias string, createNew bool, attempts int) (passphrase string, giveup bool, err error) {
		passphrase = "hardcode"
		giveup = false
		err = nil
		return
	}
	trustPin = trustpinning.TrustPinConfig{}
}

// GetInternalTargets wraps GetTargets to read config values for getting full-qualified repo from internal notary instance.
func GetInternalTargets(notaryEndpoint string, username string, repo string) ([]model2.Target, error) {
	ext, err := config.ExtEndpoint()
	if err != nil {
		log.Errorf("Error while reading external endpoint: %v", err)
		return nil, err
	}
	endpoint := strings.Split(ext, "//")[1]
	fqRepo := path.Join(endpoint, repo)
	return GetTargets(notaryEndpoint, username, fqRepo)
}

// GetTargets is a help function called by API to fetch signature information of a given repository.
// Per docker's convention the repository should contain the information of endpoint, i.e. it should look
// like "192.168.0.1/library/ubuntu", instead of "library/ubuntu" (fqRepo for fully-qualified repo)
func GetTargets(notaryEndpoint string, username string, fqRepo string) ([]model2.Target, error) {
	res := []model2.Target{}
	t, err := tokenutil.MakeToken(username, tokenutil.Notary,
		[]*token.ResourceActions{
			{
				Type:    "repository",
				Name:    fqRepo,
				Actions: []string{"pull"},
			}})
	if err != nil {
		return nil, err
	}
	authorizer := &notaryAuthorizer{
		token: t.Token,
	}
	tr := NewTransport(commonhttp.GetHTTPTransport(commonhttp.SecureTransport), authorizer)
	gun := data.GUN(fqRepo)
	notaryRepo, err := client.NewFileCachedRepository(notaryCachePath, gun, notaryEndpoint, tr, mockRetriever, trustPin)
	if err != nil {
		return res, err
	}
	targets, err := notaryRepo.ListTargets(data.CanonicalTargetsRole)
	if _, ok := err.(client.ErrRepositoryNotExist); ok {
		log.Debugf("Repository not exist, repo: %s, error: %v, returning empty signature", fqRepo, err)
		return res, nil
	} else if err != nil {
		return res, err
	}
	// Remove root.json such that when remote repository is removed the local cache can't be reused.
	rootJSON := path.Join(notaryCachePath, "tuf", fqRepo, "metadata/root.json")
	rmErr := os.Remove(rootJSON)
	if rmErr != nil {
		log.Warningf("Failed to clear cached root.json: %s, error: %v, when repo is removed from notary the signature status maybe incorrect", rootJSON, rmErr)
	}
	for _, t := range targets {
		res = append(res, model2.Target{
			Tag:    t.Name,
			Hashes: t.Hashes,
		})
	}
	return res, nil
}

// DigestFromTarget get a target and return the value of digest, in accordance to Docker-Content-Digest
func DigestFromTarget(t model2.Target) (string, error) {
	sha, ok := t.Hashes["sha256"]
	if !ok {
		return "", fmt.Errorf("no valid hash, expecting sha256")
	}
	return digest.NewDigestFromHex("sha256", hex.EncodeToString(sha)).String(), nil
}

type notaryAuthorizer struct {
	token string
}

func (n *notaryAuthorizer) Modify(req *http.Request) error {
	req.Header.Add(http.CanonicalHeaderKey("Authorization"),
		fmt.Sprintf("Bearer %s", n.token))
	return nil
}
