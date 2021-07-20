package trafficGenerator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTrafficGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TrafficGenerator Suite")
}
