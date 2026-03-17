// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package conformance

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/gateway-api/conformance"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func conformanceOpts(t *testing.T) suite.ConformanceOptions {
	gatewayNamespaceMode := false
	internalSuite := EnvoyGatewaySuite(gatewayNamespaceMode, true)

	opts := conformance.DefaultOptions(t)
	opts.SkipTests = internalSuite.SkipTests
	opts.SupportedFeatures = internalSuite.SupportedFeatures
	opts.ExemptFeatures = internalSuite.ExemptFeatures

	timeoutConfig := config.DefaultTimeoutConfig()
	timeoutConfig.GatewayMustHaveCondition = 120 * time.Second
	timeoutConfig.GatewayStatusMustHaveListeners = 120 * time.Second
	timeoutConfig.GatewayListenersMustHaveConditions = 120 * time.Second
	timeoutConfig.RouteMustHaveParents = 120 * time.Second
	timeoutConfig.MaxTimeToConsistency = 120 * time.Second
	opts.TimeoutConfig = timeoutConfig

	opts.FailFast = true

	return opts
}

// SkipTests is a list of tests that are skipped in the conformance suite.
func SkipTests(gatewayNamespaceMode bool) []suite.ConformanceTest {
	skipTests := []suite.ConformanceTest{
		// Requires TLS support (not yet implemented).
		tests.GatewayInvalidTLSConfiguration,
		// Requires TLS support (not yet implemented).
		tests.HTTPRouteHTTPSListener,
		// Requires EDS to handle headless services without selector.
		tests.HTTPRouteServiceTypes,
	}

	if gatewayNamespaceMode {
		return skipTests
	}

	skipTests = append(skipTests, tests.GatewayInfrastructure)

	return skipTests
}

func skipTestsShortNames(skipTests []suite.ConformanceTest) []string {
	shortNames := make([]string, len(skipTests))
	for i, test := range skipTests {
		shortNames[i] = test.ShortName
	}
	return shortNames
}

// EnvoyGatewaySuite is the conformance suite configuration for the Gateway API.
func EnvoyGatewaySuite(gatewayNamespaceMode, standardChannel bool) suite.ConformanceOptions {
	return suite.ConformanceOptions{
		SupportedFeatures: sets.New[features.FeatureName](
			features.SupportGateway,
			features.SupportReferenceGrant,
			features.SupportHTTPRoute,
		),
		ExemptFeatures:    meshFeatures(),
		SkipTests:         skipTestsShortNames(SkipTests(gatewayNamespaceMode)),
	}
}

func meshFeatures() sets.Set[features.FeatureName] {
	result := sets.New[features.FeatureName]()
	for _, feature := range features.MeshCoreFeatures.UnsortedList() {
		result.Insert(feature.Name)
	}
	for _, feature := range features.MeshExtendedFeatures.UnsortedList() {
		result.Insert(feature.Name)
	}
	return result
}
