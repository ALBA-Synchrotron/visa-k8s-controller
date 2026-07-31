package resource

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	v1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	"git.cells.es/mis/visa/visa-k8s-controller/internal/resolver"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

type PodBuilder struct{}

func buildContainer(name string, spec *v1.ContainerSpec) corev1.Container {
	return corev1.Container{
		Name:      name,
		Image:     spec.FullImageReference,
		Resources: spec.Resources,
		Command:   spec.Command,
		Args:      spec.Args,
		Ports:     spec.Ports,
		Env:       spec.Env,
	}
}

func copyMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (PodBuilder) Build(visaInstance *v1.VISAInstance, config *resolver.ResolvedConfig) *corev1.Pod {

	containers := make([]corev1.Container, 0, len(config.Sidecars)+1)

	labels := copyMap(visaInstance.Labels)
	labels["visa.k8s.cells.es/managed"] = "true"

	securityGroupsNames := make([]string, 0, len(visaInstance.Spec.SecurityGroups))

	for _, sg := range config.SecurityGroups {
		securityGroupsNames = append(securityGroupsNames, sg.Name)
	}

	annotations := copyMap(visaInstance.Annotations)
	annotations["visa.k8s.cells.es/security-groups"] = strings.Join(securityGroupsNames, ",")
	annotations["visa.k8s.cells.es/home-template-path"] = config.Image.Spec.HomeTemplatePath
	annotations["visa.k8s.cells.es/user-shell"] = config.Image.Spec.UserShell
	annotations["visa.k8s.cells.es/reboot-generation"] = strconv.FormatInt(visaInstance.Status.ObservedRebootGeneration, 10)

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
			Containers: containers,
		},
	}

	annotations["cloudprovider.visa.k8s/spec-hash"] = hashPodSpec(pod.Spec)

	return pod
}

func hashPodSpec(spec corev1.PodSpec) string {
	data, err := json.Marshal(spec)
	if err != nil {
		panic(err)
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
