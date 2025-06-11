package endpoints

import (
	"fmt"
	"log"
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
			http.Error(w, "Failed to load jobs", http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, jobsList)
	})

	r.Get("/job/{jobName}", func(w http.ResponseWriter, r *http.Request) {
		jobsList, err := jobs.LoadJobs()
		if err != nil {
			http.Error(w, "Failed to load jobs", http.StatusInternalServerError)
			return
		}

		jobName := chi.URLParam(r, "jobName")
		for _, job := range jobsList {
			if jobName == job.Name {
				w.WriteHeader(http.StatusOK)
				render.JSON(w, r, job)
				logs.LogAndPrint("Job %s found", job.Name)
				return
			}
		}
		w.Write([]byte(fmt.Sprintf("Job %s not found", jobName)))
		logs.LogAndPrint("Job %s not found", jobName)
	})

	r.Post("/jobs/", func(w http.ResponseWriter, r *http.Request) {
		jobdata, err := jobs.GetJobs(w, r)
		if err != nil {
			panic(err)
		}

		Validation(w, r, jobdata)

		if err := jobs.SaveJobs(jobdata); err != nil {
			log.Println(err)
		}

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
			logs.LogAndPrint("Job %s deleted successfully", jobName)
		}
	})

	r.Put("/update/{jobName}", func(w http.ResponseWriter, r *http.Request) {
		jobName := chi.URLParam(r, "jobName")
		var jobdata jobs.JobRequest
		jobdata.Name = jobName

		existingJob, flag := jobs.JobExists(jobdata)

		if flag == 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Job %s doesn't exist", jobdata.Name)))
		} else {
			// Parse new job data
			updatedJob, err := jobs.GetJobs(w, r)
			if err != nil {
				http.Error(w, "Invalid request data", http.StatusBadRequest)
				return
			}

			Validation(w, r, updatedJob)

			if reflect.DeepEqual(updatedJob, existingJob) {
				w.Write([]byte(fmt.Sprintf("No changes detected for job: %s. Skipping update.\n", jobName)))
				logs.LogAndPrint("No changes detected for job: %s. Skipping update.\n", jobName)
				return
			}

			// Try deleting the old job
			if err := scheduler.DeleteJob(jobName); err != nil {
				http.Error(w, "Job not updated", http.StatusInternalServerError)
				logs.LogAndPrint("Job %s not updated", updatedJob.Name)
				return
			}

			if err := jobs.SaveJobs(updatedJob); err != nil {
				log.Println(err)
			}

			// Register and persist the updated job
			if err := scheduler.RegisterJobs(updatedJob); err != nil {
				http.Error(w, "Failed to register updated job", http.StatusInternalServerError)
				logs.LogAndPrint("Failed to register updated job %s", updatedJob.Name)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Job %s updated successfully", updatedJob.Name)))
			logs.LogAndPrint("Job %s updated successfully", updatedJob.Name)
		}
	})

	http.ListenAndServe(":3000", r)
}

func Validation(w http.ResponseWriter, r *http.Request, jobdata jobs.JobRequest) {
	if jobdata.Type != "shell" && jobdata.Type != "http" {
		http.Error(w, "Invalid job type", http.StatusBadRequest)
		return
	}
	if len(jobdata.Command) == 0 {
		http.Error(w, "Invalid command", http.StatusBadRequest)
	}
	if _, err := cron.ParseStandard(jobdata.CronExpr); err != nil {
		http.Error(w, "Invalid cron expression", http.StatusBadRequest)
		return
	}

	if jobdata.Retries == 0 {
		jobdata.Retries += 1
	}
}
