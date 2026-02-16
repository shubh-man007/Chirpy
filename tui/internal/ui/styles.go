package ui

import "github.com/charmbracelet/lipgloss"

const (
	// Base colors.
	colorBackground = lipgloss.Color("#1e1e2e")
	colorForeground = lipgloss.Color("#cdd6f4")

	// Accent colors.
	colorPrimary   = lipgloss.Color("#89b4fa")
	colorSuccess   = lipgloss.Color("#a6e3a1")
	colorWarning   = lipgloss.Color("#f9e2af")
	colorError     = lipgloss.Color("#f38ba8")
	colorMuted     = lipgloss.Color("#6c7086")
	colorChirpyRed = lipgloss.Color("#f38ba8")
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(colorForeground).
			Background(colorPrimary).
			Bold(true).
			Padding(0, 1)

	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true).
			Padding(0, 1)

	chirpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(1).
			MarginBottom(1)

	selectedChirpStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(1).
				MarginBottom(1)

	authorStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	timestampStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	chirpyRedBadge = lipgloss.NewStyle().
			Foreground(colorChirpyRed).
			Bold(true).
			SetString("🔥 Chirpy Red")
)

const bannerArt = `
 ██████╗██╗  ██╗██╗██████╗ ██████╗ ██╗   ██╗
██╔════╝██║  ██║██║██╔══██╗██╔══██╗╚██╗ ██╔╝
██║     ███████║██║██████╔╝██████╔╝ ╚████╔╝ 
██║     ██╔══██║██║██╔══██╗██╔═══╝   ╚██╔╝  
╚██████╗██║  ██║██║██║  ██║██║        ██║   
 ╚═════╝╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚═╝        ╚═╝   
`

func renderBanner(width int) string {
	style := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Align(lipgloss.Center)

	if width > 0 {
		style = style.Width(width)
	}

	return style.Render(bannerArt)
}
