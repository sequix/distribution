package piddle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/opencontainers/go-digest"
	"github.com/sequix/distribution"
	"github.com/sequix/distribution/manifest"
)

const (
	// 请求 Manifest List（Image Index） 时带的Accept；返回时带的Content-Type
	MediaTypeImageIndex = "application/vnd.piddle.image.index.v1+json"

	// Image Manifest 中的config，用以索引image config blob
	MediaTypeImageConfig = "application/vnd.piddle.image.manifest.v1+json"

	// Image Manifest 中的layers，用以索引layer.tar blob
	MediaTypeLayer = "application/vnd.piddle.image.layer.v1.tar"

	// Image Manifest 中的layers，用以索引layer.tar.gz blob
	MediaTypeLayerGzip = "application/vnd.piddle.image.layer.v1.tar+gzip"

	// Image Manifest 中的layers，用以索引layer.tar.zstd blob
	MediaTypeLayerZstd = "application/vnd.piddle.image.layer.v1.tar+zstd"

	// Image Manifest 中的layers，用以索引layer.tar blob，且不可被push到registry
	MediaTypeNondistributableLayer = "application/vnd.piddle.image.layer.nondistributable.v1.tar"

	// Image Manifest 中的layers，用以索引layer.tar blob，且不可被push到registry
	MediaTypeNondistributableLayerGzip = "application/vnd.piddle.image.layer.nondistributable.v1.tar+gzip"

	// Image Manifest 中的layers，用以索引layer.tar blob，且不可被push到registry
	MediaTypeNondistributableLayerZstd = "application/vnd.piddle.image.layer.nondistributable.v1.tar+zstd"
)

var (
	// SchemaVersion provides a pre-initialized version structure for this
	// packages version of the manifest.
	SchemaVersion = manifest.Versioned{
		SchemaVersion: 1,
		MediaType:     MediaTypeImageIndex,
	}
)

// 注册新的manifest
func init() {
	schema2Func := func(b []byte) (distribution.Manifest, distribution.Descriptor, error) {
		m := new(DeserializedManifest)
		err := m.UnmarshalJSON(b)
		if err != nil {
			return nil, distribution.Descriptor{}, err
		}

		dgst := digest.FromBytes(b)
		return m, distribution.Descriptor{Digest: dgst, Size: int64(len(b)), MediaType: MediaTypeImageIndex}, err
	}
	err := distribution.RegisterManifestSchema(MediaTypeImageIndex, schema2Func)
	if err != nil {
		panic(fmt.Sprintf("Unable to register manifest: %s", err))
	}
}

// Manifest defines a piddle manifest.
type Manifest struct {
	manifest.Versioned

	// image manifest 索引的 blob 需要下述数据
	// 1.表示layer所在的nbd-url
	// 2.image运行的platform
	// 3.mediaType，用于自检和兼容
	// 4.annotation，KV对，任意的附加信息

	// Config references the image configuration as a blob.
	Config distribution.Descriptor `json:"config"`

	// Layers lists descriptors for the layers referenced by the
	// configuration.
	Layers []distribution.Descriptor `json:"layers"`
}

// References returns the descriptors of this manifests references.
func (m Manifest) References() []distribution.Descriptor {
	references := make([]distribution.Descriptor, 0, 1+len(m.Layers))
	references = append(references, m.Config)
	references = append(references, m.Layers...)
	return references
}

// Target returns the target of this manifest.
func (m Manifest) Target() distribution.Descriptor {
	return m.Config
}

// DeserializedManifest wraps Manifest with a copy of the original JSON.
// It satisfies the distribution.Manifest interface.
type DeserializedManifest struct {
	Manifest

	// canonical is the canonical byte representation of the Manifest.
	canonical []byte
}

// FromStruct takes a Manifest structure, marshals it to JSON, and returns a
// DeserializedManifest which contains the manifest and its JSON representation.
func FromStruct(m Manifest) (*DeserializedManifest, error) {
	var deserialized DeserializedManifest
	deserialized.Manifest = m

	var err error
	deserialized.canonical, err = json.MarshalIndent(&m, "", "   ")
	return &deserialized, err
}

// UnmarshalJSON populates a new Manifest struct from JSON data.
func (m *DeserializedManifest) UnmarshalJSON(b []byte) error {
	m.canonical = make([]byte, len(b))
	// store manifest in canonical
	copy(m.canonical, b)

	// Unmarshal canonical JSON into Manifest object
	var manifest Manifest
	if err := json.Unmarshal(m.canonical, &manifest); err != nil {
		return err
	}

	if manifest.MediaType != MediaTypeImageIndex {
		return fmt.Errorf("mediaType in manifest should be '%s' not '%s'",
			MediaTypeImageIndex, manifest.MediaType)

	}

	m.Manifest = manifest

	return nil
}

// MarshalJSON returns the contents of canonical. If canonical is empty,
// marshals the inner contents.
func (m *DeserializedManifest) MarshalJSON() ([]byte, error) {
	if len(m.canonical) > 0 {
		return m.canonical, nil
	}

	return nil, errors.New("JSON representation not initialized in DeserializedManifest")
}

// Payload returns the raw content of the manifest. The contents can be used to
// calculate the content identifier.
func (m DeserializedManifest) Payload() (string, []byte, error) {
	return m.MediaType, m.canonical, nil
}
