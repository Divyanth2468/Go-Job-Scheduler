# ⏰ Go Job Scheduler

Go Job Scheduler is a cloud-native, backend-only job scheduler written in Go. It allows you to schedule Shell, HTTP, and AWS Lambda tasks using cron expressions. Designed for simplicity, flexibility, and easy extensibility, it's perfect for managing your automated tasks.

---

## 🚀 Features

- **Schedule Diverse Job Types:** Easily schedule Shell commands, HTTP requests, or AWS Lambda functions using standard cron expressions (e.g., `@every 30s`, `@hourly`, `@daily`).
- **Configurable Retries:** Set up automatic retries for failed jobs to ensure task completion.
- **Job Persistence:** All job configurations are stored and managed using PostgreSQL, ensuring data integrity and recovery.
- **RESTful API:** A comprehensive API allows you to programmatically manage the entire job lifecycle, including adding, updating, and deleting jobs.
- **Lightweight & Serverless-Ready:** Designed with a minimal footprint, making it suitable for serverless deployments and efficient resource usage.

---

## 🛠️ Tech Stack

- **Language:** Go
- **Database:** PostgreSQL
- **Cloud Integration:** AWS Lambda (via SDK)
- **Supported Job Types:** Shell, HTTP, Lambda
- **Architecture:** Serverless-ready, modular backend

---

## 📦 Installation

To get started with Go Job Scheduler, follow these steps:

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/Divyanth2468/go-job-scheduler.git
    cd go-job-scheduler
    ```

2.  **Run the Scheduler:**

    ```bash
    go run main.go
    ```

    **Prerequisites:**

    - Ensure **Go** and **PostgreSQL** are correctly set up on your system.
    - Update your **`.env`** file with your PostgreSQL database credentials and AWS region (if you plan to use Lambda jobs).

---

## 📁 Project Structure

The project is organized into logical modules for easy navigation and understanding:

- `main.go`: The primary entry point for the scheduler server.
- `api/`: Contains all HTTP handlers and defines API routes.
- `job/`: Defines job structures and types.
- `scheduler/`: Manages cron logic and job scheduling.
- `runner/`: Executes different job types (Shell, HTTP, Lambda), also has some example request commands.
- `database/`: Handles PostgreSQL connections and database queries.
- `logs/`: Provides logging utilities and other helper functions.

---

## 🔧 Usage

Once the scheduler is running (typically on port `3000`), you can interact with it using its RESTful API:

- **Add a new job:**
  `POST /jobs/`

- **View all jobs:**
  `POST /alljobs/`

- **Modify an existing job:**
  `PUT /update/{job-name}`

- **Remove a job:**
  `POST /delete/{job-name}`

**Important Notes:**

- The default API port is `3000`.
- **Shell jobs** execute within the host operating system environment where the scheduler is running.
- **HTTP jobs** require a reachable URL.
- **Lambda jobs** need a valid AWS Lambda ARN and a JSON-formatted command as their payload.
- **Sample commands** are available in the `runner` folder.

---

## ☁️ Deployment

The scheduler is deployed on an AWS EC2 instance with PostgreSQL hosted on AWS RDS and optional AWS Lambda integration.

- **EC2**: Hosts the Go server
- **RDS**: Stores persistent job and job_run data
- **Lambda**: Allows serverless job execution via ARN

Make sure the EC2 instance has an IAM role attached or AWS credentials configured to allow Lambda invocation.

---

## 📬 Real-World Example

I'm using the **Go Job Scheduler** to automatically trigger a personal AWS Lambda function every morning. This Lambda function generates a **random fantasy story** and sends it to my email inbox (`uppuluridivyanth@gmail.com`). The job runs daily at **8:30 AM IST**, and a second schedule runs it again at **1:05 PM IST** as a backup.

This showcases how you can use this scheduler to combine serverless execution with creative automation workflows.

---
