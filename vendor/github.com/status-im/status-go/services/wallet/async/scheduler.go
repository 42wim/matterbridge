package async

import (
	"context"
	"errors"
	"fmt"
	"sync"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

var ErrTaskOverwritten = errors.New("task overwritten")

type Scheduler struct {
	queue      *orderedmap.OrderedMap[TaskType, *taskContext]
	queueMutex sync.Mutex

	context                context.Context
	cancelFn               context.CancelFunc
	doNotDeleteCurrentTask bool
}

type ReplacementPolicy = int

const (
	// ReplacementPolicyCancelOld for when the task arguments might change the result
	ReplacementPolicyCancelOld ReplacementPolicy = iota
	// ReplacementPolicyIgnoreNew for when the task arguments doesn't change the result
	ReplacementPolicyIgnoreNew
)

type TaskType struct {
	ID     int64
	Policy ReplacementPolicy
}

type taskFunction func(context.Context) (interface{}, error)
type resultFunction func(interface{}, TaskType, error)

type taskContext struct {
	taskType TaskType
	policy   ReplacementPolicy

	taskFn taskFunction
	resFn  resultFunction
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		queue: orderedmap.New[TaskType, *taskContext](),
	}
}

// Enqueue provides a queue of task types allowing only one task at a time of the corresponding type. The running task is the first one in the queue (s.queue.Oldest())
//
// Schedule policy for new tasks
//   - pushed at the back of the queue (s.queue.PushBack()) if none of the same time already scheduled
//   - overwrite the queued one of the same type, depending on the policy
//   - In case of ReplacementPolicyIgnoreNew, the new task will be ignored
//   - In case of ReplacementPolicyCancelOld, the old running task will be canceled or if not yet run overwritten and the new one will be executed when its turn comes.
//
// The task function (taskFn) might not be executed if
//   - the task is ignored
//   - the task is overwritten. The result function (resFn) will be called with ErrTaskOverwritten
//
// The result function (resFn) will always be called if the task is not ignored
func (s *Scheduler) Enqueue(taskType TaskType, taskFn taskFunction, resFn resultFunction) (ignored bool) {
	s.queueMutex.Lock()
	defer s.queueMutex.Unlock()

	taskRunning := s.queue.Len() > 0
	existingTask, typeInQueue := s.queue.Get(taskType)

	newTask := &taskContext{
		taskType: taskType,
		policy:   taskType.Policy,
		taskFn:   taskFn,
		resFn:    resFn,
	}

	if taskRunning {
		if typeInQueue {
			if s.queue.Oldest().Value.taskType == taskType {
				// If same task type is running
				if existingTask.policy == ReplacementPolicyCancelOld {
					// If a previous task is running, cancel it
					if s.cancelFn != nil {
						s.cancelFn()
						s.cancelFn = nil
					} else {
						// In case of multiple tasks of the same type, the previous one is overwritten
						go func() {
							existingTask.resFn(nil, existingTask.taskType, ErrTaskOverwritten)
						}()
					}

					s.doNotDeleteCurrentTask = true

					// Add it again to refresh the order of the task
					s.queue.Delete(taskType)
					s.queue.Set(taskType, newTask)
				} else {
					ignored = true
				}
			} else {
				// if other task type is running
				// notify the queued one that it is overwritten or ignored
				if existingTask.policy == ReplacementPolicyCancelOld {
					go func() {
						existingTask.resFn(nil, existingTask.taskType, ErrTaskOverwritten)
					}()
					// Overwrite the queued one of the same type
					existingTask.taskFn = taskFn
					existingTask.resFn = resFn
				} else {
					ignored = true
				}
			}
		} else {
			// Policy does not matter for the fist enqueued task of a type
			s.queue.Set(taskType, newTask)
		}
	} else {
		// If no task is running add and run it. The worker will take care of scheduling new tasks added while running
		s.queue.Set(taskType, newTask)
		existingTask = newTask
		s.runTask(existingTask, taskFn, func(res interface{}, runningTask *taskContext, err error) {
			s.finishedTask(res, runningTask, resFn, err)
		})
	}

	return ignored
}

func (s *Scheduler) runTask(tc *taskContext, taskFn taskFunction, resFn func(interface{}, *taskContext, error)) {
	thisContext, thisCancelFn := context.WithCancel(context.Background())
	s.cancelFn = thisCancelFn
	s.context = thisContext

	go func() {
		res, err := taskFn(thisContext)

		// Release context resources
		thisCancelFn()

		if errors.Is(err, context.Canceled) {
			resFn(res, tc, fmt.Errorf("task canceled: %w", err))
		} else {
			resFn(res, tc, err)
		}
	}()
}

// finishedTask is the only one that can remove a task from the queue
// if the current running task completed (doNotDeleteCurrentTask is true)
func (s *Scheduler) finishedTask(finishedRes interface{}, doneTask *taskContext, finishedResFn resultFunction, finishedErr error) {
	s.queueMutex.Lock()

	// We always have a running task
	current := s.queue.Oldest()
	// Delete current task if not overwritten
	if s.doNotDeleteCurrentTask {
		s.doNotDeleteCurrentTask = false
	} else {
		s.queue.Delete(current.Value.taskType)
	}

	// Run next task
	if pair := s.queue.Oldest(); pair != nil {
		nextTask := pair.Value
		s.runTask(nextTask, nextTask.taskFn, func(res interface{}, runningTask *taskContext, err error) {
			s.finishedTask(res, runningTask, runningTask.resFn, err)
		})
	} else {
		s.cancelFn = nil
	}
	s.queueMutex.Unlock()

	// Report result
	finishedResFn(finishedRes, doneTask.taskType, finishedErr)
}

func (s *Scheduler) Stop() {
	s.queueMutex.Lock()
	defer s.queueMutex.Unlock()

	if s.cancelFn != nil {
		s.cancelFn()
		s.cancelFn = nil
	}

	// Empty the queue so the running task will not be restarted
	for pair := s.queue.Oldest(); pair != nil; pair = pair.Next() {
		// Notify the queued one that they are canceled
		if pair.Value.policy == ReplacementPolicyCancelOld {
			go func() {
				pair.Value.resFn(nil, pair.Value.taskType, context.Canceled)
			}()
		}
		s.queue.Delete(pair.Value.taskType)
	}
}
