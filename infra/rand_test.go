package infra

import (
	"fmt"
	"testing"
)

func Test_RandArgs(t *testing.T) {

	args := RandArgs(20, 20)
	fmt.Println(args)

}
