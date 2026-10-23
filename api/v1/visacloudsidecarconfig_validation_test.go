package v1

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("VISACloudSidecarConfig validation", func() {
	ginkgo.It("accepts valid object", func() {
		device := &VISACloudSidecarConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-guacd-sidecar",
				Namespace: "default",
			},
			Spec: VISACloudSidecarConfigSpec{
				ContainerSpec: ContainerSpec{
					FullImageReference: "guacd:latest",
				},
			},
		}
		gomega.Expect(k8sClient.Create(ctx, device)).To(gomega.Succeed())
	})

	ginkgo.It("rejects empty image reference", func() {
		device := &VISACloudSidecarConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-guacd-sidecar",
				Namespace: "default",
			},
			Spec: VISACloudSidecarConfigSpec{
				ContainerSpec: ContainerSpec{
					FullImageReference: "aguacate:latest",
				},
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
