package middleware

import (
	"fmt"
	"regexp"

	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

const (
	// RepositorySubexp is the name for sub regex that maps to repository name in the url
	RepositorySubexp = "repository"
	// ReferenceSubexp is the name for sub regex that maps to reference (tag or digest) url
	ReferenceSubexp = "reference"
	// DigestSubexp is the name for sub regex that maps to digest in the url
	DigestSubexp = "digest"
)

var (
	// V2ManifestURLRe is the regular expression for matching request v2 handler to view/delete manifest
	V2ManifestURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/manifests/(?P<%s>%s|%s)$`, RepositorySubexp, reference.NameRegexp.String(), ReferenceSubexp, reference.TagRegexp.String(), digest.DigestRegexp.String()))
	// V2TagListURLRe is the regular expression for matching request to v2 handler to list tags
	V2TagListURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/tags/list`, RepositorySubexp, reference.NameRegexp.String()))
	// V2BlobURLRe is the regular expression for matching request to v2 handler to retrieve head/delete a blob
	V2BlobURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/blobs/(?P<%s>%s)$`, RepositorySubexp, reference.NameRegexp.String(), DigestSubexp, digest.DigestRegexp.String()))
	// V2BlobUploadURLRe is the regular expression for matching the request to v2 handler to upload a blob, the upload uuid currently is not put into a group
	V2BlobUploadURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/blobs/uploads[/a-zA-Z0-9\-_\.=]*$`, RepositorySubexp, reference.NameRegexp.String()))
	// V2CatalogURLRe is the regular expression for mathing the request to v2 handler to list catalog
	V2CatalogURLRe = regexp.MustCompile(`^/v2/_catalog$`)
)
