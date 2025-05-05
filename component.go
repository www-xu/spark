package spark

import "context"

type IComponent[T any] interface {
	Instantiate() error
	Close() error
}

type ISingleSourceComponent[T any] interface {
	Get(ctx context.Context) T
}

type IMultiSourceComponent[T any] interface {
	Get(ctx context.Context, name string) T
}
