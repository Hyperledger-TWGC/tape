package bitmap_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBitmap(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bitmap Suite")
}
