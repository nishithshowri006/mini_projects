package store

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
}

type Job struct {
	ID         int
	URL        string
	Status     string
	CreatedAt  time.Time
	StartedAt  *time.Time // pointer because it can be NULL
	FinishedAt *time.Time
	StatusCode int
	Error      string
}

func InsertJob(ctx context.Context, db *sql.DB, urls []string) error {
	query := `
	INSERT INTO jobs (url)
	VALUES (?)
	`
	statement, err := db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer statement.Close()
	for _, url := range urls {
		_, err = statement.ExecContext(ctx, url)
		if err != nil {
			return err
		}
	}
	return nil
}

func ClaimJob(ctx context.Context, db *sql.DB) (*Job, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	query := `
	UPDATE jobs SET status='processing', started_at=CURRENT_TIMESTAMP
	WHERE id = (
    SELECT id FROM jobs WHERE status='pending' ORDER BY created_at LIMIT 1
	)
	RETURNING id;
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx)
	var id int
	err = row.Scan(&id)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	row = db.QueryRowContext(ctx, "select id,url,status, created_at from jobs where id = ?", id)
	j := new(Job)
	if err := row.Scan(&j.ID, &j.URL, &j.Status, &j.CreatedAt); err != nil {
		return nil, err
	}
	return j, nil
}

func CompleteJob(ctx context.Context, db *sql.DB, id int, statusCode int, errMsg string) error {
	var status string
	if errMsg == "" {
		status = "completed"
	} else {
		status = "failed"
	}
	query := `
	Update jobs
	SET status = ?, finished_at = ?, status_code = ?, error = ?
	where id = ?
	`
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	smt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	_, err = smt.ExecContext(ctx, status, time.Now(), statusCode, errMsg, id)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func GetJobStatus(ctx context.Context, db *sql.DB) ([]Job, error) {
	jobs := make([]Job, 0)
	statement := `
	SELECT id, url, status, status_code, finished_at
	from jobs
	where status in ('failed','completed')
	order by finished_at desc
	`
	rows, err := db.QueryContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.URL, &j.Status, &j.StatusCode, &j.FinishedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func DeleteJobs(ctx context.Context, db *sql.DB) error {
	statement := `DELETE FROM jobs;`
	_, err := db.ExecContext(ctx, statement)
	if err != nil {
		return err
	}
	return nil
}

func NewStorage(fileName string) (*sql.DB, error) {
	statement := `
		CREATE TABLE IF NOT EXISTS jobs  (
	    id          INTEGER PRIMARY KEY AUTOINCREMENT,
	    url         TEXT NOT NULL,
	    status      TEXT NOT NULL DEFAULT 'pending',
	    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
	    started_at  DATETIME,
	    finished_at DATETIME,
	    status_code INTEGER,
	    error       TEXT
	);`

	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(statement); err != nil {
		return nil, err
	}
	return db, nil
}
