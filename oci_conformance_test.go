/* OCI image registry compatibility tests
 *
 * Copyright (c) Siemens AG, 2024
 *
 * Authors:
 *  Tobias Schaffner <tobias.schaffner@siemens.com>
 *  Silvano Cirujano Cuesta <silvano.cirujano-cuesta@siemens.com>
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main_test

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/regclient/regclient/types/ref"
)

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

func TestDefaultMediaType(t *testing.T) {
	t.Run("Manifest with `mediaType` `application/vnd.oci.image.manifest.v1+json` is accepted.", testDefaultMediaType)
}

func TestDefaultConfigType(t *testing.T) {
	t.Run("Manifest with `config/mediaType` `application/vnd.oci.image.config.v1+json` is accepted.", testDefaultConfigType)
}

func TestEmptyConfigFileAndArtifactType(t *testing.T) {
	t.Run("Manifest with custom `artifactType` is accepted.", testEmptyConfigFileAndArtifactType)
}

func TestArtifactTypeOverConfigType(t *testing.T) {
	t.Run("Manifest with custom `config/mediaType`, as artifact type, is accepted.", testArtifactTypeOverConfigType)
}

func TestBlobMediaType(t *testing.T) {
	t.Run("Manifest with custom `blob/mediaType` is accepted.", testBlobMediaType)
}

func TestWrongManifestMediaTypeFails(t *testing.T) {
	t.Run("Manifest with wrong `mediaType` is rejected.", testWrongManifestMediaTypeFails)
}

func TestManifestWithSubjectEntry(t *testing.T) {
	t.Run("Manifest with `subject` property is accepted.", testManifestWithSubjectEntry)
}

func TestNoIndexMediaType(t *testing.T) {
	t.Run("Index without mediaType is accepted.", testNoIndexMediaType)
}

func TestDefaultIndexMediaType(t *testing.T) {
	t.Run("Index with `mediaType` `application/vnd.oci.image.index.v1+json` is accepted.", testDefaultIndexMediaType)
}

func TestIndexArtifactType(t *testing.T) {
	t.Run("Index with custom `artifactType` is accepted.", testIndexArtifactType)
}

func TestWrongIndexMediaTypeFails(t *testing.T) {
	t.Run("Index with wrong `mediaType` is rejected.", testWrongIndexMediaTypeFails)
}

func TestNestedIndexes(t *testing.T) {
	t.Run("Indexes referring other indexes are accepted.", testNestedIndexes)
}
