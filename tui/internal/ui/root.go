package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shubh-man007/Chirpy/tui/internal/api"
	"github.com/shubh-man007/Chirpy/tui/internal/models"
)

// ScreenType identifies the current high-level screen.
type ScreenType int

const (
	ScreenLogin ScreenType = iota
	ScreenFeed
	ScreenProfile
	ScreenFollowing
	ScreenFollowers
	ScreenSearch
	ScreenCompose
	ScreenUserProfile
)

// RootModel is the top-level Bubble Tea model that routes
// messages and composes views from child screens.
type RootModel struct {
	client      *api.Chirpy
	currentUser *models.User

	width  int
	height int

	currentScreen ScreenType
	screenStack   []ScreenType

	showHelp    bool
	globalError string

	// Child screen models.
	loginModel   LoginModel
	feedModel    FeedModel
	composeModel ComposeModel
}

// NewRootModel constructs the root model with the given API client.
func NewRootModel(client *api.Chirpy) RootModel {
	m := RootModel{
		client:        client,
		currentScreen: ScreenLogin,
	}

	m.loginModel = NewLoginModel(client)
	m.feedModel = NewFeedModel(client)
	m.composeModel = NewComposeModel(client)

	return m
}

// Init implements tea.Model.
func (m RootModel) Init() tea.Cmd {
	return m.loginModel.Init()
}

// Update implements tea.Model.
func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Forward size to children.
		m.loginModel, _ = m.loginModel.Update(msg)
		m.feedModel, _ = m.feedModel.Update(msg)
		m.composeModel, _ = m.composeModel.Update(msg)
		return m, nil

	case tea.KeyMsg:
		// Global navigation (except on login screen).
		if m.currentScreen != ScreenLogin {
			switch msg.String() {
			case "1":
				m.currentScreen = ScreenFeed
				return m, nil
			case "c":
				m.currentScreen = ScreenCompose
				return m, nil
			case "esc":
				// Simple back: return to feed from any non-login screen.
				m.currentScreen = ScreenFeed
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			// Allow quitting from anywhere.
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		}

	case LoginSuccessMsg:
		// Store authenticated user and navigate to feed.
		m.currentUser = msg.User
		m.currentScreen = ScreenFeed
		return m, m.feedModel.InitFeed()

	case ChirpPostedMsg:
		// After posting, jump back to feed and refresh it.
		m.currentScreen = ScreenFeed
		return m, m.feedModel.InitFeed()
	}

	// Route to active screen.
	return m.updateCurrentScreen(msg)
}

func (m RootModel) updateCurrentScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.currentScreen {
	case ScreenLogin:
		var cmd tea.Cmd
		m.loginModel, cmd = m.loginModel.Update(msg)
		return m, cmd
	case ScreenFeed:
		var cmd tea.Cmd
		m.feedModel, cmd = m.feedModel.Update(msg)
		return m, cmd
	case ScreenCompose:
		var cmd tea.Cmd
		m.composeModel, cmd = m.composeModel.Update(msg)
		return m, cmd
	default:
		// Until other screens are implemented, fall back to login.
		var cmd tea.Cmd
		m.loginModel, cmd = m.loginModel.Update(msg)
		return m, cmd
	}
}

// View implements tea.Model.
func (m RootModel) View() string {
	if m.width == 0 || m.height == 0 {
		// We haven't received a WindowSizeMsg yet; render a minimal view.
		return "Loading Chirpy TUI..."
	}

	var body string

	switch m.currentScreen {
	case ScreenLogin:
		body = m.loginModel.View()
	case ScreenFeed:
		body = m.feedModel.View()
	case ScreenCompose:
		body = m.composeModel.View()
	default:
		// For now, show a placeholder until other screens are implemented.
		body = "Chirpy TUI is under construction."
	}

	if m.showHelp {
		help := m.renderHelp()
		return lipgloss.JoinVertical(lipgloss.Left, body, help)
	}

	return body
}

func (m RootModel) renderHelp() string {
	lines := []string{
		"[q] Quit  [?] Toggle help",
	}
	text := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return footerStyle.Width(m.width).Render(text)
}

// NewProgram constructs a Bubble Tea program for the given client.
func NewProgram(client *api.Chirpy) *tea.Program {
	return tea.NewProgram(
		NewRootModel(client),
		tea.WithAltScreen(), // Use the full terminal screen.
	)
}

