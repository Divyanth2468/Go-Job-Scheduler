package main

import (
	endpoints "github.com/Divyanth2468/go-job-scheduler/internal/api"
	"github.com/Divyanth2468/go-job-scheduler/internal/jobs"
	"github.com/Divyanth2468/go-job-scheduler/internal/logs"
	"github.com/Divyanth2468/go-job-scheduler/internal/scheduler"
)

func main() {
	jobs.Init()
	scheduler.Init()
	scheduler.LogFileSetup()
	savedJobs, err := jobs.LoadJobs()

	if err != nil {
		logs.LogAndPrint("[ERROR] Failed to load jobs: %v\n", err)
	} else {
		for _, job := range savedJobs {
			err := scheduler.RegisterJobs(job)
			if err != nil {
				logs.LogAndPrint("[ERROR] Failed to register job %s: %v\n", job.Name, err)
			}
		}
		logs.LogAndPrint("[BOOT] Registered %d persisted jobs.\n", len(savedJobs))
	}

	endpoints.Endpoints()

}
