package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/shubh-man007/Chirpy/tui/internal/api"
	"github.com/shubh-man007/Chirpy/tui/internal/models"
)

type LoginModel struct {
	client *api.Chirpy

	emailInput    textinput.Model
	passwordInput textinput.Model
	focusIndex    int // 0=email, 1=password

	width  int
	height int

	errorMsg string
	loading  bool
	spin     spinner.Model
}

func NewLoginModel(client *api.Chirpy) LoginModel {
	email := textinput.New()
	email.Placeholder = "user@example.com"
	email.Prompt = "Email: "
	email.Focus()

	password := textinput.New()
	password.Placeholder = "password"
	password.Prompt = "Password: "
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'

	s := spinner.New()
	s.Spinner = spinner.Dot

	return LoginModel{
		client:        client,
		emailInput:    email,
		passwordInput: password,
		focusIndex:    0,
		spin:          s,
	}
}

func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginModel) Update(msg tea.Msg) (LoginModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case "tab", "shift+tab":
			m.toggleFocus()
			return m, nil
		case "enter":
			return m.startLogin()
		}

	case LoginFailureMsg:
		m.loading = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
		} else {
			m.errorMsg = "Login failed."
		}
		return m, nil
	}

	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.emailInput, cmd = m.emailInput.Update(msg)
	cmds = append(cmds, cmd)

	m.passwordInput, cmd = m.passwordInput.Update(msg)
	cmds = append(cmds, cmd)

	if m.loading {
		m.spin, cmd = m.spin.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *LoginModel) toggleFocus() {
	if m.focusIndex == 0 {
		m.focusIndex = 1
		m.emailInput.Blur()
		m.passwordInput.Focus()
	} else {
		m.focusIndex = 0
		m.passwordInput.Blur()
		m.emailInput.Focus()
	}
}

func (m LoginModel) startLogin() (LoginModel, tea.Cmd) {
	email := strings.TrimSpace(m.emailInput.Value())
	password := m.passwordInput.Value()

	if email == "" || password == "" {
		m.errorMsg = "Email and password are required."
		return m, nil
	}

	m.loading = true
	m.errorMsg = ""

	return m, tea.Batch(
		m.spin.Tick,
		loginCmd(m.client, email, password),
	)
}

func loginCmd(client *api.Chirpy, email, password string) tea.Cmd {
	return func() tea.Msg {
		res, err := client.Login(email, password)
		if err != nil {
			return LoginFailureMsg{Err: err}
		}

		user := &models.User{
			ID:          res.UserID,
			IsChirpyRed: res.IsChirpyRed,
		}

		return LoginSuccessMsg{User: user}
	}
}

func (m LoginModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Chirpy TUI"
	}

	banner := renderBanner(m.width)

	form := lipgloss.JoinVertical(
		lipgloss.Left,
		m.emailInput.View(),
		"",
		m.passwordInput.View(),
	)

	var parts []string
	parts = append(parts, banner, "")

	if m.loading {
		loadingLine := lipgloss.NewStyle().
			Foreground(colorMuted).
			Render(m.spin.View() + " Logging in...")
		parts = append(parts, loadingLine, "")
	}

	if m.errorMsg != "" {
		parts = append(parts, errorStyle.Render("⚠ "+m.errorMsg), "")
	}

	help := footerStyle.Render("[Tab] Switch field  [Enter] Login  [q] Quit")

	parts = append(parts, form, "", help)

	view := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		contentStyle.Render(view),
	)
}
