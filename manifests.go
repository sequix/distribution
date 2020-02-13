package distribution

import (
	"context"
	"fmt"
	"mime"

	"github.com/opencontainers/go-digest"
)

// (sequix) Manifest 需要实现的接口
// Manifest represents a registry object specifying a set of
// references and an optional target
type Manifest interface {
	// 返回构成该Manifest的元素，元素是任何可以用Descriptor表示的类型；
	// Descriptor可以用来描述nbd url；
	// Descriptord的顺序没有要求，但最好是有意义的顺序，比如从base layer到top layer;

	// References returns a list of objects which make up this manifest.
	// A reference is anything which can be represented by a
	// distribution.Descriptor. These can consist of layers, resources or other
	// manifests.
	//
	// While no particular order is required, implementations should return
	// them from highest to lowest priority. For example, one might want to
	// return the base layer before the top layer.
	References() []Descriptor

	// 返回manifest的mediaType和相应其bytes表示
	// Payload provides the serialized format of the manifest, in addition to
	// the media type.
	Payload() (mediaType string, payload []byte, err error)
}

// 新建Manifest的builder的interface
// ManifestBuilder creates a manifest allowing one to include dependencies.
// Instances can be obtained from a version-specific manifest package.  Manifest
// specific data is passed into the function which creates the builder.
type ManifestBuilder interface {
	// Build creates the manifest from his builder.
	Build(ctx context.Context) (Manifest, error)

	// 返回添加进的layers，要按添加顺序返回
	// References returns a list of objects which have been added to this
	// builder. The dependencies are returned in the order they were added,
	// which should be from base to head.
	References() []Descriptor

	// 添加一个layer
	// AppendReference includes the given object in the manifest after any
	// existing dependencies. If the add fails, such as when adding an
	// unsupported dependency, an error may be returned.
	//
	// The destination of the reference is dependent on the manifest type and
	// the dependency type.
	AppendReference(dependency Describable) error
}

// manifest的client
// ManifestService describes operations on image manifests.
type ManifestService interface {
	// Exists returns true if the manifest exists.
	Exists(ctx context.Context, dgst digest.Digest) (bool, error)

	// Get retrieves the manifest specified by the given digest
	Get(ctx context.Context, dgst digest.Digest, options ...ManifestServiceOption) (Manifest, error)

	// Put creates or updates the given manifest returning the manifest digest
	Put(ctx context.Context, manifest Manifest, options ...ManifestServiceOption) (digest.Digest, error)

	// Delete removes the manifest specified by the given digest. Deleting
	// a manifest that doesn't exist will return ErrManifestNotFound
	Delete(ctx context.Context, dgst digest.Digest) error
}

// manifest client的遍历接口
// ManifestEnumerator enables iterating over manifests
type ManifestEnumerator interface {
	// Enumerate calls ingester for each manifest.
	Enumerate(ctx context.Context, ingester func(digest.Digest) error) error
}

// 可以被添加进manifest的对象要实现这个接口
// Describable is an interface for descriptors
type Describable interface {
	Descriptor() Descriptor
}

// ManifestMediaTypes returns the supported media types for manifests.
func ManifestMediaTypes() (mediaTypes []string) {
	for t := range mappings {
		if t != "" {
			mediaTypes = append(mediaTypes, t)
		}
	}
	return
}

// UnmarshalFunc implements manifest unmarshalling a given MediaType
type UnmarshalFunc func([]byte) (Manifest, Descriptor, error)

// manifest的解析函数注册在这里
var mappings = make(map[string]UnmarshalFunc)

// UnmarshalManifest looks up manifest unmarshal functions based on
// MediaType
func UnmarshalManifest(ctHeader string, p []byte) (Manifest, Descriptor, error) {
	// Need to look up by the actual media type, not the raw contents of
	// the header. Strip semicolons and anything following them.
	var mediaType string
	if ctHeader != "" {
		var err error
		mediaType, _, err = mime.ParseMediaType(ctHeader)
		if err != nil {
			return nil, Descriptor{}, err
		}
	}

	unmarshalFunc, ok := mappings[mediaType]
	if !ok {
		unmarshalFunc, ok = mappings[""]
		if !ok {
			return nil, Descriptor{}, fmt.Errorf("unsupported manifest media type and no default available: %s", mediaType)
		}
	}

	return unmarshalFunc(p)
}

// 注册manifest的函数
// RegisterManifestSchema registers an UnmarshalFunc for a given schema type.  This
// should be called from specific
func RegisterManifestSchema(mediaType string, u UnmarshalFunc) error {
	if _, ok := mappings[mediaType]; ok {
		return fmt.Errorf("manifest media type registration would overwrite existing: %s", mediaType)
	}
	mappings[mediaType] = u
	return nil
}
