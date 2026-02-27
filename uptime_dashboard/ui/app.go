package ui

import (
	"context"
	"io"
	"maps"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nishithshowri006/mini_projects/url_dashboard/pinger"
)

type AppModel struct {
	table    table.Model
	UrlMap   map[string]pinger.PingResponse
	Pinger   *pinger.Pinger
	ctx      context.Context
	urlOrder []string
}

func NewApp(ctx context.Context, fp io.Reader, meta *pinger.URLMetadata) (*AppModel, error) {
	columns := []table.Column{
		{Title: "URL", Width: 60},
		{Title: "STATUS", Width: 6},
		{Title: "RT", Width: 12},
		{Title: "LAST UPDATED", Width: 30},
	}
	t := table.New(table.WithColumns(columns))
	s := table.DefaultStyles()
	s.Header = s.Header.
		PaddingTop(2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	m := make(map[string]pinger.PingResponse)
	buffCount := len(meta.URLs)
	p := pinger.NewPinger(buffCount, time.Second*time.Duration(meta.Interval))
	order := make([]string, 0)
	for _, url := range meta.URLs {
		m[url] = pinger.PingResponse{}
		order = append(order, url)
	}
	a := AppModel{
		table:    t,
		UrlMap:   m,
		Pinger:   p,
		ctx:      ctx,
		urlOrder: order,
	}
	return &a, nil
}

func (a *AppModel) Init() tea.Cmd {
	go a.Pinger.StartLoop(a.ctx, maps.Keys(a.UrlMap))
	return a.tick
}

func (a *AppModel) tick() tea.Msg {
	return tea.Msg(<-a.Pinger.Pings)
}

func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.table.SetWidth(msg.Width)
		a.table.SetHeight(msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String(), "q":
			return a, tea.Quit
		}
	case pinger.PingResponse:
		a.UpdateRows(msg)
	}
	a.table, cmd = a.table.Update(msg)
	return a, tea.Batch(cmd, a.tick)
}

func (a *AppModel) UpdateRows(resp pinger.PingResponse) {
	a.UrlMap[resp.URL] = resp
	rows := make([]table.Row, len(a.UrlMap))
	i := 0
	for _, u := range a.urlOrder {
		url, ok := a.UrlMap[u]
		if !ok {
			continue
		}
		rows[i] = table.Row{
			url.URL,
			strconv.Itoa(url.StatusCode),
			url.ResponseTime.Truncate(time.Millisecond).String(),
			url.PingedAt.Format(time.DateTime),
		}
		i++
	}
	a.table.SetRows(rows)
}

func (a *AppModel) View() string {
	return a.table.View()
}
