package maybe

type Maybe[T any] struct {
	v     T
	valid bool
}

func Just[T any](maybe Maybe[T]) (T, bool) {
	if maybe.valid {
		return maybe.v, true
	}
	return maybe.v, false
}

func New[T any](v T) Maybe[T] {
	return Maybe[T]{
		v:     v,
		valid: true,
	}
}

func None[T any]() Maybe[T] {
	return Maybe[T]{}
}
