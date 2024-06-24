package dagRun

import "errors"

var (
	ErrNilTask      = errors.New("dagRun: nil task")
	ErrTaskExist    = errors.New("dagRun: task already exist")
	ErrNoTaskName   = errors.New("dagRun: no task name")
	ErrNilFunc      = errors.New("dagRun: nil func")
	ErrTaskNotExist = errors.New("dagRun: task not found")
	ErrSealed       = errors.New("dagRun: dag is sealed")
	ErrNotAsyncJob  = errors.New("dagRun: not async job")
)
