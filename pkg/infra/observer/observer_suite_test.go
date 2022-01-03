package observer_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestObserver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Observer Suite")
}
