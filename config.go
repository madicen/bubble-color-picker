package bubblepicker

import "github.com/charmbracelet/lipgloss"

// Config holds static options for a picker built with New.
type Config struct {
	InitialColor string
	Presets      []string
	FrameStyle   lipgloss.Style
	CustomFrame  bool
	AutoDismiss  bool
}

// Option configures a picker via New(opts...).
type Option func(*Config)

// WithInitialColor sets the starting HSL from a hex string (e.g. "#ff0000").
// Invalid or empty hex falls back to the default red.
func WithInitialColor(hex string) Option {
	return func(c *Config) {
		c.InitialColor = hex
	}
}

// WithPresets adds a row of brand swatches above the hue bar. Each string should be
// a hex color; invalid entries are skipped. Users can focus the preset row (Tab),
// move with ←/→, and press Enter to apply a preset to the picker.
func WithPresets(hexes []string) Option {
	return func(c *Config) {
		if len(hexes) == 0 {
			c.Presets = nil
			return
		}
		c.Presets = append([]string(nil), hexes...)
	}
}

// WithStyle sets the outer frame style (border, padding, margin). The picker still
// draws the inner hue bar and grid; BorderForeground on the frame is merged with
// the current color accent when using the default double border.
func WithStyle(s lipgloss.Style) Option {
	return func(c *Config) {
		c.FrameStyle = s
		c.CustomFrame = true
	}
}

// WithAutoDismiss sets whether the picker marks selections with ColorChangedMsg.Dismiss
// so hosts can remove the picker from the layout immediately after a choice (e.g. modal).
func WithAutoDismiss(v bool) Option {
	return func(c *Config) {
		c.AutoDismiss = v
	}
}
