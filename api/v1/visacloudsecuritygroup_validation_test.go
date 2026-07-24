package v1

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("VISACloudSecurityGroup validation", func() {
	ginkgo.It("accepts valid object", func() {
		device := &VISACloudSecurityGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sec-group",
				Namespace: "default",
			},
		}
		gomega.Expect(k8sClient.Create(ctx, device)).To(gomega.Succeed())
	})
})
