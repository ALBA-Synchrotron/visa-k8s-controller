package resource

import (
	"crypto/sha256"
	"encoding/hex"

	v1 "github.com/ALBA-Synchrotron/visa-k8s-controller/api/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

func HashVISACloudInstanceSpec(spec v1.VISACloudInstanceSpec) string {
	data, err := json.Marshal(spec)
	if err != nil {
		panic(err)
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
