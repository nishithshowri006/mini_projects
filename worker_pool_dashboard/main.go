package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/nishith/mini_projects/worker_pool_dashboard/store"
	"github.com/nishith/mini_projects/worker_pool_dashboard/task"
	"github.com/nishith/mini_projects/worker_pool_dashboard/ui"
)

type JobList struct {
	Jobs []struct {
		URL string `json:"url"`
	} `json:"jobs"`
}

func ParseJson(fp io.Reader) ([]string, error) {
	var jl JobList
	if err := json.NewDecoder(fp).Decode(&jl); err != nil {
		return nil, err
	}
	urls := make([]string, len(jl.Jobs))
	for i := range urls {
		urls[i] = jl.Jobs[i].URL
	}
	return urls, nil
}

func main() {
	fp, err := os.Open("jobs.json")
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	urls, err := ParseJson(fp)
	if err != nil {
		log.Fatal(err)
	}
	db, err := store.NewStorage("example.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := store.InsertJob(ctx, db, urls); err != nil {
		log.Fatal(err)
	}
	tq := task.NewTaskQueue(ctx, 2)
	p := tea.NewProgram(ui.NewAppModel(tq, db, ctx))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	if err := store.DeleteJobs(ctx, db); err != nil {
		log.Fatal(err)
	}
	fmt.Println("All jobs have been executed")
}
