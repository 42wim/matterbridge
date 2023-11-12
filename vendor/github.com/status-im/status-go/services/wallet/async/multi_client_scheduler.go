package async

type MultiClientScheduler struct {
	scheduler *Scheduler
}

func NewMultiClientScheduler() *MultiClientScheduler {
	return &MultiClientScheduler{
		scheduler: NewScheduler(),
	}
}

func (s *MultiClientScheduler) Stop() {
	s.scheduler.Stop()
}

func makeTaskType(requestID int32, origTaskType TaskType) TaskType {
	return TaskType{
		ID:     int64(requestID)<<32 | origTaskType.ID,
		Policy: origTaskType.Policy,
	}
}

func (s *MultiClientScheduler) Enqueue(requestID int32, taskType TaskType, taskFn taskFunction, resFn resultFunction) (ignored bool) {
	return s.scheduler.Enqueue(makeTaskType(requestID, taskType), taskFn, resFn)
}
