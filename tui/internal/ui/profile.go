package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"github.com/shubh-man007/Chirpy/tui/internal/api"
	"github.com/shubh-man007/Chirpy/tui/internal/models"
)

const profileChirpsLimit = 20

// ProfileTab indexes for Chirps, Followers, Following
const (
	TabChirps int = iota
	TabFollowers
	TabFollowing
)

// ProfileModel lets the user view their full profile (user details, follow counts, chirps),
// browse followers/following in tabs, update credentials, and delete the account.
type ProfileModel struct {
	client *api.Chirpy
	user   *models.User

	width  int
	height int

	profile  *models.ProfileResponse
	tabIndex int

	followers []models.FollowerRow
	following []models.FollowingRow

	loadingProfile   bool
	loadingFollowers bool
	loadingFollowing bool

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

// InitProfile fetches the full profile. Call when navigating to profile screen.
func (m ProfileModel) InitProfile() tea.Cmd {
	if m.user == nil || m.user.ID == "" {
		return nil
	}
	return tea.Batch(m.spin.Tick, fetchMyProfileCmd(m.client, "", profileChirpsLimit))
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
			return m, nil
		}

		switch msg.String() {
		case "1":
			m.tabIndex = TabChirps
			return m, nil
		case "2":
			m.tabIndex = TabFollowers
			if len(m.followers) == 0 && !m.loadingFollowers {
				m.loadingFollowers = true
				return m, tea.Batch(m.spin.Tick, fetchMyFollowersCmd(m.client))
			}
			return m, nil
		case "3":
			m.tabIndex = TabFollowing
			if len(m.following) == 0 && !m.loadingFollowing {
				m.loadingFollowing = true
				return m, tea.Batch(m.spin.Tick, fetchMyFollowingCmd(m.client))
			}
			return m, nil
		case "u":
			m.updating = true
			m.emailInput.Reset()
			m.passwordInput.Reset()
			m.emailInput.Focus()
			m.passwordInput.Blur()
			m.errorMsg = ""
			return m, nil
		case "d":
			if m.user != nil {
				m.confirmDelete = true
			}
			return m, nil
		case "r":
			// Refresh profile.
			if m.user != nil {
				m.loadingProfile = true
				m.profile = nil
				return m, tea.Batch(m.spin.Tick, fetchMyProfileCmd(m.client, "", profileChirpsLimit))
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
			} else {
				next := (m.tabIndex + 1) % 3
				if msg.String() == "shift+tab" {
					next = (m.tabIndex + 2) % 3
				}
				m.tabIndex = next
				if m.tabIndex == TabFollowers && len(m.followers) == 0 && !m.loadingFollowers {
					m.loadingFollowers = true
					return m, tea.Batch(m.spin.Tick, fetchMyFollowersCmd(m.client))
				}
				if m.tabIndex == TabFollowing && len(m.following) == 0 && !m.loadingFollowing {
					m.loadingFollowing = true
					return m, tea.Batch(m.spin.Tick, fetchMyFollowingCmd(m.client))
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

	case ProfileLoadedMsg:
		m.loadingProfile = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
			return m, nil
		}
		m.profile = msg.Profile
		m.errorMsg = ""
		return m, nil

	case FollowersLoadedMsg:
		m.loadingFollowers = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
			return m, nil
		}
		m.followers = msg.Followers
		return m, nil

	case FollowingLoadedMsg:
		m.loadingFollowing = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
			return m, nil
		}
		m.following = msg.Following
		return m, nil

	case UserUpdatedMsg:
		m.updating = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
			return m, nil
		}
		if m.user != nil && msg.User != nil {
			m.user = msg.User
		}
		// Refresh profile after update.
		if m.user != nil {
			m.loadingProfile = true
			return m, tea.Batch(m.spin.Tick, fetchMyProfileCmd(m.client, "", profileChirpsLimit))
		}
		return m, nil

	case UserDeletedMsg:
		m.deleting = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
			return m, nil
		}
		return m, nil

	case ErrorMsg:
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
	if m.loadingProfile || m.loadingFollowers || m.loadingFollowing || m.updating || m.deleting {
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

	if m.user == nil {
		lines = append(lines,
			lipgloss.NewStyle().Foreground(colorMuted).Render("No user information available."),
		)
	} else {
		if m.loadingProfile && m.profile == nil {
			lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render(m.spin.View()+" Loading profile..."))
		} else if m.profile != nil {
			status := "Standard"
			if m.profile.IsChirpyRed {
				status = "Chirpy Red"
			}
			info := lipgloss.JoinVertical(
				lipgloss.Left,
				"Email: "+m.profile.Email,
				"User ID: "+m.profile.ID,
				"Status: "+status,
				fmt.Sprintf("Followers: %d  Following: %d  Chirps: %d",
					m.profile.FollowersCount, m.profile.FollowingCount, m.profile.ChirpsCount),
			)
			lines = append(lines, chirpBoxStyle.Width(m.width-4).Render(info))

			// Tabs
			tabs := []string{"[1] Chirps", "[2] Followers", "[3] Following"}
			tabStyle := lipgloss.NewStyle().Foreground(colorMuted)
			activeTab := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
			var tabStrs []string
			for i, t := range tabs {
				if i == m.tabIndex {
					tabStrs = append(tabStrs, activeTab.Render(t))
				} else {
					tabStrs = append(tabStrs, tabStyle.Render(t))
				}
			}
			lines = append(lines, "")
			lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, tabStrs...))
			lines = append(lines, "")

			switch m.tabIndex {
			case TabChirps:
				if len(m.profile.Chirps) == 0 {
					lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render("No chirps yet."))
				} else {
					boxWidth := m.width - 4
					if boxWidth < 20 {
						boxWidth = m.width
					}
					for _, c := range m.profile.Chirps {
						body := wordwrap.String(c.Body, boxWidth-4)
						header := fmt.Sprintf("%s", relativeTime(c.CreatedAt))
						content := lipgloss.JoinVertical(lipgloss.Left,
							timestampStyle.Render(header),
							body,
						)
						lines = append(lines, chirpBoxStyle.Width(boxWidth).Render(content))
					}
					if m.profile.NextCursor != nil {
						lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render("(More chirps: scroll with 'r' to refresh and load more)"))
					}
				}
			case TabFollowers:
				if m.loadingFollowers {
					lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render(m.spin.View()+" Loading followers..."))
				} else if len(m.followers) == 0 {
					lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render("No followers."))
				} else {
					for _, f := range m.followers {
						badge := ""
						if f.IsChirpyRed {
							badge = " Chirpy Red"
						}
						lines = append(lines, chirpBoxStyle.Width(m.width-4).Render(
							lipgloss.JoinVertical(lipgloss.Left,
								authorStyle.Render(f.Email),
								"ID: "+f.ID+badge,
							),
						))
					}
				}
			case TabFollowing:
				if m.loadingFollowing {
					lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render(m.spin.View()+" Loading following..."))
				} else if len(m.following) == 0 {
					lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render("Not following anyone."))
				} else {
					for _, f := range m.following {
						badge := ""
						if f.IsChirpyRed {
							badge = " Chirpy Red"
						}
						lines = append(lines, chirpBoxStyle.Width(m.width-4).Render(
							lipgloss.JoinVertical(lipgloss.Left,
								authorStyle.Render(f.Email),
								"ID: "+f.ID+badge,
							),
						))
					}
				}
			}
		} else {
			lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render("Press 'r' to load profile."))
		}
	}

	lines = append(lines, "")
	lines = append(lines, "Press 'u' to update credentials, 'd' to delete account, 'r' to refresh.")

	if m.updating {
		form := lipgloss.JoinVertical(lipgloss.Left,
			m.emailInput.View(),
			m.passwordInput.View(),
		)
		if m.errorMsg != "" {
			form = lipgloss.JoinVertical(lipgloss.Left, form, errorStyle.Render("⚠ "+m.errorMsg))
		}
		lines = append(lines, "", form)
	} else if m.confirmDelete {
		lines = append(lines, "", errorStyle.Render("Delete account? This cannot be undone. [y/N]"))
	} else if m.errorMsg != "" {
		lines = append(lines, "", errorStyle.Render("⚠ "+m.errorMsg))
	}

	body := contentStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))

	now := time.Now().Format("2006-01-02 15:04")
	left := "[1] Chirps [2] Followers [3] Following [u] Update [d] Delete [r] Refresh [q] Quit"
	space := ""
	if m.width > lipgloss.Width(left+now) {
		space = strings.Repeat(" ", m.width-lipgloss.Width(left+now))
	}
	footer := footerStyle.Width(m.width).Render(left + space + now)

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func fetchMyProfileCmd(client *api.Chirpy, cursor string, limit int) tea.Cmd {
	return func() tea.Msg {
		profile, err := client.GetMyProfile(cursor, limit)
		return ProfileLoadedMsg{Profile: profile, Err: err}
	}
}

// fetchMyFollowersCmd uses GET /api/followers (auth) for the profile screen.
func fetchMyFollowersCmd(client *api.Chirpy) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetFollowers()
		if err != nil {
			return FollowersLoadedMsg{Err: err}
		}
		return FollowersLoadedMsg{Followers: resp.Followers}
	}
}

// fetchMyFollowingCmd uses GET /api/following (auth) for the profile screen.
func fetchMyFollowingCmd(client *api.Chirpy) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetFollowing()
		if err != nil {
			return FollowingLoadedMsg{Err: err}
		}
		return FollowingLoadedMsg{Following: resp.Following}
	}
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
