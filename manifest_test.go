/* OCI image registry compatibility tests
 *
 * Copyright (c) Siemens AG, 2024
 *
 * Authors:
 *  Tobias Schaffner <tobias.schaffner@siemens.com>
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main_test

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	v1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/ref"
)

var client *regclient.RegClient
var reference ref.Ref

func checkError(t *testing.T, err error) {
	if err != nil {
		_, filename, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d %v", filename, line, err)
	}
}

func expectError(t *testing.T, err error) {
	_, filename, line, _ := runtime.Caller(1)
	if err == nil {
		t.Fatalf("%s:%d succeeded unexpectedly!", filename, line)
	} else {
		t.Logf("%s:%d expected error: %v", filename, line, err)
	}
}

func ignoreError[T any](val T, _ error) T {
	return val
}

func loginToRegistry(host string, user string, password string, tls bool) *regclient.RegClient {
	configHost := config.Host{
		Name: host,
		User: user,
		Pass: password,
	}

	if !tls {
		configHost.TLS = config.TLSDisabled
	}

	return regclient.New(regclient.WithConfigHost(configHost))
}

func blobPut(client *regclient.RegClient, ref ref.Ref, path string) error {
	ctx := context.Background()

	raw, err := os.Open(path)
	if err != nil {
		return err
	}

	_, err = client.BlobPut(ctx, ref, descriptor.Descriptor{}, raw)
	return err
}

func manifestPutRaw(client *regclient.RegClient, ref ref.Ref, m []byte) (manifest.Manifest, error) {
	ctx := context.Background()

	manifest, err := manifest.New(manifest.WithRaw(m))
	if err != nil {
		return manifest, err
	}

	return manifest, client.ManifestPut(ctx, ref, manifest)
}

func manifestPutOCI(client *regclient.RegClient, ref ref.Ref, m v1.Manifest) (manifest.Manifest, error) {
	ctx := context.Background()

	manifest, err := manifest.New(manifest.WithOrig(m))
	if err != nil {
		return manifest, err
	}

	return manifest, client.ManifestPut(ctx, ref, manifest)
}

func getTestManifest() v1.Manifest {
	m := v1.Manifest{
		Config: descriptor.Descriptor{
			MediaType: mediatype.OCI1ImageConfig,
			Digest:    ignoreError(digest.Parse("sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a")),
			Size:      2,
		},
		Layers: []descriptor.Descriptor{
			{
				MediaType: mediatype.OCI1LayerGzip,
				Digest:    ignoreError(digest.Parse("sha256:e63246ad2bce533bcfc8cdfcbc936eba500552aa49ff4527204b4c36d99c3e98")),
				Size:      69,
			},
		},
	}

	m.SchemaVersion = 2

	return m
}

func TestMain(t *testing.T) {
	// Set registry details
	host := *flag.String("host", os.Getenv("REGISTRY_HOST"), "The registry host")
	user := *flag.String("user", os.Getenv("REGISTRY_USER"), "The login user")
	password := *flag.String("password", os.Getenv("REGISTRY_PASSWORD"), "The login password")
	namespace := *flag.String("namespace", os.Getenv("REGISTRY_NAMESPACE"), "The namespace that should be used when pushing")
	var err error

	client = loginToRegistry(host, user, password, !strings.HasPrefix(host, "http://"))

	host = strings.Replace(strings.Replace(host, "http://", "", 1), "https://", "", 1)
	reference, err = ref.New(host + "/" + namespace + ":demo")
	checkError(t, err)
}

func TestNoManifestMediaType(t *testing.T) {
	t.Run("Manifest without a `mediaType` is accepted.", testNoManifestMediaType)
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// mediaType [...] This property SHOULD be used [...]
// Therefore not specifying this property MUST be supported.
func testNoManifestMediaType(t *testing.T) {
	// Push artifact and config
	checkError(t, blobPut(client, reference, "test-data/demo-artifact.txt"))
	checkError(t, blobPut(client, reference, "test-data/demo-config.txt"))

	// Create manifest
	m := getTestManifest()

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(m, "", "    "))))
	_, err := manifestPutOCI(client, reference, m)
	checkError(t, err)
}

func TestDefaultMediaType(t *testing.T) {
	t.Run("Manifest with `mediaType` `application/vnd.oci.image.manifest.v1+json` is accepted.", testDefaultMediaType)
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// mediaType [...] when used, this field MUST contain [...] application/vnd.oci.image.manifest.v1+json [...]
func testDefaultMediaType(t *testing.T) {
	// Push artifact and config
	checkError(t, blobPut(client, reference, "test-data/demo-artifact.txt"))
	checkError(t, blobPut(client, reference, "test-data/demo-config.txt"))

	// Create manifest
	m := getTestManifest()
	m.MediaType = mediatype.OCI1Manifest

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(m, "", "    "))))
	_, err := manifestPutOCI(client, reference, m)
	checkError(t, err)
}

func TestDefaultConfigType(t *testing.T) {
	t.Run("Manifest with `config/mediaType` `application/vnd.oci.image.config.v1+json` is accepted.", testDefaultConfigType)
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// config/mediaType [...] Implementations MUST support at least the following media types: application/vnd.oci.image.config.v1+json [...]
func testDefaultConfigType(t *testing.T) {
	// Push artifact and config
	checkError(t, blobPut(client, reference, "test-data/demo-artifact.txt"))
	checkError(t, blobPut(client, reference, "test-data/demo-config.txt"))

	// Create manifest
	m := getTestManifest()
	m.MediaType = mediatype.OCI1Manifest

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(m, "", "    "))))
	_, err := manifestPutOCI(client, reference, m)
	checkError(t, err)
}

func TestEmptyConfigFileAndArtifactType(t *testing.T) {
	t.Run("Manifest with custom `artifactType` is accepted.", testEmptyConfigFileAndArtifactType)
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// artifactType [...] This MUST be set when config.mediaType is set to the empty value [...]
func testEmptyConfigFileAndArtifactType(t *testing.T) {
	// Push artifact and config
	checkError(t, blobPut(client, reference, "test-data/demo-artifact.txt"))
	checkError(t, blobPut(client, reference, "test-data/demo-config.txt"))

	// Create manifest
	m := getTestManifest()
	m.MediaType = mediatype.OCI1Manifest
	m.ArtifactType = "application/my-artifact"
	m.Config.MediaType = mediatype.OCI1Empty

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(m, "", "    "))))
	_, err := manifestPutOCI(client, reference, m)
	checkError(t, err)
}

func TestArtifactTypeOverConfigType(t *testing.T) {
	t.Run("Manifest with custom `config/mediaType`, as artifact type, is accepted.", testArtifactTypeOverConfigType)
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// config/mediaType [...] MUST NOT error on encountering a value that is unknown to the implementation [...]
func testArtifactTypeOverConfigType(t *testing.T) {
	// Push artifact and config
	checkError(t, blobPut(client, reference, "test-data/demo-artifact.txt"))
	checkError(t, blobPut(client, reference, "test-data/demo-config.txt"))

	// Create manifest
	m := getTestManifest()
	m.MediaType = mediatype.OCI1Manifest
	m.Config.MediaType = "application/my-artifact-legacy"

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(m, "", "    "))))
	_, err := manifestPutOCI(client, reference, m)
	checkError(t, err)
}

func TestBlobMediaType(t *testing.T) {
	t.Run("Manifest with custom `blob/mediaType` is accepted.", testBlobMediaType)
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// layers/mediaType [...] MUST NOT error on encountering a mediaType that is unknown to the implementation [...]
func testBlobMediaType(t *testing.T) {
	// Push artifact and config
	checkError(t, blobPut(client, reference, "test-data/demo-artifact.txt"))
	checkError(t, blobPut(client, reference, "test-data/demo-config.txt"))

	// Create manifest
	m := getTestManifest()
	m.MediaType = mediatype.OCI1Manifest
	m.Layers[0].MediaType = "application/my-blob-format"

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(m, "", "    "))))
	_, err := manifestPutOCI(client, reference, m)
	checkError(t, err)
}

func TestWrongManifestMediaTypeFails(t *testing.T) {
	t.Run("Manifest with wrong `mediaType` is rejected.", testWrongManifestMediaTypeFails)
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// mediaType [...] when used, this field MUST contain [...] application/vnd.oci.image.manifest.v1+json [...]
func testWrongManifestMediaTypeFails(t *testing.T) {
	// Push artifact and config
	checkError(t, blobPut(client, reference, "test-data/demo-artifact.txt"))
	checkError(t, blobPut(client, reference, "test-data/demo-config.txt"))

	// Create manifest
	m := getTestManifest()
	m.MediaType = "application/wrong.type+json"

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(m, "", "    "))))
	_, err := manifestPutOCI(client, reference, m)
	expectError(t, err)
}

func TestManifestWithSubjectEntry(t *testing.T) {
	t.Run("Manifest with `subject` property is accepted.", testManifestWithSubjectEntry)
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// subject [...] This OPTIONAL property specifies a descriptor of another manifest [...]
func testManifestWithSubjectEntry(t *testing.T) {
	// Push artifact and config
	checkError(t, blobPut(client, reference, "test-data/demo-artifact.txt"))
	checkError(t, blobPut(client, reference, "test-data/demo-config.txt"))

	// Create first manifest
	m := getTestManifest()
	m.MediaType = mediatype.OCI1Manifest

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(m, "", "    "))))
	first_manifest, err := manifestPutOCI(client, reference, m)
	checkError(t, err)

	// Create second manifest
	m = getTestManifest()
	m.MediaType = mediatype.OCI1Manifest
	var subject descriptor.Descriptor
	subject.Digest = first_manifest.GetDescriptor().Digest
	subject.Size = first_manifest.GetDescriptor().Size
	subject.MediaType = mediatype.OCI1Manifest
	m.Subject = &subject

	// Log and push to registry
	t.Log(string(ignoreError(json.MarshalIndent(m, "", "    "))))
	_, err = manifestPutOCI(client, reference, m)
	checkError(t, err)
}
