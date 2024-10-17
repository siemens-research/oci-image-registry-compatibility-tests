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

	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	v1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/ref"
)

var (
	titleIndex = "OCI Manifest Index (Image List)"
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

func getTestManifests() []descriptor.Descriptor {
	// Push file and config
	Expect(blobPut(client, reference, "test-data/demo-file.txt")).To(BeNil())
	Expect(blobPut(client, reference, "test-data/demo-config.txt")).To(BeNil())

	// Create manifest
	m := getTestManifest()
	m.MediaType = mediatype.OCI1Manifest

	// Log and push to registry
	g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(m, "", "    "))))
	manifest, err := manifestPutOCI(client, reference, m)
	Expect(err).To(BeNil())
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
var testNoIndexMediaType = func() {
	g.Context(titleIndex, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(client, reference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(client, reference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push index without any mediaType", func() {
				// Create manifest to refer from index
				manifests := getTestManifests()

				// Create index
				i := getTestIndex()
				i.Manifests = manifests

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(i, "", "    "))))
				_, err := indexPutOCI(client, reference, i)
				Expect(err).To(BeNil())
			})
		})
	})
}

// OCI Image Specification - Index -> https://github.com/opencontainers/image-spec/blob/v1.1.0/image-index.md
// Specification says:
// mediaType [...] when used, this field MUST contain [...] application/vnd.oci.image.index.v1+json [...]
var testDefaultIndexMediaType = func() {
	g.Context(titleIndex, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(client, reference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(client, reference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push index with default mediaType (application/vnd.oci.image.index.v1+json)", func() {
				// Create manifest to refer from index
				manifests := getTestManifests()

				// Create index
				i := getTestIndex()
				i.MediaType = mediatype.OCI1ManifestList
				i.Manifests = manifests

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(i, "", "    "))))
				_, err := indexPutOCI(client, reference, i)
				Expect(err).To(BeNil())
			})
		})
	})
}

// OCI Image Specification - Index -> https://github.com/opencontainers/image-spec/blob/v1.1.0/image-index.md
// Specification says:
// artifactType [...] MUST comply with RFC 6838
var testIndexArtifactType = func() {
	g.Context(titleIndex, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(client, reference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(client, reference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push index with custom artifactType", func() {
				// Create manifest to refer from index
				manifests := getTestManifests()

				// Create index
				i := getTestIndex()
				i.MediaType = mediatype.OCI1ManifestList
				i.ArtifactType = "application/my-artifact"
				i.Manifests = manifests

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(i, "", "    "))))
				_, err := indexPutOCI(client, reference, i)
				Expect(err).To(BeNil())
			})
		})
	})
}

// OCI Image Specification - Index -> https://github.com/opencontainers/image-spec/blob/v1.1.0/image-index.md
// Specification says:
// mediaType [...] when used, this field MUST contain [...] application/vnd.oci.image.index.v1+json [...]
var testWrongIndexMediaTypeFails = func() {
	g.Context(titleIndex, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(client, reference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(client, reference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push index with invalid mediaType", func() {
				// Create manifest to refer from index
				manifests := getTestManifests()

				// Create index
				i := getTestIndex()
				i.MediaType = "application/wrong.type+json"
				i.Manifests = manifests

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(i, "", "    "))))
				_, err := indexPutOCI(client, reference, i)
				Expect(err).To(MatchError("manifest contains an unexpected media type: expected application/vnd.oci.image.index.v1+json, received application/wrong.type+json"))
			})
		})
	})
}

// OCI Image Specification - Index -> https://github.com/opencontainers/image-spec/blob/v1.1.0/image-index.md
// Specification says:
// manifests/mediaType SHOULD support [...] media types application/vnd.oci.image.index.v1+json
var testNestedIndexes = func() {
	g.Context(titleIndex, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(client, reference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(client, reference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push manifest with invalid mediaType", func() {
				// Create manifest to refer from index
				manifests := getTestManifests()

				// Create lower index
				i := getTestIndex()
				i.MediaType = mediatype.OCI1ManifestList
				i.Manifests = manifests

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(i, "", "    "))))
				nestedIndex, err := indexPutOCI(client, reference, i)
				Expect(err).To(BeNil())

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
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(i, "", "    "))))
				_, err = indexPutOCI(client, reference, i)
				Expect(err).To(BeNil())
			})
		})
	})
}
