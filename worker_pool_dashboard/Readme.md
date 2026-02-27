# Worker Pool Dashboard

A CLI dashboard to visualize a worker pool processing jobs from a SQLite database. This project demonstrates concurrent task execution and real-time TUI updates.

## Features

- **SQLite Integration**: Persistent job storage using SQLite.
- **Worker Pool**: Efficiently process multiple jobs concurrently.
- **TUI Progress**: Real-time visualization of job statuses and worker activity.
- **JSON Input**: Load initial jobs from a simple JSON file.

## Prerequisites

- Go installed on your system.
- CGO enabled (required for `github.com/mattn/go-sqlite3`).

## Installation

1. Clone the repository.
2. Navigate to the `worker_pool_dashboard` directory.
3. Install dependencies:
   ```bash
   go mod download
   ```

## Usage

1. Ensure `jobs.json` exists in the project root with the list of URLs to process. Example:

```json
{
  "jobs": [
    { "url": "https://google.com" },
    { "url": "https://github.com" },
    { "url": "https://go.dev" }
  ]
}
```

2. Run the application:
   ```bash
   go run main.go
   ```

## Architecture

- **Store**: Handles SQLite database operations (Insert, Fetch, Delete).
- **Task**: Manages the worker pool and task queue.
- **UI**: Bubble Tea based TUI for real-time visualization.
