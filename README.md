# ⏰ Go Job Scheduler

A cloud-native, backend-only job scheduler written in Go. It supports scheduling shell and HTTP tasks using cron expressions and is designed for simplicity, flexibility, and easy extensibility.

---

## 🚀 Features

- **Schedule shell or HTTP jobs** using cron expressions
- **Retry failed jobs** a configurable number of times
- **Lightweight, serverless-ready architecture**
- **RESTful API** to manage job lifecycle (add, update, delete)
- **Easily extensible** to add persistence or additional job types

---

## 🛠️ Tech Stack

- **Language:** Go
- **Architecture:** Serverless-ready, modular backend
- **Job Types:** Shell & HTTP

---

## 📦 Installation

### 1. Clone the repository

```bash
git clone https://github.com/Divyanth2468/go-job-scheduler.git
cd go-job-scheduler
```

### 2. Run the Scheduler

```bash
go run main.go
```

✅ Ensure **Go** is installed and available in your system path (`go version` to verify).

---

## 📁 Project Structure

- `main.go` – entry point for the server
- `job/` – job definitions and structs
- `scheduler/` – cron logic and cron job additions
- `api/` – HTTP handlers and routes
- `runner/` – sample curl commands and executable jobs

---

## 🔧 Usage

Once the scheduler is running:

- Use `POST /jobs/` to add jobs
- Use `POST /alljobs/` to view all jobs
- Use `PUT /update/{job-name}` to modify jobs
- Use `POST /delete/{job-name}` to remove jobs

---

## 📌 Notes

- Default API port: `3000`
- Accepts cron expressions like `@every 30s`, `@hourly`, `@daily`
- Shell commands run on the host OS environment
- HTTP jobs expect a valid reachable URL
