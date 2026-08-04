package resource

import (
	"crypto/sha256"
	"encoding/hex"

	v1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

func HashVISAInstanceSpec(spec v1.VISAInstanceSpec) string {
	data, err := json.Marshal(spec)
	if err != nil {
		panic(err)
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
