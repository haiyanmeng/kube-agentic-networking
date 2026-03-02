/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package conformance

import (
	"io/fs"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/gateway-api/conformance"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func TestGatewayAPIConformance(t *testing.T) {
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("Error loading Kubernetes config: %v", err)
	}

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("Error adding client-go scheme: %v", err)
	}
	if err := apiextensionsv1.AddToScheme(scheme); err != nil {
		t.Fatalf("Error adding apiextensionsv1 (CRD) to scheme: %v", err)
	}
	if err := gatewayv1.AddToScheme(scheme); err != nil {
		t.Fatalf("Error adding gatewayv1 to scheme: %v", err)
	}
	if err := gatewayv1alpha2.AddToScheme(scheme); err != nil {
		t.Fatalf("Error adding gatewayv1alpha2 to scheme: %v", err)
	}
	if err := gatewayv1beta1.AddToScheme(scheme); err != nil {
		t.Fatalf("Error adding gatewayv1beta1 to scheme: %v", err)
	}

	_ = gatewayv1.AddToScheme(clientgoscheme.Scheme)
	_ = gatewayv1alpha2.AddToScheme(clientgoscheme.Scheme)
	_ = gatewayv1beta1.AddToScheme(clientgoscheme.Scheme)

	client, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		t.Fatalf("Error initializing Kubernetes client: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Error initializing Kubernetes clientset: %v", err)
	}

	cSuite, err := suite.NewConformanceTestSuite(suite.ConformanceOptions{
		Client:               client,
		RestConfig:           cfg,
		Clientset:            clientset,
		GatewayClassName:     "kube-agentic-networking",
		Debug:                true,
		CleanupBaseResources: true,
		ManifestFS:           []fs.FS{os.DirFS("."), conformance.Manifests},
		SupportedFeatures: sets.New(
			features.SupportGateway,
			features.SupportHTTPRoute,
			features.SupportBackendTLSPolicy,
		),
		ExemptFeatures: sets.New[features.FeatureName](
			features.SupportReferenceGrant,
		),
		SkipTests: []string{
			// "BackendTLSPolicy",
			// "BackendTLSPolicyConflictResolution",
			// "BackendTLSPolicyInvalidCACertificateRef",
			"BackendTLSPolicyInvalidKind",
			"BackendTLSPolicyObservedGenerationBump",
			"BackendTLSPolicySANValidation",
			"GatewayClassObservedGenerationBump",
			"GatewayHTTPListenerIsolation",
			"GatewayInfrastructure",
			"GatewayInvalidRouteKind",
			"GatewayInvalidTLSConfiguration",
			"GatewayModifyListeners",
			"GatewayObservedGenerationBump",
			"GatewayOptionalAddressValue",
			"GatewaySecretInvalidReferenceGrant",
			"GatewaySecretMissingReferenceGrant",
			"GatewaySecretReferenceGrantAllInNamespace",
			"GatewaySecretReferenceGrantSpecific",
			"GatewayStaticAddresses",
			"GatewayWithAttachedRoutes",
			"GatewayWithAttachedRoutesWithPort8080",
			"GRPCExactMethodMatching",
			"GRPCRouteHeaderMatching",
			"GRPCRouteListenerHostnameMatching",
			"GRPCRouteNamedRule",
			"GRPCRouteWeight",
			"HTTPRouteBackendProtocolH2C",
			"HTTPRouteBackendProtocolWebSocket",
			"HTTPRouteBackendRequestHeaderModifier",
			"HTTPRouteCORSAllowCredentialsBehavior",
			"HTTPRouteCrossNamespace",
			"HTTPRouteDisallowedKind",
			"HTTPRouteExactPathMatching",
			"HTTPRouteHeaderMatching",
			"HTTPRouteHostnameIntersection",
			"HTTPRouteHTTPSListener",
			"HTTPRouteInvalidBackendRefUnknownKind",
			"HTTPRouteInvalidCrossNamespaceBackendRef",
			"HTTPRouteInvalidCrossNamespaceParentRef",
			"HTTPRouteInvalidNonExistentBackendRef",
			"HTTPRouteInvalidParentRefNotMatchingListenerPort",
			"HTTPRouteInvalidParentRefNotMatchingSectionName",
			"HTTPRouteInvalidParentRefSectionNameNotMatchingPort",
			"HTTPRouteInvalidReferenceGrant",
			"HTTPRouteListenerHostnameMatching",
			"HTTPRouteListenerPortMatching",
			"HTTPRouteMatching",
			"HTTPRouteMatchingAcrossRoutes",
			"HTTPRouteMethodMatching",
			"HTTPRouteNamedRule",
			"HTTPRouteObservedGenerationBump",
			"HTTPRoutePartiallyInvalidViaInvalidReferenceGrant",
			"HTTPRoutePathMatchOrder",
			"HTTPRouteQueryParamMatching",
			"HTTPRouteRedirectHostAndStatus",
			"HTTPRouteRedirectPath",
			"HTTPRouteRedirectPort",
			"HTTPRouteRedirectPortAndScheme",
			"HTTPRouteRedirectScheme",
			"HTTPRouteReferenceGrant",
			"HTTPRouteRequestHeaderModifier",
			"HTTPRouteRequestHeaderModifierBackendWeights",
			"HTTPRouteRequestMirror",
			"HTTPRouteRequestMultipleMirrors",
			"HTTPRouteRequestPercentageMirror",
			"HTTPRouteResponseHeaderModifier",
			"HTTPRouteRewriteHost",
			"HTTPRouteRewritePath",
			"HTTPRouteServiceTypes",
			"HTTPRouteSimpleSameNamespace",
			"HTTPRouteTimeoutBackendRequest",
			"HTTPRouteTimeoutRequest",
			"HTTPRouteWeight",
			"MeshBasic",
			"MeshConsumerRoute",
			"MeshFrontend",
			"MeshFrontendHostname",
			"MeshGRPCRouteWeight",
			"MeshHTTPRouteBackendRequestHeaderModifier",
			"MeshHTTPRouteMatching",
			"MeshHTTPRouteNamedRule",
			"MeshHTTPRouteQueryParamMatching",
			"MeshHTTPRouteRedirectHostAndStatus",
			"MeshHTTPRouteRedirectPath",
			"MeshHTTPRouteRedirectPort",
			"MeshHTTPRouteRequestHeaderModifier",
			"MeshHTTPRouteRewritePath",
			"MeshHTTPRouteSchemeRedirect",
			"MeshHTTPRouteSimpleSameNamespace",
			"MeshHTTPRouteWeight",
			"MeshPorts",
			"MeshTrafficSplit",
			"TLSRouteInvalidReferenceGrant",
			"TLSRouteSimpleSameNamespace",
			"UDPRoute",
		},
	})
	if err != nil {
		t.Fatalf("Error creating conformance test suite: %v", err)
	}

	cSuite.Setup(t, tests.ConformanceTests)
	if err := cSuite.Run(t, tests.ConformanceTests); err != nil {
		t.Fatalf("Error running conformance tests: %v", err)
	}
}
