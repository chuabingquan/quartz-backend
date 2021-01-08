package postgres

import (
	"database/sql"
	"fmt"
	"quartz"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// JobService ...
type JobService struct {
	DB *sqlx.DB
}

// Job ...
func (js *JobService) Job(jobID string) (quartz.Job, error) {
	query := `SELECT * FROM jobs WHERE id=$1`
	job := quartz.Job{}

	err := js.DB.QueryRowx(query, jobID).StructScan(&job)
	if err == sql.ErrNoRows {
		return quartz.Job{}, quartz.ErrEntityNotFound
	} else if err != nil {
		return quartz.Job{}, fmt.Errorf("Failed to query job from database: %w", err)
	}

	schedule, err := js.getJobSchedule(jobID)
	if err != nil {
		return quartz.Job{}, err
	}

	job.Schedule = schedule
	return job, nil
}

// Jobs ...
func (js *JobService) Jobs() ([]quartz.Job, error) {
	query := `SELECT * FROM jobs`

	rows, err := js.DB.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("Failed to query jobs from database: %w", err)
	}
	defer rows.Close()

	jobs := []quartz.Job{}
	for rows.Next() {
		job := quartz.Job{}

		err := rows.StructScan(&job)
		if err != nil {
			return nil, fmt.Errorf("Failed to scan job from row to struct: %w", err)
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("An error occurred with rows when querying for jobs from database: %w", err)
	}

	for _, job := range jobs {
		schedule, err := js.getJobSchedule(job.ID)
		if err != nil {
			return nil, err
		}
		job.Schedule = schedule
	}

	return jobs, nil
}

// CreateJob ...
func (js *JobService) CreateJob(j quartz.Job) (string, error) {
	tx, err := js.DB.Beginx()
	if err != nil {
		return "", fmt.Errorf("Failed to begin transaction to insert job into database: %w", err)
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			panic(fmt.Errorf("Failed to rollback create job transaction: %w", err))
		}
	}()

	jobID := uuid.New().String()
	j.ID = jobID

	insertJobQuery := `INSERT INTO jobs(id, name, timezone, container_id) VALUES(:id, :name, :timezone, :container_id)`
	_, err = tx.NamedExec(insertJobQuery, &j)
	if err != nil {
		return "", fmt.Errorf("Failed to insert new job into database: %w", err)
	}

	insertScheduleQuery := `INSERT INTO schedule(id, expression, job_id) VALUES($1, $2, $3)`
	for _, cron := range j.Schedule {
		_, err := tx.Exec(insertScheduleQuery, uuid.New().String(), cron.Expression, jobID)
		if err != nil {
			return "", fmt.Errorf("Failed to insert job schedule into database: %w", err)
		}
	}

	tx.Commit()

	return jobID, nil
}

// DeleteJob ...
func (js *JobService) DeleteJob(jobID string) error {
	query := `DELETE * FROM jobs WHERE id=$1`

	res, err := js.DB.Exec(query, jobID)
	if err != nil {
		return fmt.Errorf("Failed to delete job from database: %w", err)
	}

	if rows, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("Failed to get rows affected after deleting job from database: %w", err)
	} else if rows < 1 {
		return quartz.ErrEntityNotFound
	}

	return nil
}

func (js *JobService) getJobSchedule(jobID string) ([]quartz.Cron, error) {
	query := `SELECT * FROM schedule WHERE job_id=$1`

	rows, err := js.DB.Queryx(query, jobID)
	if err != nil {
		return nil, fmt.Errorf("Failed to query job schedule from database: %w", err)
	}
	defer rows.Close()

	schedule := []quartz.Cron{}
	for rows.Next() {
		cron := quartz.Cron{}

		err := rows.StructScan(&cron)
		if err != nil {
			return nil, fmt.Errorf("Failed to scan job schedule from row to struct: %w", err)
		}

		schedule = append(schedule, cron)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("An error occurred with rows when querying for job schedule from database: %w", err)
	}

	return schedule, nil
}
