package infra

import "github.com/lifei6671/gorand"

func RandArgs() []string {
	key := gorand.NewUUID4().String()
	args := []string{"", key, "qyDPVDGnarJTnhjdBCCd"}
	return args
}
