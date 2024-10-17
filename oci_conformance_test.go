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
	"strconv"
	"log"

	g "github.com/onsi/ginkgo/v2"
	"github.com/regclient/regclient/types/ref"
	"github.com/onsi/ginkgo/v2/reporters"
	. "github.com/onsi/gomega"
)

const (
	suiteDescription = "OCI Feature Tests"
    envVarRootURL = "REGISTRY_HOST"
    envVarNamespace = "REGISTRY_NAMESPACE"
    envVarUsername = "REGISTRY_USER"
    envVarPassword = "REGISTRY_PASSWORD"
    envVarDebug = "REGISTRY_DEBUG"
)

var (
	httpWriter                         *httpDebugWriter
    Version = "unknown"
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

	debug, _ := strconv.ParseBool(os.Getenv(envVarDebug))

	httpWriter = newHTTPDebugWriter(debug)

    reportJUnitFilename := filepath.Join(".", "report.xml")
    reportHTMLFilename := filepath.Join(".", "report.html")
	RegisterFailHandler(g.Fail)
	suiteConfig, reporterConfig := g.GinkgoConfiguration()
	hr := newHTMLReporter(reportHTMLFilename)
	g.ReportAfterEach(hr.afterReport)
	g.ReportAfterSuite("html custom reporter", func(r g.Report) {
		if err := hr.endSuite(r); err != nil {
			log.Printf("\nWARNING: cannot write HTML summary report: %v", err)
		}
	})
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
