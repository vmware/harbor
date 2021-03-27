package annotation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	reg "github.com/goharbor/harbor/src/pkg/registry"

	"github.com/distribution/distribution/manifest/schema2"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// V1alpha1 is the version of annotation parser
	V1alpha1 = "v1alpha1"
)

type v1alpha1Parser struct {
	regCli reg.Client
}

func (p *v1alpha1Parser) Parse(ctx context.Context, artifact *artifact.Artifact, manifest []byte) error {
	if artifact.ManifestMediaType != v1.MediaTypeImageManifest && artifact.ManifestMediaType != schema2.MediaTypeManifest {
		return nil
	}
	// get manifest
	mani := &v1.Manifest{}
	if err := json.Unmarshal(manifest, mani); err != nil {
		return err
	}

	// parse skip-list annotation io.goharor.artifact.v1alpha1.skip-list
	parseV1alpha1SkipList(artifact, mani)

	// parse icon annotation io.goharbor.artifact.v1alpha1.icon
	err := parseV1alpha1Icon(artifact, mani, p.regCli)
	if err != nil {
		return err
	}

	return nil
}

func parseV1alpha1SkipList(artifact *artifact.Artifact, manifest *v1.Manifest) {
	metadata := artifact.ExtraAttrs
	skipListAnnotationKey := fmt.Sprintf("%s.%s.%s", AnnotationPrefix, V1alpha1, SkipList)
	skipList, ok := manifest.Config.Annotations[skipListAnnotationKey]
	if ok {
		skipKeyList := strings.Split(skipList, ",")
		for _, skipKey := range skipKeyList {
			delete(metadata, skipKey)
		}
		artifact.ExtraAttrs = metadata
	}
}

func parseV1alpha1Icon(artifact *artifact.Artifact, manifest *v1.Manifest, reg reg.Client) error {
	iconAnnotationKey := fmt.Sprintf("%s.%s.%s", AnnotationPrefix, V1alpha1, Icon)
	var iconDigest string
	for _, layer := range manifest.Layers {
		_, ok := layer.Annotations[iconAnnotationKey]
		if ok {
			iconDigest = layer.Digest.String()
			break
		}
	}
	if iconDigest == "" {
		return nil
	}
	// pull icon layer
	_, icon, err := reg.PullBlob(artifact.RepositoryName, iconDigest)
	if err != nil {
		return err
	}
	// check the size of the size <= 1MB
	data, err := ioutil.ReadAll(io.LimitReader(icon, 1<<20))
	if err != nil {
		if err == io.EOF {
			return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("the maximum size of the icon is 1MB")
		}
		return err
	}
	// check the content type
	contentType := http.DetectContentType(data)
	switch contentType {
	case GIF, PNG, JPEG:
	default:
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("unsupported content type: %s", contentType)
	}
	artifact.Icon = iconDigest
	return nil
}
