// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build conformance

package conformance

import (
	"flag"
	"os"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func TestGatewayAPIConformance(t *testing.T) {
	flag.Parse()
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))

	if flags.RunTest != nil && *flags.RunTest != "" {
		tlog.Logf(t, "Running Conformance test %s with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.RunTest, *flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	} else {
		tlog.Logf(t, "Running Conformance tests with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	}

	opts := conformanceOpts(t)
	opts.RunTest = *flags.RunTest

	// If focusing on a single test, clear the skip list to ensure it runs.
	if opts.RunTest != "" {
		opts.SkipTests = nil
	}

	cSuite, err := suite.NewConformanceTestSuite(opts)
	if err != nil {
		t.Fatalf("Error creating conformance test suite: %v", err)
	}

	var coreTests []suite.ConformanceTest
	for _, test := range tests.ConformanceTests {
		supported := true
		for _, feature := range test.Features {
			if feature != features.SupportGateway &&
				feature != features.SupportReferenceGrant &&
				feature != features.SupportHTTPRoute {
				supported = false
				break
			}
		}
		if supported {
			coreTests = append(coreTests, test)
		}
	}

	cSuite.Setup(t, coreTests)
	if err := cSuite.Run(t, coreTests); err != nil {
		t.Fatalf("Error running conformance tests: %v", err)
	}
}
