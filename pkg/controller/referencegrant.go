/*
Copyright The Kubernetes Authors.

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

package controller

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	gatewayinformersv1beta1 "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions/apis/v1beta1"
)

func (c *Controller) setupReferenceGrantEventHandlers(referenceGrantInformer gatewayinformersv1beta1.ReferenceGrantInformer) error {
	_, err := referenceGrantInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onReferenceGrantAdd,
		UpdateFunc: c.onReferenceGrantUpdate,
		DeleteFunc: c.onReferenceGrantDelete,
	})
	return err
}

func (c *Controller) onReferenceGrantAdd(obj interface{}) {
	rg := obj.(*gatewayv1beta1.ReferenceGrant)
	klog.V(4).InfoS("Adding ReferenceGrant", "referencegrant", klog.KObj(rg))
	c.enqueueGatewaysForReferenceGrant(rg)
}

func (c *Controller) onReferenceGrantUpdate(old, newObj interface{}) {
	oldRG := old.(*gatewayv1beta1.ReferenceGrant)
	newRG := newObj.(*gatewayv1beta1.ReferenceGrant)
	if !reflect.DeepEqual(oldRG.Spec, newRG.Spec) {
		klog.V(4).InfoS("Updating ReferenceGrant", "referencegrant", klog.KObj(newRG))
		c.enqueueGatewaysForReferenceGrant(newRG)
	}
}

func (c *Controller) onReferenceGrantDelete(obj interface{}) {
	rg, ok := obj.(*gatewayv1beta1.ReferenceGrant)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		rg, ok = tombstone.Obj.(*gatewayv1beta1.ReferenceGrant)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a ReferenceGrant %#v", obj))
			return
		}
	}
	klog.V(4).InfoS("Deleting ReferenceGrant", "referencegrant", klog.KObj(rg))
	c.enqueueGatewaysForReferenceGrant(rg)
}

func (c *Controller) enqueueGatewaysForReferenceGrant(rg *gatewayv1beta1.ReferenceGrant) {
	routes, err := c.gateway.httprouteLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to list HTTPRoutes: %v", err)
		return
	}

	gatewaysToEnqueue := make(map[string]struct{})

	for _, route := range routes {
		for _, rule := range route.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				ns := route.Namespace
				if backendRef.Namespace != nil {
					ns = string(*backendRef.Namespace)
				}
				if ns != rg.Namespace {
					continue
				}

				// If the route references a service in the namespace of the ReferenceGrant,
				// enqueue all Gateways referenced by this route.
				for _, ref := range route.Spec.ParentRefs {
					if (ref.Group != nil && string(*ref.Group) != gatewayv1.GroupName) ||
						(ref.Kind != nil && string(*ref.Kind) != "Gateway") {
						continue
					}
					namespace := route.Namespace
					if ref.Namespace != nil {
						namespace = string(*ref.Namespace)
					}
					key := namespace + "/" + string(ref.Name)
					gatewaysToEnqueue[key] = struct{}{}
				}
			}
		}
	}

	// Also find Gateways that reference Secrets in the namespace of the ReferenceGrant.
	gws, err := c.gateway.gatewayLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to list Gateways: %v", err)
		return
	}

	for _, gw := range gws {
		for _, listener := range gw.Spec.Listeners {
			if listener.TLS != nil {
				for _, ref := range listener.TLS.CertificateRefs {
					ns := gw.Namespace
					if ref.Namespace != nil {
						ns = string(*ref.Namespace)
					}
					if ns == rg.Namespace {
						key := gw.Namespace + "/" + gw.Name
						gatewaysToEnqueue[key] = struct{}{}
					}
				}
			}
		}
	}

	for key := range gatewaysToEnqueue {
		c.gatewayqueue.Add(key)
	}
}
