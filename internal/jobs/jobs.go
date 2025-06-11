package jobs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Divyanth2468/go-job-scheduler/internal/logs"
)

type JobRequest struct {
	Name     string `json:"name"`
	CronExpr string `json:"cron_expr"`
	Type     string `json:"type"`
	Command  string `json:"command"`
	Retries  int    `json:"retries"`
}

// Get current directory and move one level up
var logFilePath string
var Jobs []JobRequest

func Init() {
	wd, err := os.Getwd()

	if err != nil {
		log.Fatal("Unable to get working directory:", err)
	}
	parentDir := filepath.Dir(wd)
	logFilePath = filepath.Join(parentDir, "internal", "data", "jobs.json")
}

func GetJobs(w http.ResponseWriter, r *http.Request) (JobRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return JobRequest{}, err

	}
	var jobdata JobRequest
	err = json.Unmarshal(body, &jobdata)
	if err != nil {
		http.Error(w, "Error Unmarshalling JSON", http.StatusBadRequest)
		return JobRequest{}, err
	}

	// Now the JSON data is in jobdata
	log.Println("Recieved data: ", jobdata)

	_, flag := JobExists(jobdata)

	if flag == 1 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Job %s exists", jobdata.Name)))
		logs.LogAndPrint("Job %s already exists", jobdata.Name)
	}

	if flag == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Job %s added successfully", jobdata.Name)))
		logs.LogAndPrint("Job %s added successfully", jobdata.Name)
	}

	return jobdata, nil
}

func JobExists(jobdata JobRequest) (JobRequest, int) {
	var j JobRequest
	flag := 0
	if len(Jobs) != 0 {
		for _, job := range Jobs {
			if job.Name == jobdata.Name {
				j = job
				flag = 1
			}
		}
	}
	return j, flag
}

func LoadJobs() ([]JobRequest, error) {
	data, err := os.ReadFile(logFilePath)
	if err != nil {
		return nil, err
	}
	var jobs []JobRequest
	err = json.Unmarshal(data, &jobs)
	Jobs = jobs
	return jobs, err
}

func SaveJobs(job JobRequest) error {
	jobs, err := LoadJobs()
	if err != nil {
		jobs = []JobRequest{}
	}
	// Check if the job already exists
	for _, j := range jobs {
		if j.Name == job.Name {
			log.Println("[SAVE] Job with same name already exists:", job.Name)
			return nil // Don't save duplicate
		}
	}

	jobs = append(jobs, job)
	Jobs = jobs
	data, _ := json.MarshalIndent(jobs, "", "  ")
	logs.LogAndPrint("Jobs are being saved")
	return os.WriteFile(logFilePath, data, 0644)
}

func DeleteFromJobsData(jobname string) error {
	jobsList, err := LoadJobs()
	if err != nil {
		return err
	}

	newList := make([]JobRequest, 0)
	for _, job := range jobsList {
		if job.Name != jobname {
			newList = append(newList, job)
		}
	}

	Jobs = newList
	data, err := json.MarshalIndent(newList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(logFilePath, data, 0644)
}
