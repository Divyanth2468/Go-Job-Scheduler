package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Divyanth2468/go-job-scheduler/internal/logs"
	_ "github.com/lib/pq"
)

var Db *sql.DB

func Init() {
	logs.LogFileSetup()
	dbName := "postgres"
	password := "password"
	connStr := fmt.Sprintf("postgres://postgres:%s@localhost/%s?sslmode=disable", password, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logs.LogAndPrint(err.Error())
		log.Fatal(err.Error(), "\n")
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Error connecting to DB: ", err, "\n")
	}

	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'jobsdb')").Scan(&exists)
	if err != nil {
		log.Fatal("Error checking database existence: ", err, "\n")
	}

	if !exists {
		_, err = db.Exec("CREATE DATABASE jobsdb")
		if err != nil {
			log.Fatal("Failed to create database: ", err, "\n")
		}
		logs.LogAndPrint("Database 'jobsdb' created successfully!\n")
	} else {
		logs.LogAndPrint("Database 'jobsdb' already exists.\n")
	}

	if err = db.Close(); err != nil {
		logs.LogAndPrint("Error closing: %v\n", err.Error())
	}

	dbName = "jobsdb"
	connStr = fmt.Sprintf("postgres://postgres:%s@localhost/%s?sslmode=disable", password, dbName)
	Db, err = sql.Open("postgres", connStr)
	if err != nil {
		logs.LogAndPrint(err.Error() + "\n")
	}
	if _, err = Db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`); err != nil {
		logs.LogAndPrint("UUID extention not being created %v\n", err)
	}
	if _, err = Db.Exec(`CREATE TABLE IF NOT EXISTS jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT UNIQUE NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('http', 'shell', 'lambda')),
  cron_expr TEXT NOT NULL,
	lambda_arn TEXT,
  command TEXT NOT NULL,
  retries INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`); err != nil {
		// log.Panic("Jobs Table not created")
		logs.LogAndPrint("Jobs Table not created %v\n", err)
	} else {
		logs.LogAndPrint("Jobs Table created or already exists.\n")
	}

	if _, err = Db.Exec(`CREATE TABLE IF NOT EXISTS job_runs (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		job_id UUID REFERENCES jobs(id),
		job_name TEXT NOT NULL,
		job_type TEXT NOT NULL CHECK (job_type IN ('http', 'shell', 'lambda')),
		status TEXT CHECK (status IN ('success', 'failure')) NOT NULL,
		log TEXT,
		run_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`); err != nil {
		logs.LogAndPrint("Job runs Table not created %v\n", err)
	} else {
		logs.LogAndPrint("Job runs table created or already exists.\n")
	}
}
