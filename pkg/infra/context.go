package infra

import "context"

func TapeContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}
