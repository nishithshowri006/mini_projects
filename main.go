package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nishith/mini_projects/url_dashboard/ui"
)

func main() {
	fileName := flag.String("fileName", "example.yaml", "Enter the yaml file to be parsed")
	flag.Parse()
	fp, err := os.Open(*fileName)
	if err != nil {
		log.Fatalln(err)
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
	app, err := ui.NewApp(ctx, fp, 10)
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
