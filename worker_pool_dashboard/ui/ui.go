package ui

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nishith/mini_projects/worker_pool_dashboard/store"
	"github.com/nishith/mini_projects/worker_pool_dashboard/task"
)

type Tick struct{}
type AppModel struct {
	Queue       *task.TaskQueue
	workerTable table.Model
	jobsTable   table.Model
	DB          *sql.DB
	ctx         context.Context
	// wErr		chan string
	// jErr		chan string
}

type AllJobsDone struct{}

func NewAppModel(tq *task.TaskQueue, db *sql.DB, ctx context.Context) *AppModel {
	q := tq
	c1 := []table.Column{
		{Title: "Worker ID", Width: 20},
		{Title: "JobID", Width: 20},
		{Title: "Status", Width: 20},
	}
	c2 := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "URL", Width: 20},
		{Title: "STATUS", Width: 20},
		{Title: "CODE", Width: 20},
		{Title: "FINISHED", Width: 20},
	}
	t1 := table.New(table.WithColumns(c1), table.WithHeight(20))
	t2 := table.New(table.WithColumns(c2), table.WithHeight(20))
	r1 := make([]table.Row, tq.WorkerCount)
	for i := range tq.WorkersStatus {
		r1[i] = table.Row{fmt.Sprintf("Worker %d", i), fmt.Sprintf("%d", tq.WorkersStatus[i].JobID), tq.WorkersStatus[i].JobStatus}
	}
	t1.SetRows(r1)
	s1 := table.DefaultStyles()
	s2 := table.DefaultStyles()
	s1.Header = s1.Header.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.White)
	s2.Header = s2.Header.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.White)
	s1.Cell = s1.Cell.Foreground(lipgloss.White)
	s2.Cell = s2.Cell.Foreground(lipgloss.White)
	s1.Selected = s1.Selected.Background(lipgloss.Magenta)
	s2.Selected = s2.Selected.Background(lipgloss.Magenta)
	t1.SetStyles(s1)
	t2.SetStyles(s2)
	return &AppModel{q, t1, t2, db, ctx}
}

func (m *AppModel) Init() tea.Cmd {
	m.workerTable.Focus()
	go m.Queue.StartWorkers(m.DB)
	return tea.Batch(m.tick(), m.allJobsDone())
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.jobsTable.SetWidth(msg.Width)
		m.workerTable.SetWidth(msg.Width)
		m.jobsTable.SetHeight(msg.Height / 4)
		m.workerTable.SetHeight(msg.Height / 2)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.workerTable.Focused() {
				m.jobsTable.Focus()
			} else {
				m.workerTable.Focus()
			}
		}
	case AllJobsDone:
		return m, tea.Quit
	case Tick:
		m.UpdateJobRows()
		m.UpdateWorkerRows()
		cmds = append(cmds, m.tick())
	}
	m.workerTable, cmd = m.workerTable.Update(msg)
	cmds = append(cmds, cmd)
	m.jobsTable, cmd = m.jobsTable.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *AppModel) tick() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return Tick{}
	})
}

func (m *AppModel) allJobsDone() tea.Cmd {
	return func() tea.Msg {
		<-m.Queue.Done
		return AllJobsDone{}
	}
}
func (m *AppModel) UpdateJobRows() {
	jobs, err := store.GetJobStatus(m.ctx, m.DB)
	if err != nil {
		return
	}
	rows := make([]table.Row, len(jobs))
	for i, j := range jobs {
		var finished string
		if j.FinishedAt != nil {
			finished = j.FinishedAt.Format(time.DateTime)
		} else {
			finished = "NA"
		}
		rows[i] = table.Row{
			strconv.Itoa(j.ID),
			j.URL,
			j.Status,
			strconv.Itoa(j.StatusCode),
			finished,
		}
	}
	m.jobsTable.SetRows(rows)
}

func (m *AppModel) UpdateWorkerRows() {
	m.Queue.Mux.RLock()
	defer m.Queue.Mux.RUnlock()
	rows := make([]table.Row, m.Queue.WorkerCount)
	for w := range m.Queue.WorkersStatus {
		rows[w] = table.Row{
			fmt.Sprintf("Worker %d", w),
			fmt.Sprintf("Job Id %d", m.Queue.WorkersStatus[w].JobID),
			m.Queue.WorkersStatus[w].JobStatus,
		}
	}
	m.workerTable.SetRows(rows)
}
func (m *AppModel) View() tea.View {
	c := lipgloss.JoinVertical(lipgloss.Left, m.workerTable.View(), m.jobsTable.View())
	v := tea.NewView(c)
	v.AltScreen = true
	return v
}
