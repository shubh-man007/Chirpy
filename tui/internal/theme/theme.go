package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Palette struct {
	Name           string
	Normal         string
	SubtleLight    string
	SubtleDark     string
	BorderLight    string
	BorderDark     string
	HighlightLight string
	HighlightDark  string
}

var palettes = []Palette{
	{
		Name:           "Catppuccin Mocha",
		Normal:         "#CDD6F4",
		SubtleLight:    "#E6E9EF",
		SubtleDark:     "#585B70",
		BorderLight:    "#89B4FA",
		BorderDark:     "#89B4FA",
		HighlightLight: "#F5C2E7",
		HighlightDark:  "#F5C2E7",
	},
	{
		Name:           "Nord",
		Normal:         "#ECEFF4",
		SubtleLight:    "#D8DEE9",
		SubtleDark:     "#4C566A",
		BorderLight:    "#81A1C1",
		BorderDark:     "#81A1C1",
		HighlightLight: "#88C0D0",
		HighlightDark:  "#88C0D0",
	},
	{
		Name:           "Dracula",
		Normal:         "#F8F8F2",
		SubtleLight:    "#E6E6E6",
		SubtleDark:     "#6272A4",
		BorderLight:    "#BD93F9",
		BorderDark:     "#BD93F9",
		HighlightLight: "#FF79C6",
		HighlightDark:  "#FF79C6",
	},
}

func All() []Palette {
	cp := make([]Palette, len(palettes))
	copy(cp, palettes)
	return cp
}

func Default() Palette {
	return palettes[0]
}

func ByName(name string) (Palette, bool) {
	for _, p := range palettes {
		if strings.EqualFold(p.Name, name) {
			return p, true
		}
	}
	return Palette{}, false
}

type savedTheme struct {
	Name string `json:"name"`
}

func SaveSelected(name string) error {
	if _, ok := ByName(name); !ok {
		return fmt.Errorf("unknown theme: %s", name)
	}
	path, err := themeConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(savedTheme{Name: name})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadSelected() (Palette, error) {
	path, err := themeConfigPath()
	if err != nil {
		return Palette{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Palette{}, err
	}
	var s savedTheme
	if err := json.Unmarshal(data, &s); err != nil {
		return Palette{}, err
	}
	p, ok := ByName(s.Name)
	if !ok {
		return Palette{}, fmt.Errorf("theme not found: %s", s.Name)
	}
	return p, nil
}

func themeConfigPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "chirpy", "theme.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "chirpy", "theme.json"), nil
}
