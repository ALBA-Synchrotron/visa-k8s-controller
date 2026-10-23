package resource

import (
	"maps"
	"strconv"
	"strings"

	v1 "github.com/ALBA-Synchrotron/visa-k8s-controller/api/v1"
	"github.com/ALBA-Synchrotron/visa-k8s-controller/internal/resolver"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodBuilder struct{}

func buildContainer(name string, spec *v1.ContainerSpec) corev1.Container {
	return corev1.Container{
		Name:            name,
		Image:           spec.FullImageReference,
		Resources:       spec.Resources,
		Command:         spec.Command,
		Args:            spec.Args,
		Ports:           spec.Ports,
		Env:             spec.Env,
		ImagePullPolicy: spec.ImagePullPolicy,
	}
}

func copyMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	maps.Copy(dst, src)
	return dst
}

func (PodBuilder) Build(visaInstance *v1.VISACloudInstance, config *resolver.ResolvedConfig, imagePullSecrets []corev1.LocalObjectReference) *corev1.Pod {

	containers := make([]corev1.Container, 0, len(config.Sidecars)+1)

	labels := copyMap(visaInstance.Labels)
	labels["cloudprovider.visa.k8s/managed"] = "true"

	securityGroupsNames := make([]string, 0, len(visaInstance.Spec.SecurityGroups))

	for _, sg := range config.SecurityGroups {
		securityGroupsNames = append(securityGroupsNames, sg.Name)
	}

	annotations := copyMap(visaInstance.Annotations)
	annotations["cloudprovider.visa.k8s/security-groups"] = strings.Join(securityGroupsNames, ",")
	annotations["cloudprovider.visa.k8s/home-template-path"] = config.Image.Spec.HomeTemplatePath
	annotations["cloudprovider.visa.k8s/user-shell"] = config.Image.Spec.UserShell
	annotations["cloudprovider.visa.k8s/reboot-generation"] = strconv.FormatInt(visaInstance.Status.ObservedRebootGeneration, 10)

	containers = append(containers, buildContainer("primary", &config.Image.Spec.ContainerSpec))

	for _, sidecar := range config.Sidecars {
		containers = append(containers, buildContainer(sidecar.Name, &sidecar.Spec.ContainerSpec))
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        visaInstance.Name + "-runtime",
			Namespace:   visaInstance.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			Containers:       containers,
			DNSPolicy:        config.Image.Spec.DNSPolicy,
			ImagePullSecrets: imagePullSecrets,
		},
	}

	annotations["cloudprovider.visa.k8s/spec-hash"] = HashVISACloudInstanceSpec(visaInstance.Spec)

	return pod
}
