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
	"runtime"
	"testing"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
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
