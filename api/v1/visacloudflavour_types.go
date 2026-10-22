/*
Copyright 2026.

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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type VISACloudFlavourSpec struct {
	// +optional
	// +listType=map
	// +listMapKey=name
	Sidecars []v1.LocalObjectReference `json:"sidecars,omitempty"`
	// +kubebuilder:validation:Required
	Limits v1.ResourceList `json:"limits"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[1-9][0-9]*(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$`
	DiskSize string `json:"diskSize"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
type VISACloudFlavour struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec VISACloudFlavourSpec `json:"spec"`
}

// +kubebuilder:object:root=true
type VISACloudFlavourList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VISACloudFlavour `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &VISACloudFlavour{}, &VISACloudFlavourList{})
		return nil
	})
}
