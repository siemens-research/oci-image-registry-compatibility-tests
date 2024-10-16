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
	"path/filepath"

	g "github.com/onsi/ginkgo/v2"
	"github.com/regclient/regclient/types/ref"
	"github.com/onsi/ginkgo/v2/reporters"
	. "github.com/onsi/gomega"
)

const (
	suiteDescription = "OCI Feature Tests"
)

func TestConformance(t *testing.T) {
	g.Describe(suiteDescription, func() {
		testNoManifestMediaType()
		testDefaultMediaType()
		testDefaultConfigType()
		testEmptyConfigFileAndArtifactType()
		testArtifactTypeOverConfigType()
		testBlobMediaType()
		testWrongManifestMediaTypeFails()
		testManifestWithSubjectEntry()
		testNoIndexMediaType()
		testDefaultIndexMediaType()
		testIndexArtifactType()
		testWrongIndexMediaTypeFails()
		testNestedIndexes()
	})

    reportJUnitFilename := filepath.Join(".", "junit.xml")
	RegisterFailHandler(g.Fail)
	suiteConfig, reporterConfig := g.GinkgoConfiguration()
	g.ReportAfterSuite("junit custom reporter", func(r g.Report) {
		if reportJUnitFilename != "" {
			_ = reporters.GenerateJUnitReportWithConfig(r, reportJUnitFilename, reporters.JunitReportConfig{
				OmitLeafNodeType: true,
			})
		}
	})

	// Set registry details
	host := *flag.String("host", os.Getenv("REGISTRY_HOST"), "The registry host")
	user := *flag.String("user", os.Getenv("REGISTRY_USER"), "The login user")
	password := *flag.String("password", os.Getenv("REGISTRY_PASSWORD"), "The login password")
	namespace := *flag.String("namespace", os.Getenv("REGISTRY_NAMESPACE"), "The namespace that should be used when pushing")
	var err error

	client = loginToRegistry(host, user, password, !strings.HasPrefix(host, "http://"))

	host = strings.Replace(strings.Replace(host, "http://", "", 1), "https://", "", 1)
	reference, err = ref.New(host + "/" + namespace + ":demo")
	Expect(err).To(BeNil())

	g.RunSpecs(t, "OCI conformance tests", suiteConfig, reporterConfig)
}
