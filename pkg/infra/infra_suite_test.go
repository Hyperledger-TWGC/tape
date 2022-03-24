package infra_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInfra(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Infra Suite")
}
