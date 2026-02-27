package task

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/nishith/mini_projects/worker_pool_dashboard/store"
)

type WorkerStatus struct {
	JobID     int
	JobStatus string
}
type TaskQueue struct {
	WorkerCount   int
	ctx           context.Context
	Mux           *sync.RWMutex
	WorkersStatus map[int]WorkerStatus
	Done          chan struct{}
}

const (
	STARTING    = "STARTING"
	FINISHED    = "COMPLETED"
	NOJOBS      = "NO JOBS AVAILABLE"
	PROCESSING  = "PROCESSING"
	FAILED      = "FAILED"
	WORKERERROR = "ERROR IN WORKER"
)

func (t *TaskQueue) StartWorkers(db *sql.DB) {
	var wg sync.WaitGroup
	for i := range t.WorkerCount {
		t.WorkersStatus[i] = WorkerStatus{JobStatus: STARTING}
		wg.Go(func() {
			t.Worker(i, db)
		})
	}
	wg.Wait()
	t.Done <- struct{}{}
}

func ping(ctx context.Context, url string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return -1, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return -1, err
		}
		return resp.StatusCode, errors.New(string(body))
	}
	return resp.StatusCode, nil
}

func (t *TaskQueue) Worker(id int, db *sql.DB) {
jobLoop:
	for {
		select {
		case <-t.ctx.Done():
			break jobLoop
		default:
			job, err := store.ClaimJob(t.ctx, db)
			if errors.Is(err, sql.ErrNoRows) {
				t.Mux.Lock()
				t.WorkersStatus[id] = WorkerStatus{JobID: -1, JobStatus: NOJOBS}
				t.Mux.Unlock()
				break jobLoop
			} else if err != nil {
				t.Mux.Lock()
				t.WorkersStatus[id] = WorkerStatus{JobID: -1, JobStatus: NOJOBS}
				t.Mux.Unlock()
				continue
			}
			t.Mux.Lock()
			t.WorkersStatus[id] = WorkerStatus{JobID: job.ID, JobStatus: PROCESSING}
			t.Mux.Unlock()
			time.Sleep(4 * time.Second)
			status, err := ping(t.ctx, job.URL)
			if err != nil {
				job.Error = err.Error()
				t.Mux.Lock()
				t.WorkersStatus[id] = WorkerStatus{JobID: job.ID, JobStatus: FINISHED}
				t.Mux.Unlock()
			}
			if err := store.CompleteJob(t.ctx, db, job.ID, status, job.Error); err != nil {
				log.Println(err)
				continue
			}
			t.Mux.Lock()
			t.WorkersStatus[id] = WorkerStatus{JobID: job.ID, JobStatus: FINISHED}
			t.Mux.Unlock()
		}

	}
}

func NewTaskQueue(ctx context.Context, wc int) *TaskQueue {
	m := make(map[int]WorkerStatus)
	d := make(chan struct{})
	return &TaskQueue{
		WorkerCount:   wc,
		ctx:           ctx,
		Mux:           &sync.RWMutex{},
		WorkersStatus: m,
		Done:          d,
	}
}
