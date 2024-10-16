/* OCI image registry compatibility tests
 *
 * Copyright (c) Siemens AG, 2024
 *
 * Authors:
 *  Tobias Schaffner <tobias.schaffner@siemens.com>
 *  Silvano Cirujano Cuesta <silvano.cirujano-cuesta@siemens.com>
 */

package main_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	v1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/ref"
)

func indexPutOCI(client *regclient.RegClient, ref ref.Ref, m v1.Index) (manifest.Manifest, error) {
	ctx := context.Background()

	index, err := manifest.New(manifest.WithOrig(m))
	if err != nil {
		return index, err
	}

	return index, client.ManifestPut(ctx, ref, index)
}

func getTestIndex() v1.Index {
	m := v1.Index{
		Manifests: []descriptor.Descriptor{},
	}

	m.SchemaVersion = 2

	return m
}

func getTestManifests(t *testing.T) []descriptor.Descriptor {
	// Push file and config
	checkError(t, blobPut(client, reference, "test-data/demo-file.txt"))
	checkError(t, blobPut(client, reference, "test-data/demo-config.txt"))

	// Create manifest
	m := getTestManifest()
	m.MediaType = mediatype.OCI1Manifest

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(m, "", "    "))))
	manifest, err := manifestPutOCI(client, reference, m)
	checkError(t, err)
	manifests := []descriptor.Descriptor{
		{
			MediaType: mediatype.OCI1Manifest,
			Digest:    manifest.GetDescriptor().Digest,
		},
	}
	return manifests
}

// OCI Image Specification - Index -> https://github.com/opencontainers/image-spec/blob/v1.1.0/image-index.md
// Specification says:
// mediaType [...] This property SHOULD be used [...]
// Therefore not specifying this property MUST be supported.
func testNoIndexMediaType(t *testing.T) {
	// Create manifest to refer from index
	manifests := getTestManifests(t)

	// Create index
	i := getTestIndex()
	i.Manifests = manifests

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(i, "", "    "))))
	_, err := indexPutOCI(client, reference, i)
	checkError(t, err)

    // fetch index and manifest
}

// OCI Image Specification - Index -> https://github.com/opencontainers/image-spec/blob/v1.1.0/image-index.md
// Specification says:
// mediaType [...] when used, this field MUST contain [...] application/vnd.oci.image.index.v1+json [...]
func testDefaultIndexMediaType(t *testing.T) {
	// Create manifest to refer from index
	manifests := getTestManifests(t)

	// Create index
	i := getTestIndex()
	i.MediaType = mediatype.OCI1ManifestList
	i.Manifests = manifests

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(i, "", "    "))))
	_, err := indexPutOCI(client, reference, i)
	checkError(t, err)
}

// OCI Image Specification - Index -> https://github.com/opencontainers/image-spec/blob/v1.1.0/image-index.md
// Specification says:
// artifactType [...] MUST comply with RFC 6838
func testIndexArtifactType(t *testing.T) {
	// Create manifest to refer from index
	manifests := getTestManifests(t)

	// Create index
	i := getTestIndex()
	i.MediaType = mediatype.OCI1ManifestList
	i.ArtifactType = "application/my-artifact"
	i.Manifests = manifests

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(i, "", "    "))))
	_, err := indexPutOCI(client, reference, i)
	checkError(t, err)
}

// OCI Image Specification - Index -> https://github.com/opencontainers/image-spec/blob/v1.1.0/image-index.md
// Specification says:
// mediaType [...] when used, this field MUST contain [...] application/vnd.oci.image.index.v1+json [...]
func testWrongIndexMediaTypeFails(t *testing.T) {
	// Create manifest to refer from index
	manifests := getTestManifests(t)

	// Create index
	i := getTestIndex()
	i.MediaType = "application/wrong.type+json"
	i.Manifests = manifests

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(i, "", "    "))))
	_, err := indexPutOCI(client, reference, i)
	expectError(t, err)
}

// OCI Image Specification - Index -> https://github.com/opencontainers/image-spec/blob/v1.1.0/image-index.md
// Specification says:
// manifests/mediaType SHOULD support [...] media types application/vnd.oci.image.index.v1+json
func testNestedIndexes(t *testing.T) {
	// Create manifest to refer from index
	manifests := getTestManifests(t)

	// Create lower index
	i := getTestIndex()
	i.MediaType = mediatype.OCI1ManifestList
	i.Manifests = manifests

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(i, "", "    "))))
	nestedIndex, err := indexPutOCI(client, reference, i)
	checkError(t, err)

	// Create top index
	i = getTestIndex()
	i.MediaType = mediatype.OCI1ManifestList
	var nestedIndexDesc descriptor.Descriptor
	nestedIndexDesc.Digest = nestedIndex.GetDescriptor().Digest
	nestedIndexDesc.Size = nestedIndex.GetDescriptor().Size
	nestedIndexDesc.MediaType = mediatype.OCI1ManifestList
	manifests = append(i.Manifests, nestedIndexDesc)
	i.Manifests = manifests

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(i, "", "    "))))
	_, err = indexPutOCI(client, reference, i)
	checkError(t, err)
}
