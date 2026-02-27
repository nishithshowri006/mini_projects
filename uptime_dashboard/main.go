package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/goccy/go-yaml"
	"github.com/nishithshowri006/mini_projects/url_dashboard/pinger"
	"github.com/nishithshowri006/mini_projects/url_dashboard/ui"
)

func ParseConfig(fp io.Reader) (*pinger.URLMetadata, error) {
	var intervals pinger.IntervalList
	decoder := yaml.NewDecoder(fp)
	if err := decoder.Decode(&intervals); err != nil {
		return nil, err
	}
	return &intervals.URLMetadata, nil
}
func main() {
	fileName := flag.String("fileName", "example.yaml", "Enter the yaml file to be parsed")
	flag.Parse()
	fp, err := os.Open(*fileName)
	if err != nil {
		log.Fatal(err)
	}

	meta, err := ParseConfig(fp)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

	go func() {
		<-sigch
		cancel()
	}()
	app, err := ui.NewApp(ctx, fp, meta)
	if err != nil {
		log.Fatal(err)
	}
	p := tea.NewProgram(app, tea.WithContext(ctx), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		if errors.Is(err, tea.ErrProgramKilled) {
			return
		} else {
			log.Fatal(err)
		}
	}
}
