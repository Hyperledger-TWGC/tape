//go:build !race
// +build !race

package trafficGenerator_test

import (
	"testing"
	"unicode/utf8"

	"github.com/hyperledger-twgc/tape/pkg/infra/trafficGenerator"
)

func FuzzConvertString(f *testing.F) {
	testcases := []string{"data", "randomString1", "uuid", "randomNumber1_9", "{\"k1\":\"uuid\",\"key2\":\"randomNumber10000_20000\",\"keys\":\"randomString10\"}"}
	for _, tc := range testcases {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, orig string) {
		data, err := trafficGenerator.ConvertString(orig)
		if utf8.ValidString(orig) && err != nil && !utf8.ValidString(data) && len(data) != 0 {
			t.Errorf(err.Error() + " " + orig + " " + data)
		}
		if !utf8.ValidString(data) {
			t.Errorf("fail to convert utf8 string" + data)
		}
	})
}
