package workers

type Runner = func()

type ArgTask[T any] = func(T) error
type Task = func() error

type ErrorHandler = func(error)
type RecoveryHandler = func(any)
type FinishHandler = func()
