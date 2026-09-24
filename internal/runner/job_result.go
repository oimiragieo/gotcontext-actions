package runner

import (
	"sync"

	"github.com/oimiragieo/gotcontext-actions/internal/model"
)

var jobResultLocks sync.Map // *model.Job -> *sync.Mutex

func jobResultMutex(job *model.Job) *sync.Mutex {
	if job == nil {
		return &sync.Mutex{}
	}
	v, _ := jobResultLocks.LoadOrStore(job, &sync.Mutex{})
	return v.(*sync.Mutex)
}

func storeJobResult(job *model.Job, result string) {
	if job == nil {
		return
	}
	mu := jobResultMutex(job)
	mu.Lock()
	defer mu.Unlock()
	// Failure wins across concurrent matrix cells.
	if job.Result == "failure" && result == "success" {
		return
	}
	job.Result = result
}

func loadJobResult(job *model.Job) string {
	if job == nil {
		return ""
	}
	mu := jobResultMutex(job)
	mu.Lock()
	defer mu.Unlock()
	return job.Result
}
