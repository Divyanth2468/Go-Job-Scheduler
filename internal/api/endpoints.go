package endpoints

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/Divyanth2468/go-job-scheduler/internal/jobs"
	"github.com/Divyanth2468/go-job-scheduler/internal/logs"
	"github.com/Divyanth2468/go-job-scheduler/internal/scheduler"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/robfig/cron/v3"
)

var cronParser = cron.NewParser(
	cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

func Endpoints() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/alljobs", func(w http.ResponseWriter, r *http.Request) {
		jobsList, err := jobs.LoadJobs()
		if err != nil {
			http.Error(w, "Failed to load jobs\n", http.StatusInternalServerError)
			return
		}
		if jobsList == nil {
			if _, err := w.Write([]byte("No jobs exists")); err != nil {
				logs.LogAndPrint("Error sending reply: %v\n", err)
				return
			}
		} else {
			render.JSON(w, r, jobsList)
		}

	})

	r.Get("/job/{jobName}", func(w http.ResponseWriter, r *http.Request) {
		jobsList, err := jobs.LoadJobs()
		if err != nil {
			http.Error(w, "Failed to load jobs\n", http.StatusInternalServerError)
			return
		}

		jobName := chi.URLParam(r, "jobName")
		for _, job := range jobsList {
			if jobName == job.Name {
				w.WriteHeader(http.StatusOK)
				render.JSON(w, r, job)
				logs.LogAndPrint("Job %s found\n", job.Name)
				return
			}
		}
		w.Write([]byte(fmt.Sprintf("Job %s not found", jobName)))
		logs.LogAndPrint("Job %s not found\n", jobName)
	})

	r.Post("/jobs/", func(w http.ResponseWriter, r *http.Request) {
		jobdata, err := jobs.GetJobs(w, r)
		if err != nil {
			logs.LogAndPrint(err.Error())
			panic(err)
		}

		Validation(w, r, jobdata)
		jobdata.Retries += 1
		id, err := jobs.SaveJobs(jobdata)
		if err != nil {
			logs.LogAndPrint(err.Error())
			http.Error(w, "Internal Server error", http.StatusInternalServerError)
		}

		jobdata.ID = id

		if err := scheduler.RegisterJobs(jobdata); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	})

	r.Post("/delete/{jobName}", func(w http.ResponseWriter, r *http.Request) {
		jobName := chi.URLParam(r, "jobName")
		if err := scheduler.DeleteJob(jobName); err != nil {
			logs.LogAndPrint(err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Job deleted successfully"))
			logs.LogAndPrint("Job %s deleted successfully\n", jobName)
		}
	})

	r.Put("/update/{jobName}", func(w http.ResponseWriter, r *http.Request) {
		jobName := chi.URLParam(r, "jobName")
		var jobdata jobs.JobRequest
		jobdata.Name = jobName

		existingJob, flag := jobs.JobExists(jobdata.Name)

		if flag == 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Job %s doesn't exist", jobdata.Name)))
		} else {
			// Parse new job data
			updatedJob, err := jobs.GetJobs(w, r)
			if err != nil {
				http.Error(w, "Invalid request data\n", http.StatusBadRequest)
				return
			}

			Validation(w, r, updatedJob)
			updatedJob.Retries += 1

			if reflect.DeepEqual(updatedJob, existingJob) {
				w.Write([]byte(fmt.Sprintf("No changes detected for job: %s. Skipping update.\n", jobName)))
				logs.LogAndPrint("No changes detected for job: %s. Skipping update.\n", jobName)
				return
			}

			// Try deleting the old job
			if err := scheduler.UpdateJob(jobName, updatedJob); err != nil {
				http.Error(w, "Job not updated\n", http.StatusInternalServerError)
				logs.LogAndPrint("Job %s not updated, Error: %v\n", updatedJob.Name, err)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Job %s updated successfully", updatedJob.Name)))
			logs.LogAndPrint("Job %s updated successfully\n", updatedJob.Name)
		}
	})

	r.Get("/job_run/{job_name}", func(w http.ResponseWriter, r *http.Request) {
		jobName := chi.URLParam(r, "job_name")
		job, flag := jobs.JobExists(jobName)
		if flag != 1 {
			http.Error(w, "Job Doesn't exist", http.StatusBadRequest)
		} else {
			runs, err := jobs.GetJobRunsById(job.ID)
			if err != nil {
				http.Error(w, "Internal Error, unable to retrieve job runs", http.StatusInternalServerError)
				logs.LogAndPrint("Unable to retrieve job runs: %v\n", err.Error())
			}
			render.JSON(w, r, runs)
		}
	})

	http.ListenAndServe(":3000", r)
}

func Validation(w http.ResponseWriter, r *http.Request, jobdata jobs.JobRequest) {
	if _, err := cronParser.Parse(jobdata.CronExpr); err != nil {
		logs.LogAndPrint("Invalid cron expression %v %s\n", err, jobdata.CronExpr)
		http.Error(w, "Invalid cron expression", http.StatusBadRequest)
		return
	}
}
