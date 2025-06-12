package scheduler

import (
	"fmt"
	"log"

	"github.com/Divyanth2468/go-job-scheduler/internal/jobs"
	"github.com/Divyanth2468/go-job-scheduler/internal/logs"
	"github.com/Divyanth2468/go-job-scheduler/internal/runner"
	"github.com/robfig/cron/v3"
)

var cronScheduler *cron.Cron

var jobMap = make(map[string]cron.EntryID)

func Init() {
	// Create a cron scheduler
	cronScheduler = cron.New()

	// Start the cron scheduler
	cronScheduler.Start()

	log.Println("[INIT] Cron scheduler started and logging initialized.")
}

func RegisterJobs(job jobs.JobRequest) error {
	if _, exists := jobMap[job.Name]; exists {
		logs.LogAndPrint("[SKIP] Job '%s' already exists, skipping registration.\n", job.Name)
		return nil
	} else {
		id, err := cronScheduler.AddFunc(job.CronExpr, func() {
			runner.Execute(job)
		})
		if err != nil {
			return err
		}
		jobMap[job.Name] = id
		logs.LogAndPrint("[REGISTER] Job registered: Name=%s, Type=%s, Schedule=%s, Retries=%d, ID=%d\n", job.Name, job.Type, job.CronExpr, job.Retries, id)
		return nil
	}
}

func DeleteJob(jobName string) error {
	entryId, exists := jobMap[jobName]
	if !exists {
		log.Printf("Job with name %s doesn't exist\n", jobName)
		return fmt.Errorf("job with name %s doesn't exist", jobName)
	}
	cronScheduler.Remove(entryId)
	delete(jobMap, jobName)
	err := jobs.DeleteFromJobsData(jobName)

	if err != nil {
		logs.LogAndPrint("[DELETE] Failed to delete job from disk: Name=%s, Error=%v\n", jobName, err)
		return err
	}

	logs.LogAndPrint("[DELETE] Job deleted: Name=%s, EntryID=%d\n", jobName, entryId)
	return nil
}

func UpdateJob(jobName string, job jobs.JobRequest) error {
	entryId, exists := jobMap[jobName]
	if !exists {
		log.Printf("Job with name %s doesn't exist\n", jobName)
		return fmt.Errorf("job with name %s doesn't exist", jobName)
	}
	cronScheduler.Remove(entryId)
	delete(jobMap, jobName)

	if err := jobs.UpdateJobData(job); err != nil {
		return err
	}

	if err := RegisterJobs(job); err != nil {
		return err
	}
	return nil
}
