package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"github.com/shubh-man007/Chirpy/tui/internal/api"
	"github.com/shubh-man007/Chirpy/tui/internal/models"
)

// BrowseModel lets the user look up another user's chirps by user ID
// and follow/unfollow them.
type BrowseModel struct {
	client *api.Chirpy

	width  int
	height int

	input   textinput.Model
	viewport viewport.Model

	currentUserID string
	chirps        []models.Chirp

	loading  bool
	errorMsg string
}

func NewBrowseModel(client *api.Chirpy) BrowseModel {
	ti := textinput.New()
	ti.Placeholder = "Enter user ID"
	ti.Prompt = "User ID: "
	ti.Focus()

	vp := viewport.New(0, 0)

	return BrowseModel{
		client:   client,
		input:    ti,
		viewport: vp,
	}
}

func (m BrowseModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m BrowseModel) Update(msg tea.Msg) (BrowseModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalculateViewport()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			userID := strings.TrimSpace(m.input.Value())
			if userID == "" {
				m.errorMsg = "User ID is required."
				return m, nil
			}
			m.currentUserID = userID
			m.loading = true
			m.errorMsg = ""
			return m, fetchUserChirpsCmd(m.client, userID)
		case "f":
			if m.currentUserID == "" {
				return m, nil
			}
			return m, followUserCmd(m.client, m.currentUserID)
		case "u":
			if m.currentUserID == "" {
				return m, nil
			}
			return m, unfollowUserCmd(m.client, m.currentUserID)
		}

	case UserChirpsLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
			m.chirps = nil
		} else {
			m.errorMsg = ""
			m.chirps = msg.Chirps
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
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *BrowseModel) recalculateViewport() {
	if m.width == 0 || m.height == 0 {
		return
	}

	header := headerStyle.Width(m.width).Render(" Chirpy | Browse Users ")
	footer := footerStyle.Width(m.width).Render("")

	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 3
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

func (m *BrowseModel) buildViewportContent() {
	var lines []string

	if m.loading {
		lines = append(lines,
			lipgloss.NewStyle().Foreground(colorMuted).
				Render("Loading user chirps..."),
		)
	} else if len(m.chirps) == 0 {
		lines = append(lines,
			lipgloss.NewStyle().Foreground(colorMuted).
				Render("No chirps found for this user."),
		)
	} else {
		for _, c := range m.chirps {
			header := fmt.Sprintf("%s · %s", c.UserID, relativeTime(c.CreatedAt))
			body := wordwrap.String(c.Body, m.viewport.Width-4)
			content := lipgloss.JoinVertical(
				lipgloss.Left,
				authorStyle.Render(header),
				body,
			)
			lines = append(lines, chirpBoxStyle.Width(m.viewport.Width).Render(content))
		}
	}

	if m.errorMsg != "" {
		lines = append([]string{errorStyle.Render("⚠ " + m.errorMsg), ""}, lines...)
	}

	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m BrowseModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Browse"
	}

	header := headerStyle.Width(m.width).Render(" Chirpy | Browse Users ")

	var bodyLines []string
	bodyLines = append(bodyLines, m.input.View(), "")
	bodyLines = append(bodyLines, m.viewport.View())

	body := contentStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, bodyLines...),
	)

	// Footer with current time at the bottom-right.
	now := time.Now().Format("2006-01-02 15:04")
	left := "[Enter] Load chirps  [f] Follow  [u] Unfollow  [q] Quit"
	space := ""
	totalWidth := lipgloss.Width(left + now)
	if m.width > totalWidth {
		space = strings.Repeat(" ", m.width-totalWidth)
	}
	footer := footerStyle.Width(m.width).Render(left + space + now)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
		footer,
	)
}

func fetchUserChirpsCmd(client *api.Chirpy, userID string) tea.Cmd {
	return func() tea.Msg {
		chirps, err := client.GetChirpsByUser(userID, "desc")
		return UserChirpsLoadedMsg{Chirps: chirps, Err: err}
	}
}

func followUserCmd(client *api.Chirpy, userID string) tea.Cmd {
	return func() tea.Msg {
		err := client.FollowUser(userID)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return nil
	}
}

func unfollowUserCmd(client *api.Chirpy, userID string) tea.Cmd {
	return func() tea.Msg {
		err := client.UnfollowUser(userID)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return nil
	}
}

