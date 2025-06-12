package runner

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/Divyanth2468/go-job-scheduler/internal/jobs"
	"github.com/Divyanth2468/go-job-scheduler/internal/logs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

func Execute(job jobs.JobRequest) {
	status := "failure"
	var op string
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
				status = "success"
				logs.LogAndPrint("[SHELL] Success: Name=%s, Output=%s\n", job.Name, string(output))
				break
			}
			logs.LogAndPrint("[SHELL] Attempt failed: %v\nOutput: %s", err, string(output))
			time.Sleep(time.Second * 1)
		}
		if flag == 0 {
			logs.LogAndPrint("[SHELL] Failed after retries: Name=%s, Error=%v\n", job.Name, err)
		}
		if err != nil {
			op = err.Error()
		} else {
			op = string(output)
		}

	} else if job.Type == "http" {
		flag := 0
		code := 0
		logs.LogAndPrint("[HTTP] Sending HTTP GET: Name=%s, URL=%s\n", job.Name, job.Command)
		var errhttp error
		client := &http.Client{Timeout: 5 * time.Second}
		for range job.Retries {
			r, err := client.Get(job.Command)
			if err == nil && r.StatusCode >= 200 && r.StatusCode < 300 {
				logs.LogAndPrint("[HTTP] Success: Name=%s, StatusCode=%d\n", job.Name, r.StatusCode)
				flag = 1
				status = "success"
				code = r.StatusCode
				break
			}
			time.Sleep(time.Second * 1)
			errhttp = err
		}
		if flag == 0 {
			logs.LogAndPrint("[HTTP] Failed after retries: Name=%s, LastError=%v\n", job.Name, errhttp)
		}
		if errhttp != nil {
			op = errhttp.Error()
		} else {
			op = string("Status code is: " + strconv.Itoa(code))
		}
	} else {
		flag := 0
		var errlambda error
		var res_payload string
		for range job.Retries {
			ctx := context.TODO()

			cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-north-1"))

			if err != nil {
				logs.LogAndPrint("Error setting context: %v\n", err.Error())
				errlambda = err
			}

			lambdaClient := lambda.NewFromConfig(cfg)

			// Prepare the payload
			payload := map[string]string{
				"url": job.Command, // Command holds the URL or whatever you want to send
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				logs.LogAndPrint(err.Error())
				errlambda = err
			}

			// Invoke Lambda
			input := &lambda.InvokeInput{
				FunctionName: aws.String(job.LambdaArn),
				Payload:      payloadBytes,
			}

			result, err := lambdaClient.Invoke(ctx, input)
			if err != nil {
				logs.LogAndPrint(err.Error())
				errlambda = err
			}

			if errlambda == nil {
				flag = 1
				logs.LogAndPrint("[LAMBDA] Response: %s", result.Payload)
				res_payload = string(result.Payload)
				status = "success"
				break
			}
		}
		if flag == 0 {
			logs.LogAndPrint("[LAMBDA] Failed after retries: Name=%s, LastError=%v\n", job.Name, errlambda)
		}
		if errlambda != nil {
			op = string(errlambda.Error())
		} else {
			op = res_payload
		}
	}
	if jobrunerr := jobs.SaveJobRuns(jobs.JobRuns{
		JobID:   job.ID,
		JobName: job.Name,
		JobType: job.Type,
		Status:  status,
		RunAt:   time.Now(),
		Log:     op,
	}); jobrunerr != nil {
		logs.LogAndPrint("Error writing to job_runs %v\n", jobrunerr)
	}

}
