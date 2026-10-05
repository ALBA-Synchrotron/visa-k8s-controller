package v1

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("VISASidecarConfig validation", func() {
	ginkgo.It("accepts valid object", func() {
		device := &VISASidecarConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-guacd-sidecar",
				Namespace: "default",
			},
			Spec: VISASidecarConfigSpec{
				ContainerSpec: ContainerSpec{
					FullImageReference: "guacd:latest",
				},
			},
		}
		gomega.Expect(k8sClient.Create(ctx, device)).To(gomega.Succeed())
	})

	ginkgo.It("rejects empty image reference", func() {
		device := &VISASidecarConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-guacd-sidecar",
				Namespace: "default",
			},
			Spec: VISASidecarConfigSpec{
				ContainerSpec: ContainerSpec{
					FullImageReference: "aguacate:latest",
				},
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
