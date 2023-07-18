package dagRun

import (
	"fmt"
)

type TaskManager[T Task[V], V any] struct {
	Registry[T]
}

func NewTaskManager[T Task[V], V any]() TaskManager[T, V] {
	return TaskManager[T, V]{Registry: NewRegistry[T]()}
}

// GetAllTaskWithDepsByName get all Tasks with their parent dependencies by names
func (t TaskManager[T, V]) GetAllTaskWithDepsByName(taskNames []string) (map[string]T, error) {
	allTasks := make(map[string]T)
	var registerTaskByName func(taskName string) error
	registerTaskByName = func(taskName string) error {
		if _, ok := allTasks[taskName]; ok {
			return nil
		}
		task, err := t.Get(taskName)
		if err != nil {
			return fmt.Errorf("not get task by name:%s", taskName)
		}
		allTasks[task.Name()] = task
		for _, dep := range task.Dependencies() {
			err := registerTaskByName(dep)
			if err != nil {
				return err
			}
		}
		return err
	}
	for _, name := range taskNames {
		err := registerTaskByName(name)
		if err != nil {
			return nil, err
		}
	}
	return allTasks, nil
}
