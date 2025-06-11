package scheduler

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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

func LogFileSetup() {
	// Get current directory and move one level up
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get working directory:", err)
	}
	parentDir := filepath.Dir(wd)
	logFilePath := filepath.Join(parentDir, "internal", "logs", "logs.txt")

	// Create/open log file
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to create/open log file:", err)
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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
		log.Printf("Job with name %s doesn't exist", jobName)
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
