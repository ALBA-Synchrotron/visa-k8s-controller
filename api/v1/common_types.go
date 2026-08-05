package v1

import corev1 "k8s.io/api/core/v1"

type SidecarConfigRef struct {
	// +kubebuilder:validation:MinLength=1
	// +required
	Name string `json:"name"`
}

type ContainerSpec struct {
	// +required
	// +kubebuilder:validation:MinLength=1
	FullImageReference string `json:"fullImageReference"`
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	// +optional
	Ports []corev1.ContainerPort `json:"ports,omitempty"`
	// +optional
	Command []string `json:"command,omitempty"`
	// +optional
	Args []string `json:"args,omitempty"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	// +kubebuilder:validation:Enum:=Always;Never;IfNotPresent
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}
