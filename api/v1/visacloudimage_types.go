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

type VISACloudImageSpec struct {
	// required
	ContainerSpec `json:"containerSpec"`
	// +required
	// +kubebuilder:validation:MinLength=1
	HomeTemplatePath string `json:"homeTemplatePath"`
	// +required
	// +kubebuilder:validation:MinLength=1
	UserShell string `json:"userShell"`
	// +required
	// +kubebuilder:validation:MinLength=1
	Size string `json:"size"`
	// +optional
	// +kubebuilder:validation:Enum:=ClusterFirstWithHostNet;ClusterFirst;Default
	v1.DNSPolicy `json:"dnsPolicy,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
type VISACloudImage struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +required
	Spec VISACloudImageSpec `json:"spec"`
}

// +kubebuilder:object:root=true
type VISACloudImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VISACloudImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &VISACloudImage{}, &VISACloudImageList{})
		return nil
	})
}
