package runner

import (
	"net/http"
	"os/exec"
	"time"

	"github.com/Divyanth2468/go-job-scheduler/internal/jobs"
	"github.com/Divyanth2468/go-job-scheduler/internal/logs"
)

func Execute(job jobs.JobRequest) {
	if job.Type == "shell" {
		flag := 0
		var output []byte
		var err error
		logs.LogAndPrint("[SHELL] Executing shell job: Name=%s, Command=%s\n", job.Name, job.Command)
		for range job.Retries {
			cmd := exec.Command("sh", "-c", job.Command)
			output, err = cmd.CombinedOutput()
			if err == nil {
				flag = 1
				logs.LogAndPrint("[SHELL] Success: Name=%s, Output=%s\n", job.Name, string(output))
				break
			}
			logs.LogAndPrint("[SHELL] Attempt failed: %v\nOutput: %s", err, string(output))
			time.Sleep(time.Second * 1)
		}
		if flag == 0 {
			logs.LogAndPrint("[SHELL] Failed after retries: Name=%s, Error=%v\n", job.Name, err)
		}

	} else {
		flag := 0
		logs.LogAndPrint("[HTTP] Sending HTTP GET: Name=%s, URL=%s\n", job.Name, job.Command)
		var errhttp error
		client := &http.Client{Timeout: 5 * time.Second}
		for range job.Retries {
			r, err := client.Get(job.Command)
			if err == nil && r.StatusCode >= 200 && r.StatusCode < 300 {
				logs.LogAndPrint("[HTTP] Success: Name=%s, StatusCode=%d\n", job.Name, r.StatusCode)
				flag = 1
				break
			}
			time.Sleep(time.Second * 1)
			errhttp = err
		}
		if flag == 0 {
			logs.LogAndPrint("[HTTP] Failed after retries: Name=%s, LastError=%v\n", job.Name, errhttp)
		}
	}
}
