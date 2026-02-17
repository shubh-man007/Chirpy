package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/reflow/wordwrap"

	"github.com/shubh-man007/Chirpy/tui/internal/api"
	"github.com/shubh-man007/Chirpy/tui/internal/models"
)

type FeedModel struct {
	client *api.Chirpy

	width  int
	height int

	viewport viewport.Model
	chirps   []models.Chirp

	cursor  int
	loading bool
	hasMore bool

	limit  int
	offset int

	errorMsg string
	spin     spinner.Model
}

func NewFeedModel(client *api.Chirpy) FeedModel {
	vp := viewport.New(0, 0)
	vp.SetYOffset(0)

	s := spinner.New()
	s.Spinner = spinner.Dot

	return FeedModel{
		client:   client,
		viewport: vp,
		cursor:   0,
		limit:    20,
		offset:   0,
		hasMore:  true,
		spin:     s,
	}
}

func (m *FeedModel) InitFeed() tea.Cmd {
	m.loading = true
	m.errorMsg = ""
	m.chirps = nil
	m.cursor = 0
	m.offset = 0
	m.hasMore = true

	return tea.Batch(
		m.spin.Tick,
		fetchFeedCmd(m.client, m.limit, m.offset, false),
	)
}

func (m FeedModel) Update(msg tea.Msg) (FeedModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalculateViewport()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.chirps)-1 {
				m.cursor++
			}
		case "r":
			// Refresh feed.
			m.offset = 0
			m.hasMore = true
			m.chirps = nil
			m.cursor = 0
			m.loading = true
			return m, tea.Batch(
				m.spin.Tick,
				fetchFeedCmd(m.client, m.limit, m.offset, false),
			)
		}

	case FeedLoadedMsg:
		m.loading = false
		if msg.Append {
			m.chirps = append(m.chirps, msg.Chirps...)
		} else {
			m.chirps = msg.Chirps
		}

		if len(msg.Chirps) < m.limit {
			m.hasMore = false
		}

		m.buildViewportContent()
		return m, nil

	case ErrorMsg:
		m.loading = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
		}
		m.buildViewportContent()
		return m, nil
	}

	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	if m.loading {
		m.spin, cmd = m.spin.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.viewport.AtBottom() && !m.loading && m.hasMore {
		m.loading = true
		m.offset += m.limit
		cmds = append(cmds,
			tea.Batch(
				m.spin.Tick,
				fetchFeedCmd(m.client, m.limit, m.offset, true),
			),
		)
	}

	return m, tea.Batch(cmds...)
}

func (m *FeedModel) recalculateViewport() {
	if m.width == 0 || m.height == 0 {
		return
	}

	header := headerStyle.Width(m.width).Render(" Chirpy | Feed ")
	footer := footerStyle.Width(m.width).Render("[↑/k ↓/j] Navigate  [r] Refresh  [q] Quit")

	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if contentHeight < 1 {
		contentHeight = 1
	}

	m.viewport.Width = m.width - 4
	if m.viewport.Width < 20 {
		m.viewport.Width = m.width
	}
	m.viewport.Height = contentHeight

	m.buildViewportContent()
}

func (m *FeedModel) buildViewportContent() {
	var lines []string

	if m.loading && len(m.chirps) == 0 {
		loadingLine := lipgloss.NewStyle().
			Foreground(colorMuted).
			Render(m.spin.View() + " Loading feed...")
		lines = append(lines, loadingLine)
	} else {
		if len(m.chirps) == 0 {
			empty := lipgloss.NewStyle().
				Foreground(colorMuted).
				Render("Your feed is empty.\n\n" +
					"- Press 'c' to compose your first chirp.\n" +
					"- Press '2' to view your profile.\n" +
					"- Press 'b' to browse users by ID and follow them.")
			lines = append(lines, empty)
		} else {
			for i, c := range m.chirps {
				lines = append(lines, m.renderChirp(c, i == m.cursor))
			}

			if m.hasMore {
				lines = append(lines, "")
				lines = append(lines, lipgloss.NewStyle().
					Foreground(colorMuted).
					Render("↓ More chirps loading as you scroll ↓"))
			}
		}
	}

	if m.errorMsg != "" {
		lines = append([]string{errorStyle.Render("⚠ " + m.errorMsg), ""}, lines...)
	}

	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m *FeedModel) renderChirp(c models.Chirp, selected bool) string {
	header := fmt.Sprintf("%s · %s", truncate.StringWithTail(c.UserID, 16, "…"), relativeTime(c.CreatedAt))
	body := wordwrap.String(c.Body, m.viewport.Width-4)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		authorStyle.Render(header),
		body,
	)

	if selected {
		return selectedChirpStyle.Width(m.viewport.Width).Render(content)
	}

	return chirpBoxStyle.Width(m.viewport.Width).Render(content)
}

func (m FeedModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading feed..."
	}

	header := headerStyle.Width(m.width).Render(" Chirpy | Feed ")

	now := time.Now().Format("2006-01-02 15:04")
	left := "[↑/k ↓/j] Navigate  [r] Refresh  [q] Quit"
	space := ""
	totalWidth := lipgloss.Width(left + now)
	if m.width > totalWidth {
		space = strings.Repeat(" ", m.width-totalWidth)
	}
	footer := footerStyle.Width(m.width).Render(left + space + now)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.viewport.View(),
		footer,
	)
}

func fetchFeedCmd(client *api.Chirpy, limit, offset int, append bool) tea.Cmd {
	return func() tea.Msg {
		chirps, err := client.GetFeed(limit, offset)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return FeedLoadedMsg{
			Chirps: chirps,
			Append: append,
		}
	}
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
