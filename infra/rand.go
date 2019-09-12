package infra

import "github.com/lifei6671/gorand"

func RandArgs(k, v uint) []string {
	key := gorand.RandomAlphabetic(k)
	value := gorand.RandomAlphabetic(v)
	args := []string{"", key, value}
	return args
}
