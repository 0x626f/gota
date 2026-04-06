package workers

type Runner = func()

type Task = func() error
type Callback[T any] = func(T) error

type ErrorHandler = func(error)
type RecoveryHandler = func(any)
