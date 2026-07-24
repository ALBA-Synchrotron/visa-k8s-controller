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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type DeviceType string

const (
	DeviceTypePCIPassthrough DeviceType = "PCI_PASSTHROUGH"
	DeviceTypeVirtualGPU     DeviceType = "VIRTUAL_GPU"
)

type VISACloudDeviceSpec struct {
	// +required
	// +kubebuilder:validation:Enum=PCI_PASSTHROUGH;VIRTUAL_GPU
	Type DeviceType `json:"type"`
	// +required
	// +kubebuilder:validation:MinLength=1
	Identifier string `json:"identifier"`
}

// +kubebuilder:object:root=true

type VISACloudDevice struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// +required
	Spec VISACloudDeviceSpec `json:"spec"`
}

// +kubebuilder:object:root=true

type VISACloudDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []VISACloudDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &VISACloudDevice{}, &VISACloudDeviceList{})
		return nil
	})
}
