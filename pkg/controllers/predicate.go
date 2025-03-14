/*
Copyright 2021 The OpenYurt Authors.

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

package controllers

import (
	devicev1alpha1 "github.com/openyurtio/device-controller/apis/device.openyurt.io/v1alpha1"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func genFirstUpdateFilter(objKind string) predicate.Predicate {
	return predicate.Funcs{
		// ignore the update event that is generated due to a
		// new deviceprofile being added to the Edgex Foundry
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldDp, ok := e.ObjectOld.(devicev1alpha1.EdgeXObject)
			if !ok {
				klog.Infof("fail to assert object to deviceprofile, object kind is %s", objKind)
				return false
			}
			newDp, ok := e.ObjectNew.(devicev1alpha1.EdgeXObject)
			if !ok {
				klog.Infof("fail to assert object to deviceprofile, object kind is %s", objKind)
				return false
			}
			if oldDp.IsAddedToEdgeX() == false &&
				newDp.IsAddedToEdgeX() == true {
				return false
			}
			return true
		},
	}
}
