package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/shubh-man007/Chirpy/tui/internal/api"
	"github.com/shubh-man007/Chirpy/tui/internal/models"
)

// ProfileModel lets the user inspect their basic account information,
// update credentials, and delete the account.
type ProfileModel struct {
	client *api.Chirpy
	user   *models.User

	width  int
	height int

	updating      bool
	deleting      bool
	confirmDelete bool

	emailInput    textinput.Model
	passwordInput textinput.Model

	errorMsg string
	spin     spinner.Model
}

func NewProfileModel(client *api.Chirpy) ProfileModel {
	email := textinput.New()
	email.Placeholder = "new-email@example.com"
	email.Prompt = "New email: "

	password := textinput.New()
	password.Placeholder = "new password"
	password.Prompt = "New password: "
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'

	s := spinner.New()
	s.Spinner = spinner.Dot

	return ProfileModel{
		client:        client,
		emailInput:    email,
		passwordInput: password,
		spin:          s,
	}
}

func (m *ProfileModel) SetUser(u *models.User) {
	m.user = u
}

func (m ProfileModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ProfileModel) Update(msg tea.Msg) (ProfileModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Confirmation for deletion.
		if m.confirmDelete {
			switch msg.String() {
			case "y", "Y":
				if m.user == nil || m.deleting {
					return m, nil
				}
				m.deleting = true
				m.errorMsg = ""
				return m, deleteUserCmd(m.client, m.user.ID)
			case "n", "N", "esc":
				m.confirmDelete = false
				return m, nil
			}
			return m, nil
		}

		if m.updating || m.deleting {
			// Ignore other keys while an operation is in progress.
			return m, nil
		}

		switch msg.String() {
		case "u":
			// Start update flow: focus email input.
			m.updating = true
			m.emailInput.Reset()
			m.passwordInput.Reset()
			m.emailInput.Focus()
			m.passwordInput.Blur()
			m.errorMsg = ""
			return m, nil
		case "d":
			// Ask for deletion confirmation.
			if m.user != nil {
				m.confirmDelete = true
			}
			return m, nil
		case "tab", "shift+tab":
			if m.updating {
				if m.emailInput.Focused() {
					m.emailInput.Blur()
					m.passwordInput.Focus()
				} else {
					m.passwordInput.Blur()
					m.emailInput.Focus()
				}
			}
			return m, nil
		case "enter":
			if m.updating {
				email := strings.TrimSpace(m.emailInput.Value())
				password := m.passwordInput.Value()
				if email == "" || password == "" {
					m.errorMsg = "Email and password are required."
					return m, nil
				}
				m.errorMsg = ""
				return m, updateUserCmd(m.client, email, password)
			}
		}

	case UserUpdatedMsg:
		m.updating = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
			return m, nil
		}
		if m.user != nil && msg.User != nil {
			m.user = msg.User
		}
		return m, nil

	case UserDeletedMsg:
		m.deleting = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
			return m, nil
		}
		// RootModel will interpret this message to reset to login if desired.
		return m, nil

	case ErrorMsg:
		// Generic error surfaced while on this screen.
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
		}
		return m, nil
	}

	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.emailInput, cmd = m.emailInput.Update(msg)
	cmds = append(cmds, cmd)

	m.passwordInput, cmd = m.passwordInput.Update(msg)
	cmds = append(cmds, cmd)

	if m.updating || m.deleting {
		m.spin, cmd = m.spin.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ProfileModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Profile"
	}

	header := headerStyle.Width(m.width).Render(" Chirpy | Profile ")

	var lines []string

	if m.user != nil {
		status := "Standard"
		if m.user.IsChirpyRed {
			status = "Chirpy Red"
		}
		info := lipgloss.JoinVertical(
			lipgloss.Left,
			"User ID: "+m.user.ID,
			"Status: "+status,
		)
		box := chirpBoxStyle.Width(m.width - 4).Render(info)
		lines = append(lines, box)
	} else {
		lines = append(lines,
			lipgloss.NewStyle().Foreground(colorMuted).
				Render("No user information available."),
		)
	}

	lines = append(lines, "")
	lines = append(lines, "Press 'u' to update credentials, 'd' to delete account.")

	if m.updating {
		form := lipgloss.JoinVertical(
			lipgloss.Left,
			m.emailInput.View(),
			m.passwordInput.View(),
		)
		if m.errorMsg != "" {
			form = lipgloss.JoinVertical(
				lipgloss.Left,
				form,
				errorStyle.Render("⚠ "+m.errorMsg),
			)
		}
		lines = append(lines, "", form)
	} else if m.confirmDelete {
		confirm := errorStyle.Render("Delete account? This cannot be undone. [y/N]")
		lines = append(lines, "", confirm)
	} else if m.errorMsg != "" {
		lines = append(lines, "", errorStyle.Render("⚠ "+m.errorMsg))
	}

	body := contentStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, lines...),
	)

	// Footer with current time at the bottom-right.
	now := time.Now().Format("2006-01-02 15:04")
	left := "[u] Update credentials  [d] Delete account  [q] Quit"
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

func updateUserCmd(client *api.Chirpy, email, password string) tea.Cmd {
	return func() tea.Msg {
		user, err := client.UpdateUserCredentials(email, password)
		return UserUpdatedMsg{User: user, Err: err}
	}
}

func deleteUserCmd(client *api.Chirpy, userID string) tea.Cmd {
	return func() tea.Msg {
		err := client.DeleteUser(userID)
		return UserDeletedMsg{Err: err}
	}
}

