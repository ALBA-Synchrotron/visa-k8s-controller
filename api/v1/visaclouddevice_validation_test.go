package v1

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("VISACloudDevice validation", func() {
	ginkgo.It("accepts VIRTUAL_GPU as type", func() {
		device := &VISACloudDevice{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-vgpu",
				Namespace: "default",
			},
			Spec: VISACloudDeviceSpec{
				Type:       DeviceTypeVirtualGPU,
				Identifier: "test-vgpu01",
			},
		}
		gomega.Expect(k8sClient.Create(ctx, device)).To(gomega.Succeed())
	})

	ginkgo.It("accepts PCI_PASSTHROUGH as type", func() {
		device := &VISACloudDevice{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pci",
				Namespace: "default",
			},
			Spec: VISACloudDeviceSpec{
				Type:       DeviceTypePCIPassthrough,
				Identifier: "test-pci01",
			},
		}
		gomega.Expect(k8sClient.Create(ctx, device)).To(gomega.Succeed())
	})

	ginkgo.It("rejects invalid type as type", func() {
		device := &VISACloudDevice{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "invalid-type",
				Namespace: "default",
			},
			Spec: VISACloudDeviceSpec{
				Type:       "Invalid Type",
				Identifier: "test-pci01",
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("rejects empty identifier", func() {
		device := &VISACloudDevice{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "invalid-type",
				Namespace: "default",
			},
			Spec: VISACloudDeviceSpec{
				Type:       "Invalid Type",
				Identifier: "",
			},
		}
		err := k8sClient.Create(ctx, device)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
