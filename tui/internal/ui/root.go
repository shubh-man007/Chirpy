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
	ScreenTheme
)

type focusArea int

const (
	focusSidebar focusArea = iota
	focusContent
)

// RootModel is the top-level Bubble Tea model that routes
// messages and composes views from child screens.
type RootModel struct {
	client      *api.Chirpy
	currentUser *models.User

	width  int
	height int

	sideBarWidth  int
	menuOptions   []string
	sideBarCursor int
	activeArea    focusArea

	currentScreen ScreenType
	screenStack   []ScreenType

	showHelp    bool
	globalError string

	// Child screen models.
	loginModel   LoginModel
	feedModel    FeedModel
	composeModel ComposeModel
	profileModel ProfileModel
	browseModel  BrowseModel
	themeModel   ThemeModel
}

// NewRootModel constructs the root model with the given API client.
func NewRootModel(client *api.Chirpy) RootModel {
	m := RootModel{
		client:        client,
		currentScreen: ScreenLogin,
		menuOptions:   []string{"Feed", "Profile", "Browse", "Themes"},
		activeArea:    focusContent,
	}

	m.loginModel = NewLoginModel(client)
	m.feedModel = NewFeedModel(client)
	m.composeModel = NewComposeModel(client)
	m.profileModel = NewProfileModel(client)
	m.browseModel = NewBrowseModel(client)
	m.themeModel = NewThemeModel()

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
		m.sideBarWidth = m.width / 8
		if m.sideBarWidth < 18 {
			m.sideBarWidth = 18
		}

		// Forward adjusted size to children. Login gets full width,
		// content screens get width reduced by sidebar.
		loginMsg := msg
		contentWidth := m.width - m.sideBarWidth - 4
		if contentWidth < 20 {
			contentWidth = m.width
		}
		contentMsg := tea.WindowSizeMsg{Width: contentWidth, Height: msg.Height}

		m.loginModel, _ = m.loginModel.Update(loginMsg)
		m.feedModel, _ = m.feedModel.Update(contentMsg)
		m.composeModel, _ = m.composeModel.Update(contentMsg)
		m.profileModel, _ = m.profileModel.Update(contentMsg)
		m.browseModel, _ = m.browseModel.Update(contentMsg)
		m.themeModel, _ = m.themeModel.Update(contentMsg)
		return m, nil

	case tea.KeyMsg:
		// Global navigation (except on login screen).
		if m.currentScreen != ScreenLogin {
			switch msg.String() {
			case "1":
				m.currentScreen = ScreenFeed
				return m, nil
			case "2":
				m.currentScreen = ScreenProfile
				return m, nil
			case "3", "b":
				m.currentScreen = ScreenSearch
				return m, nil
			case "4", "t":
				m.currentScreen = ScreenTheme
				return m, nil
			case "c":
				m.currentScreen = ScreenCompose
				return m, nil
			case "left", "ctrl+h", "esc":
				// Move focus to sidebar.
				m.activeArea = focusSidebar
				return m, nil
			case "right", "ctrl+l":
				// Move focus back to content.
				m.activeArea = focusContent
				return m, nil
			}

			// Sidebar navigation when focused.
			if m.activeArea == focusSidebar {
				switch msg.String() {
				case "up", "k":
					if m.sideBarCursor > 0 {
						m.sideBarCursor--
					}
					return m, nil
				case "down", "j":
					if m.sideBarCursor < len(m.menuOptions)-1 {
						m.sideBarCursor++
					}
					return m, nil
				case "enter":
					switch m.menuOptions[m.sideBarCursor] {
					case "Feed":
						m.currentScreen = ScreenFeed
					case "Profile":
						m.currentScreen = ScreenProfile
					case "Browse":
						m.currentScreen = ScreenSearch
					case "Themes":
						m.currentScreen = ScreenTheme
					}
					m.activeArea = focusContent
					return m, nil
				}
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
		m.profileModel.SetUser(msg.User)
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
	case ScreenProfile:
		var cmd tea.Cmd
		m.profileModel, cmd = m.profileModel.Update(msg)
		return m, cmd
	case ScreenSearch:
		var cmd tea.Cmd
		m.browseModel, cmd = m.browseModel.Update(msg)
		return m, cmd
	case ScreenTheme:
		var cmd tea.Cmd
		m.themeModel, cmd = m.themeModel.Update(msg)
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
	case ScreenProfile:
		body = m.profileModel.View()
	case ScreenSearch:
		body = m.browseModel.View()
	case ScreenTheme:
		body = m.themeModel.View()
	default:
		// For now, show a placeholder until other screens are implemented.
		body = "Chirpy TUI is under construction."
	}

	// Wrap non-login screens with a sidebar layout.
	if m.currentScreen != ScreenLogin {
		sidebar := m.renderSidebar()
		body = lipgloss.JoinHorizontal(lipgloss.Left, sidebar, body)
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

func (m RootModel) renderSidebar() string {
	if m.sideBarWidth <= 0 {
		return ""
	}

	itemWidth := m.sideBarWidth - 2
	if itemWidth < 0 {
		itemWidth = 0
	}

	inactive := lipgloss.NewStyle().
		Foreground(colorMuted).
		Width(itemWidth).
		PaddingLeft(1).
		MarginBottom(1)
	active := inactive.Foreground(colorPrimary)

	var rendered []string
	for i, label := range m.menuOptions {
		prefix := "  "
		if i == m.sideBarCursor {
			prefix = "> "
		}
		text := prefix + label
		if i == m.sideBarCursor {
			rendered = append(rendered, active.Render(text))
		} else {
			rendered = append(rendered, inactive.Render(text))
		}
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(colorMuted).
		Width(m.sideBarWidth).
		Height(m.height - 2)

	if m.activeArea == focusSidebar {
		style = style.BorderForeground(colorPrimary)
	}

	return style.Render(lipgloss.JoinVertical(lipgloss.Left, rendered...))
}

// NewProgram constructs a Bubble Tea program for the given client.
func NewProgram(client *api.Chirpy) *tea.Program {
	return tea.NewProgram(
		NewRootModel(client),
		tea.WithAltScreen(), // Use the full terminal screen.
	)
}

