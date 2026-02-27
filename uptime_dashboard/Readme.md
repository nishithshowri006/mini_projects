# Uptime Dashboard

A real-time CLI tool to monitor website health and uptime using a Terminal User Interface (TUI). Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Concurrent Monitoring**: Uses goroutines to monitor multiple URLs simultaneously.
- **Customizable Intervals**: Set check intervals per URL in the configuration file.
- **TUI Display**: Real-time status updates in a clean terminal interface.

## Installation

1. Clone the repository.
2. Navigate to the `uptime_dashboard` directory.
3. Install dependencies:
   ```bash
   go mod download
   ```

## Usage

1. Create a configuration file (e.g., `example.yaml`) with the list of URLs and intervals:

```yaml
urls:
  - url: "https://google.com"
    interval: 5s
  - url: "https://github.com"
    interval: 10s
```

2. Run the application:
   ```bash
   go run main.go -fileName example.yaml
   ```