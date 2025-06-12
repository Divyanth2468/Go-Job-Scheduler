package jobs

import (
	"time"

	database "github.com/Divyanth2468/go-job-scheduler/internal/data"
	"github.com/Divyanth2468/go-job-scheduler/internal/logs"
	"github.com/google/uuid"
)

type JobRuns struct {
	ID      uuid.UUID `json:"id"`
	JobID   uuid.UUID `json:"job_id"`
	JobName string    `json:"job_name"`
	JobType string    `json:"job_type"`
	Status  string    `json:"status"`
	RunAt   time.Time `json:"run_time"`
	Log     string    `json:"log"`
}

func SaveJobRuns(jobrun JobRuns) error {
	logs.LogAndPrint("Saving job runs")
	if _, err := database.Db.Exec(`INSERT INTO job_runs (job_id, job_name, job_type, status, run_at, log)
  VALUES ($1, $2, $3, $4, $5, $6)`, jobrun.JobID, jobrun.JobName, jobrun.JobType, jobrun.Status, jobrun.RunAt, jobrun.Log); err != nil {
		return err
	}
	return nil
}

func GetJobRunsById(id uuid.UUID) ([]JobRuns, error) {
	var jobruns []JobRuns
	rows, err := database.Db.Query(`SELECT * FROM job_runs WHERE job_id = $1 ORDER BY run_at DESC`, id)
	if err != nil {
		logs.LogAndPrint("Error fetching job runs: %v\n", err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var jobrun JobRuns
		err := rows.Scan(&jobrun.ID, &jobrun.JobID, &jobrun.JobName, &jobrun.JobType, &jobrun.Status, &jobrun.Log, &jobrun.RunAt)
		if err != nil {
			logs.LogAndPrint("Error scanning job runs: %v\n", err.Error())
			return nil, err
		}
		jobruns = append(jobruns, jobrun)
	}

	return jobruns, nil
}
