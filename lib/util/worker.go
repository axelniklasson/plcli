package util

// Job represents a job performed by a worker
type Job struct {
	Func func(args interface{}) (interface{}, error)
	Args interface{}
}

// JobResult represents the result of a job performed by a worker.
type JobResult struct {
	// Possible error when running job
	Error error
	// Arbitrary result data
	Result interface{}
}

// Worker is a generic worker function that performs as many jobs as possible from a given channel before until it's empty. Meant to
// run as a goroutine.
func Worker(id int, jobs <-chan Job, results chan<- JobResult) {
	// collect jobs from channel and do work before sending results back over channel
	for job := range jobs {
		res, err := job.Func(job.Args)
		results <- JobResult{err, res}
	}
}
