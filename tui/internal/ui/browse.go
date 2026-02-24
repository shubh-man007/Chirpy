package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"github.com/shubh-man007/Chirpy/tui/internal/api"
	"github.com/shubh-man007/Chirpy/tui/internal/models"
)

type BrowseModel struct {
	client *api.Chirpy

	width  int
	height int

	input    textinput.Model
	viewport viewport.Model

	currentUserID string
	profile       *models.ProfileResponse

	loading  bool
	errorMsg string
	spin     spinner.Model
}

func NewBrowseModel(client *api.Chirpy) BrowseModel {
	ti := textinput.New()
	ti.Placeholder = "Enter user ID"
	ti.Prompt = "User ID: "
	ti.Focus()

	vp := viewport.New(0, 0)
	s := spinner.New()
	s.Spinner = spinner.Dot

	return BrowseModel{
		client:   client,
		input:    ti,
		viewport: vp,
		spin:     s,
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
			m.profile = nil
			return m, tea.Batch(m.spin.Tick, fetchUserProfileCmd(m.client, userID))
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

	case FollowUnfollowSuccessMsg:
		if m.currentUserID != "" {
			m.loading = true
			return m, tea.Batch(m.spin.Tick, fetchUserProfileCmd(m.client, m.currentUserID))
		}
		return m, nil

	case ProfileLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
			m.profile = nil
		} else {
			m.errorMsg = ""
			m.profile = msg.Profile
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
	if m.loading {
		m.spin, cmd = m.spin.Update(msg)
		cmds = append(cmds, cmd)
	}
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
				Render(m.spin.View()+" Loading profile..."),
		)
	} else if m.profile == nil {
		lines = append(lines,
			lipgloss.NewStyle().Foreground(colorMuted).
				Render("Enter a user ID and press Enter to view their profile."),
		)
	} else {
		status := "Standard"
		if m.profile.IsChirpyRed {
			status = "Chirpy Red"
		}
		info := lipgloss.JoinVertical(lipgloss.Left,
			"Email: "+m.profile.Email,
			"User ID: "+m.profile.ID,
			"Status: "+status,
			fmt.Sprintf("Followers: %d  Following: %d  Chirps: %d",
				m.profile.FollowersCount, m.profile.FollowingCount, m.profile.ChirpsCount),
		)
		if m.profile.IsFollowing != nil {
			if *m.profile.IsFollowing {
				info = lipgloss.JoinVertical(lipgloss.Left, info, authorStyle.Render("You follow this user"))
			} else {
				info = lipgloss.JoinVertical(lipgloss.Left, info, lipgloss.NewStyle().Foreground(colorMuted).Render("Not following"))
			}
		}
		lines = append(lines, chirpBoxStyle.Width(m.viewport.Width).Render(info))
		lines = append(lines, "")

		if len(m.profile.Chirps) == 0 {
			lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render("No chirps yet."))
		} else {
			lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Chirps:"))
			for _, c := range m.profile.Chirps {
				header := fmt.Sprintf("%s", relativeTime(c.CreatedAt))
				body := wordwrap.String(c.Body, m.viewport.Width-4)
				content := lipgloss.JoinVertical(
					lipgloss.Left,
					timestampStyle.Render(header),
					body,
				)
				lines = append(lines, chirpBoxStyle.Width(m.viewport.Width).Render(content))
			}
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

	now := time.Now().Format("2006-01-02 15:04")
	left := "[Enter] Load profile  [f] Follow  [u] Unfollow  [q] Quit"
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

func fetchUserProfileCmd(client *api.Chirpy, userID string) tea.Cmd {
	return func() tea.Msg {
		profile, err := client.GetProfile(userID, "", 20)
		return ProfileLoadedMsg{Profile: profile, Err: err}
	}
}

func followUserCmd(client *api.Chirpy, userID string) tea.Cmd {
	return func() tea.Msg {
		err := client.FollowUser(userID)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return FollowUnfollowSuccessMsg{UserID: userID}
	}
}

func unfollowUserCmd(client *api.Chirpy, userID string) tea.Cmd {
	return func() tea.Msg {
		err := client.UnfollowUser(userID)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return FollowUnfollowSuccessMsg{UserID: userID}
	}
}
