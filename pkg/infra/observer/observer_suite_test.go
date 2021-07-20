package observer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestObserver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Observer Suite")
}
