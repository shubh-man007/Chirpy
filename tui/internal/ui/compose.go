package ui

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/shubh-man007/Chirpy/tui/internal/api"
)

type ComposeModel struct {
	client *api.Chirpy

	ta textarea.Model

	width  int
	height int

	posting  bool
	errorMsg string
}

const maxChirpChars = 140

func NewComposeModel(client *api.Chirpy) ComposeModel {
	ta := textarea.New()
	ta.Placeholder = "What's on your mind?"
	ta.Focus()
	ta.ShowLineNumbers = false

	return ComposeModel{
		client: client,
		ta:     ta,
	}
}

func (m ComposeModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m ComposeModel) Update(msg tea.Msg) (ComposeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalculateSize()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			if m.posting {
				return m, nil
			}
			text := strings.TrimSpace(m.ta.Value())
			count := utf8.RuneCountInString(text)
			if text == "" {
				m.errorMsg = "Chirp cannot be empty."
				return m, nil
			}
			if count > maxChirpChars {
				m.errorMsg = "Chirp cannot exceed 140 characters."
				return m, nil
			}
			m.posting = true
			m.errorMsg = ""
			return m, postChirpCmd(m.client, text)
		}

	case ChirpPostedMsg:
		m.posting = false
		m.ta.SetValue("")
		return m, nil

	case ErrorMsg:
		m.posting = false
		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	return m, cmd
}

func (m *ComposeModel) recalculateSize() {
	if m.width == 0 || m.height == 0 {
		return
	}

	header := headerStyle.Width(m.width).Render(" Chirpy | Compose ")
	footer := footerStyle.Width(m.width).Render("[Ctrl+S] Post  [Esc] Back  [q] Quit")

	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 2
	if contentHeight < 3 {
		contentHeight = 3
	}

	m.ta.SetWidth(m.width - 4)
	m.ta.SetHeight(contentHeight)
}

func (m ComposeModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Compose"
	}

	header := headerStyle.Width(m.width).Render(" Chirpy | Compose ")

	text := m.ta.Value()
	count := utf8.RuneCountInString(text)

	counterStyle := lipgloss.NewStyle().Foreground(colorMuted)
	if count > maxChirpChars {
		counterStyle = counterStyle.Foreground(colorError)
	}
	counter := counterStyle.Render(
		strings.TrimSpace(
			strings.Join([]string{
				"",
				strings.TrimSpace(
					lipgloss.NewStyle().Render(
						"Characters: " +
							lipgloss.NewStyle().Bold(true).Render(
								strconv.Itoa(count),
							) +
							" / " + strconv.Itoa(maxChirpChars),
					),
				),
			}, " "),
		),
	)

	var bodyLines []string
	bodyLines = append(bodyLines, m.ta.View(), "")
	if m.errorMsg != "" {
		bodyLines = append(bodyLines, errorStyle.Render("⚠ "+m.errorMsg))
	}
	bodyLines = append(bodyLines, counter)

	body := contentStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, bodyLines...),
	)

	footer := footerStyle.Width(m.width).Render("[Ctrl+S] Post  [Esc] Back  [q] Quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
		footer,
	)
}

func postChirpCmd(client *api.Chirpy, body string) tea.Cmd {
	return func() tea.Msg {
		chirp, err := client.CreateChirp(body)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return ChirpPostedMsg{Chirp: chirp}
	}
}
