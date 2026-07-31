package v1

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("VISACloudImage validation", func() {
	ginkgo.It("accepts valid object", func() {
		device := &VISACloudImage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-image",
				Namespace: "default",
			},
			Spec: VISACloudImageSpec{
				FullImageReference: "registry.blah.cells/test-image:latest",
				Size:               "2654.23",
				HomeTemplatePath:   "/etc/home/template",
				UserShell:          "/usr/bin/bash",
			},
		}
		gomega.Expect(k8sClient.Create(ctx, device)).To(gomega.Succeed())
	})
})
