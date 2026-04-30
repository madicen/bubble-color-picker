package bubblepicker

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

// Focus indicates which part of the picker is focused for keyboard input (Tab cycles).
type Focus int

const (
	// FocusPresets is used only when the model was built with WithPresets.
	FocusPresets Focus = iota
	FocusHueBar
	FocusGrid
)

const gridLightTop, gridLightBottom = 98.0, 2.0

// ColorChangedMsg is sent when the user confirms a color (Enter, or mouse release on the grid).
// Handle it at your root or delegated model; Dismiss is true when WithAutoDismiss(true) was used.
type ColorChangedMsg struct {
	Color   string // Hex, e.g. "#rrggbb"
	Dismiss bool   // When true, host should remove the picker from layout (auto-dismiss mode).
}

// ColorCanceledMsg is sent when the user cancels (Esc).
type ColorCanceledMsg struct{}

// ColorChosenMsg is a type alias for ColorChangedMsg for backward compatibility.
type ColorChosenMsg = ColorChangedMsg

const (
	ZoneHueBar   = "picker-hue"
	ZoneGrid     = "picker-grid"
	ZonePresets  = "picker-presets"
)

// Model is the color picker state. It implements tea.Model.
type Model struct {
	HSL   HSL
	Focus Focus

	width  int
	height int

	gridCols int
	gridRows int

	zm *zone.Manager

	titleStyle   lipgloss.Style
	valueStyle   lipgloss.Style
	helpStyle    lipgloss.Style
	outlineStyle lipgloss.Style

	presets     []string
	presetFocus int

	frameStyle  lipgloss.Style
	customFrame bool
	autoDismiss bool
}

// New builds a picker from functional options. With no options, behavior matches the
// legacy default (red, no presets, standard frame, no auto-dismiss).
func New(opts ...Option) Model {
	cfg := Config{}
	for _, o := range opts {
		o(&cfg)
	}
	m := Model{
		HSL:          HSL{H: 0, S: 100, L: 50},
		Focus:        FocusHueBar,
		gridCols:     24,
		gridRows:     12,
		titleStyle:   lipgloss.NewStyle().Bold(true),
		valueStyle:   lipgloss.NewStyle().Padding(0, 1),
		helpStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		outlineStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true),
		frameStyle:   lipgloss.NewStyle(),
	}
	if cfg.InitialColor != "" {
		if hsl, err := HexToHSL(cfg.InitialColor); err == nil {
			m.HSL = hsl.Clamp()
		}
	}
	for _, h := range cfg.Presets {
		if hsl, err := HexToHSL(h); err == nil {
			m.presets = append(m.presets, hsl.Clamp().ToHex())
		}
	}
	if len(m.presets) > 0 {
		m.Focus = FocusPresets
		m.presetFocus = 0
	}
	if cfg.CustomFrame {
		m.frameStyle = cfg.FrameStyle
		m.customFrame = true
	}
	m.autoDismiss = cfg.AutoDismiss
	return m
}

func (m Model) confirmColorCmd() tea.Cmd {
	c := m.Value()
	d := m.autoDismiss
	return func() tea.Msg {
		return ColorChangedMsg{Color: c, Dismiss: d}
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Value returns the current color as hex (e.g. "#rrggbb").
func (m Model) Value() string {
	return m.HSL.Clamp().ToHex()
}

// SetZoneManager sets the zone manager for zone-based mouse interaction.
func (m *Model) SetZoneManager(zm *zone.Manager) {
	m.zm = zm
}

// ViewSize returns display size (width, height in cells), including outer frame.
func (m Model) ViewSize() (width, height int) {
	cols := m.gridCols
	if cols <= 0 {
		cols = 24
	}
	rows := m.gridRows
	if rows <= 0 {
		rows = 12
	}
	extra := 0
	if len(m.presets) > 0 {
		extra = 2 // preset row + spacing in inner stack
	}
	innerW, innerH := cols+2, 10+rows+extra
	return innerW + 2 + 2, innerH + 2
}

func (m Model) cycleFocus(next bool) Focus {
	hasP := len(m.presets) > 0
	if !hasP {
		if m.Focus == FocusHueBar {
			return FocusGrid
		}
		return FocusHueBar
	}
	order := []Focus{FocusPresets, FocusHueBar, FocusGrid}
	idx := 0
	for i, f := range order {
		if f == m.Focus {
			idx = i
			break
		}
	}
	if next {
		idx = (idx + 1) % 3
	} else {
		idx = (idx + 2) % 3
	}
	return order[idx]
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		const maxGridCols = 24
		const maxGridRows = 10
		contentW := max(msg.Width-2, 16)
		m.width = contentW
		m.gridCols = contentW
		if m.gridCols > maxGridCols {
			m.gridCols = maxGridCols
		}
		availH := msg.Height - 8
		if len(m.presets) > 0 {
			availH -= 2
		}
		if availH > maxGridRows {
			m.gridRows = maxGridRows
		} else if availH > 4 {
			m.gridRows = availH
		} else {
			m.gridRows = 4
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if len(m.presets) > 0 && m.Focus == FocusPresets {
				if m.presetFocus >= 0 && m.presetFocus < len(m.presets) {
					if hsl, err := HexToHSL(m.presets[m.presetFocus]); err == nil {
						m.HSL = hsl.Clamp()
					}
				}
				return m, m.confirmColorCmd()
			}
			return m, m.confirmColorCmd()
		case "esc":
			return m, func() tea.Msg { return ColorCanceledMsg{} }
		case "tab":
			m.Focus = m.cycleFocus(true)
			return m, nil
		case "shift+tab":
			m.Focus = m.cycleFocus(false)
			return m, nil
		case "left", "h":
			if len(m.presets) > 0 && m.Focus == FocusPresets {
				if m.presetFocus > 0 {
					m.presetFocus--
				}
				return m, nil
			}
			if m.Focus == FocusHueBar {
				m.HSL.H = math.Mod(m.HSL.H-8+360, 360)
			} else if m.Focus == FocusGrid {
				m.HSL.S = math.Max(0, m.HSL.S-4)
			}
			m.HSL = m.HSL.Clamp()
			return m, nil
		case "right", "l":
			if len(m.presets) > 0 && m.Focus == FocusPresets {
				if m.presetFocus < len(m.presets)-1 {
					m.presetFocus++
				}
				return m, nil
			}
			if m.Focus == FocusHueBar {
				m.HSL.H = math.Mod(m.HSL.H+8, 360)
			} else if m.Focus == FocusGrid {
				m.HSL.S = math.Min(100, m.HSL.S+4)
			}
			m.HSL = m.HSL.Clamp()
			return m, nil
		case "up", "k":
			if m.Focus == FocusGrid {
				m.HSL.L = math.Min(100, m.HSL.L+4)
			}
			m.HSL = m.HSL.Clamp()
			return m, nil
		case "down", "j":
			if m.Focus == FocusGrid {
				m.HSL.L = math.Max(0, m.HSL.L-4)
			}
			m.HSL = m.HSL.Clamp()
			return m, nil
		}
		return m, nil

	case tea.MouseMsg:
		action := msg.Action
		if m.zm == nil {
			return m, nil
		}
		isClick := action == tea.MouseActionPress || action == tea.MouseActionRelease
		isMotion := action == tea.MouseActionMotion
		if !isClick && !isMotion {
			return m, nil
		}
		if len(m.presets) > 0 {
			if z := m.zm.Get(ZonePresets); z != nil && z.InBounds(msg) {
				relX, _ := z.Pos(msg)
				cell := 3 // 2 cells + 1 gap per preset
				idx := (relX - 1) / cell
				if idx < 0 {
					idx = 0
				}
				if idx >= len(m.presets) {
					idx = len(m.presets) - 1
				}
				m.presetFocus = idx
				if hsl, err := HexToHSL(m.presets[idx]); err == nil {
					m.HSL = hsl.Clamp()
				}
				if isClick && action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
					return m, m.confirmColorCmd()
				}
				return m, nil
			}
		}
		if z := m.zm.Get(ZoneHueBar); z != nil && z.InBounds(msg) {
			relX, _ := z.Pos(msg)
			w := m.gridCols
			if w <= 0 {
				w = 24
			}
			contentCol := relX - 1
			if contentCol >= 0 && contentCol < w {
				m.HSL.H = math.Mod((float64(contentCol)+0.5)/float64(w)*360, 360)
				if m.HSL.H < 0 {
					m.HSL.H += 360
				}
				m.HSL = m.HSL.Clamp()
			}
			return m, nil
		}
		if z := m.zm.Get(ZoneGrid); z != nil && z.InBounds(msg) {
			relX, relY := z.Pos(msg)
			w := m.gridCols
			if w <= 0 {
				w = 24
			}
			rows := m.gridRows
			if rows <= 0 {
				rows = 12
			}
			contentCol := relX - 1
			contentRow := relY - 1
			if contentCol >= 0 && contentCol < w && contentRow >= 0 && contentRow < rows {
				sx := (float64(contentCol) + 0.5) / float64(w) * 100
				m.HSL.S = math.Max(0, math.Min(100, sx))
				sy := (float64(contentRow) + 0.5) / float64(rows)
				m.HSL.L = gridLightTop - sy*(gridLightTop-gridLightBottom)
				m.HSL = m.HSL.Clamp()
			}
			if isClick && action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
				return m, m.confirmColorCmd()
			}
			return m, nil
		}
		return m, nil
	}

	return m, nil
}

// View renders the picker.
func (m Model) View() string {
	cols := m.gridCols
	if cols <= 0 {
		cols = 24
	}
	rows := m.gridRows
	if rows <= 0 {
		rows = 12
	}
	focusBorderFocused := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).BorderRight(true).BorderTop(true).BorderBottom(true).
		BorderForeground(lipgloss.Color("250")).
		PaddingTop(0).PaddingBottom(0).PaddingLeft(0).PaddingRight(0).
		MarginTop(0).MarginBottom(0)
	focusBorderUnfocused := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).BorderRight(true).BorderTop(true).BorderBottom(true).
		BorderForeground(lipgloss.Color("241")).
		PaddingTop(0).PaddingBottom(0).PaddingLeft(0).PaddingRight(0).
		MarginTop(0).MarginBottom(0)

	hueBar := ""
	for i := 0; i < cols; i++ {
		h := math.Mod(float64(i)/float64(cols)*360, 360)
		r, g, b := HSLToRGB(h, 100, 50)
		hexStr := rgbToHexString(r, g, b)
		style := lipgloss.NewStyle().Background(lipgloss.Color(hexStr))
		hueBar += style.Render(" ")
	}
	hueBar = lipgloss.NewStyle().Width(cols).Render(hueBar)

	hueIdx := int(m.HSL.H / 360 * float64(cols))
	if hueIdx >= cols {
		hueIdx = cols - 1
	}
	hueMarkerLine := ""
	for i := 0; i < cols; i++ {
		if i == hueIdx {
			hueMarkerLine += m.outlineStyle.Render("▲")
		} else {
			hueMarkerLine += " "
		}
	}
	hueMarkerLine = lipgloss.NewStyle().Width(cols).Render(hueMarkerLine)

	sCol := int(m.HSL.S / 100 * float64(cols))
	if sCol >= cols {
		sCol = cols - 1
	}
	sRow := int((gridLightTop - m.HSL.L) / (gridLightTop - gridLightBottom) * float64(rows))
	if sRow >= rows {
		sRow = rows - 1
	}
	if sRow < 0 {
		sRow = 0
	}
	grid := ""
	for row := 0; row < rows; row++ {
		ly := gridLightTop - float64(row)/float64(rows)*(gridLightTop-gridLightBottom)
		var line strings.Builder
		for col := 0; col < cols; col++ {
			sx := float64(col) / float64(cols) * 100
			r, g, b := HSLToRGB(m.HSL.H, sx, ly)
			hexStr := rgbToHexString(r, g, b)
			style := lipgloss.NewStyle().Background(lipgloss.Color(hexStr)).Width(1)
			isSelected := col == sCol && row == sRow
			if isSelected {
				line.WriteString(m.outlineStyle.Background(lipgloss.Color(hexStr)).Width(1).Render("●"))
			} else {
				line.WriteString(style.Render(" "))
			}
		}
		grid += line.String() + "\n"
	}

	trunc := lipgloss.NewStyle().MaxWidth(cols)
	toCols := func(content string) string {
		content = trunc.Render(content)
		w := min(lipgloss.Width(content), cols)
		return content + strings.Repeat(" ", cols-w)
	}
	wrap := func(s string) string { return " " + toCols(s) + " " }
	title := wrap(m.titleStyle.Render("Pick a color"))
	value := wrap(m.valueStyle.Render(m.Value()))
	help1 := wrap(m.helpStyle.Render("↵ pick  ⎋ close"))
	help2 := wrap(m.helpStyle.Render("⇥ switch  ←↑↓→ move"))

	hueBorder := focusBorderUnfocused
	if m.Focus == FocusHueBar {
		hueBorder = focusBorderFocused
	}
	hueBlock := hueBorder.Width(cols).Render(hueBar + "\n" + hueMarkerLine)

	gridTrimmed := strings.TrimSuffix(grid, "\n")
	gridBorder := focusBorderUnfocused
	if m.Focus == FocusGrid {
		gridBorder = focusBorderFocused
	}
	gridBlock := gridBorder.Width(cols).Render(gridTrimmed)

	var presetBlock string
	if len(m.presets) > 0 {
		var cells []string
		for i, hex := range m.presets {
			st := lipgloss.NewStyle().Background(lipgloss.Color(hex)).Width(2).Render("  ")
			if m.Focus == FocusPresets && i == m.presetFocus {
				st = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("15")).Render(
					lipgloss.NewStyle().Background(lipgloss.Color(hex)).Width(2).Render("  "))
			}
			cells = append(cells, st)
		}
		row := strings.Join(cells, " ")
		pb := lipgloss.NewStyle().Width(cols).Render(row)
		presetBorder := focusBorderUnfocused
		if m.Focus == FocusPresets {
			presetBorder = focusBorderFocused
		}
		presetBlock = presetBorder.Width(cols).Render(pb)
	}

	if m.zm != nil {
		hueBlock = m.zm.Mark(ZoneHueBar, hueBlock)
		gridBlock = m.zm.Mark(ZoneGrid, gridBlock)
		if presetBlock != "" {
			presetBlock = m.zm.Mark(ZonePresets, presetBlock)
		}
	}

	var inner string
	if presetBlock != "" {
		inner = lipgloss.JoinVertical(lipgloss.Left,
			title, presetBlock, hueBlock, gridBlock, value, help1, help2,
		)
	} else {
		inner = lipgloss.JoinVertical(lipgloss.Left,
			title, hueBlock, gridBlock, value, help1, help2,
		)
	}

	if m.customFrame {
		return m.frameStyle.Render(inner)
	}
	frame := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(m.Value())).
		Padding(0, 1)
	return frame.Render(inner)
}

func rgbToHexString(r, g, b float64) string {
	rr := byte(math.Round(math.Max(0, math.Min(1, r)) * 255))
	gg := byte(math.Round(math.Max(0, math.Min(1, g)) * 255))
	bb := byte(math.Round(math.Max(0, math.Min(1, b)) * 255))
	return fmt.Sprintf("#%02x%02x%02x", rr, gg, bb)
}
