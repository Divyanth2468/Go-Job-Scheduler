package jobs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	database "github.com/Divyanth2468/go-job-scheduler/internal/data"
	"github.com/Divyanth2468/go-job-scheduler/internal/logs"
	"github.com/google/uuid"
)

type JobRequest struct {
	ID        uuid.UUID
	Name      string `json:"name"`
	Type      string `json:"type"`
	CronExpr  string `json:"cron_expr"`
	LambdaArn string `json:"lambda_arn"`
	Command   string `json:"command"`
	Retries   int    `json:"retries"`
	CreatedAt time.Time
}

var Jobs []JobRequest

func GetJobs(w http.ResponseWriter, r *http.Request) (JobRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body\n", http.StatusBadRequest)
		return JobRequest{}, err

	}
	var jobdata JobRequest
	err = json.Unmarshal(body, &jobdata)
	if err != nil {
		http.Error(w, "Error Unmarshalling JSON\n", http.StatusBadRequest)
		return JobRequest{}, err
	}

	// Now the JSON data is in jobdata
	logs.LogAndPrint("Recieved data: %v\n", jobdata)

	_, flag := JobExists(jobdata.Name)

	if flag == 1 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Job %s exists", jobdata.Name)))
		logs.LogAndPrint("Job %s already exists\n", jobdata.Name)
	}

	if flag == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Job %s added successfully", jobdata.Name)))
		logs.LogAndPrint("Job %s added successfully\n", jobdata.Name)
	}

	return jobdata, nil
}

func JobExists(jobname string) (JobRequest, int) {
	var j JobRequest
	flag := 0
	if len(Jobs) != 0 {
		for _, job := range Jobs {
			if job.Name == jobname {
				j = job
				flag = 1
			}
		}
	}
	return j, flag
}

func LoadJobs() ([]JobRequest, error) {
	var jobs []JobRequest
	rows, err := database.Db.Query("SELECT id, name, type, cron_expr, lambda_arn, command, retries, created_at FROM jobs")
	if err != nil {
		logs.LogAndPrint(err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var job JobRequest
		err := rows.Scan(&job.ID, &job.Name, &job.Type, &job.CronExpr, &job.LambdaArn, &job.Command, &job.Retries, &job.CreatedAt)
		if err != nil {
			logs.LogAndPrint(err.Error())
			return nil, err
		}
		jobs = append(jobs, job)
	}
	Jobs = jobs
	return jobs, err
}

func SaveJobs(job JobRequest) (uuid.UUID, error) {
	// Check if the job already exists
	for _, j := range Jobs {
		if j.Name == job.Name {
			log.Println("[SAVE] Job with same name already exists:", job.Name)
			return uuid.Nil, nil // Don't save duplicate
		}
	}

	Jobs = append(Jobs, job)

	if _, err := database.Db.Exec(`INSERT INTO jobs (name, type, cron_expr, lambda_arn, command, retries, created_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7)`, job.Name, job.Type, job.CronExpr, job.LambdaArn,
		job.Command, job.Retries, time.Now()); err != nil {
		logs.LogAndPrint("Error saving to database: %v\n", err.Error())
		return uuid.Nil, err
	}

	var id uuid.UUID

	row := database.Db.QueryRow(`SELECT id FROM jobs WHERE name = $1`, job.Name)
	if err := row.Scan(&id); err != nil {
		logs.LogAndPrint("Error retrieving job ID: %v", err.Error())
		return uuid.Nil, err
	}

	return id, nil
}

func UpdateJobData(job JobRequest) error {
	if _, err := database.Db.Exec(`UPDATE jobs 
	SET cron_expr = $1, lambda_arn = $2, command = $3, type = $4, retries = $5, created_at = $6
	WHERE name = $7`, job.CronExpr, job.LambdaArn, job.Command, job.Type, job.Retries, job.CreatedAt, job.Name); err != nil {
		logs.LogAndPrint("Error updating job data of %s: %v\n", job.Name, err)
		return err
	}
	return nil
}

func DeleteFromJobsData(jobname string) error {
	var newjobs []JobRequest
	found := false

	for _, job := range Jobs {
		if job.Name == jobname {
			// Delete from DB
			if _, err := database.Db.Exec(`DELETE FROM jobs WHERE name = $1`, jobname); err != nil {
				logs.LogAndPrint("Not able to delete job %v, Error %v\n", jobname, err)
				return err
			}
			logs.LogAndPrint("Successfully deleted job %v\n", jobname)
			found = true
			continue // Skip adding this job to newjobs
		}
		newjobs = append(newjobs, job)
	}

	if found {
		Jobs = newjobs // update only if deleted
		return nil
	}
	return fmt.Errorf("no job with the name '%s' exists", jobname)
}
