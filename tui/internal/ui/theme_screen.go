package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	uiTheme "github.com/shubh-man007/Chirpy/tui/internal/theme"
)

// ThemeModel provides a simple theme selection screen, closely inspired
// by the Kindria TUI themes view.
type ThemeModel struct {
	themes      []uiTheme.Palette
	current     uiTheme.Palette
	cursor      int
	width       int
	height      int
}

func NewThemeModel() ThemeModel {
	themes := uiTheme.All()
	current := uiTheme.Default()
	if saved, err := uiTheme.LoadSelected(); err == nil {
		current = saved
	}

	cursor := 0
	for i, t := range themes {
		if strings.EqualFold(t.Name, current.Name) {
			cursor = i
			break
		}
	}

	return ThemeModel{
		themes:  themes,
		current: current,
		cursor:  cursor,
	}
}

func (m ThemeModel) Init() tea.Cmd {
	return nil
}

func (m ThemeModel) Update(msg tea.Msg) (ThemeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.themes)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.themes) == 0 {
				return m, nil
			}
			m.current = m.themes[m.cursor]
			ApplyThemePalette(m.current)
			_ = uiTheme.SaveSelected(m.current.Name)
			return m, nil
		}
	}

	return m, nil
}

func (m ThemeModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Themes"
	}

	header := headerStyle.Width(m.width).Render(" Chirpy | Themes ")

	var rows []string
	rows = append(rows,
		"Select a color theme:",
		"",
	)

	for i, t := range m.themes {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		marker := " "
		if strings.EqualFold(t.Name, m.current.Name) {
			marker = "*"
		}
		lineStyle := lipgloss.NewStyle().Foreground(colorForeground)
		if i == m.cursor {
			lineStyle = lineStyle.Foreground(colorPrimary)
		}
		rows = append(rows, lineStyle.Render(prefix+"["+marker+"] "+t.Name))
	}

	rows = append(rows, "")
	rows = append(rows, "Use ↑/↓ (j/k) to move, Enter/Space to apply.")

	body := contentStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)

	// Footer with current time at bottom-right.
	now := time.Now().Format("2006-01-02 15:04")
	left := "[↑/k ↓/j] Move  [Enter/Space] Apply  [q] Quit"
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

