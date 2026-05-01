// Package bubblepicker provides SwatchPicker: a color square that opens the full
// modal picker on click, with overlay positioning and mouse offset handled internally.

package bubblepicker

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	overlay "github.com/madicen/bubble-overlay"
)

// SwatchPicker is a color swatch (square + optional label + hex) that opens the
// modal picker when clicked. Embed it in your model, set bounds, forward messages,
// and use ViewWithOverlay(mainView, width, height) for the view. On ColorChosenMsg,
// update the swatch with SetColor(msg.Color) (or store the color in your app).
// SetZoneManager so the picker uses zone-based mouse interaction (host must Scan the view).
type SwatchPicker struct {
	color   string
	label   string
	row     int
	col     int
	w       int
	h       int
	picker  Model
	open    bool
	focused bool // When true, arrow is highlighted (for keyboard-only clients)

	// ignoreNextRelease: when true, the next left-button MouseActionRelease is ignored.
	// Set when we open on press so the same click's release does not confirm the color.
	ignoreNextRelease bool

	// Optional zone manager: when set, picker uses zones and receives raw screen mouse events.
	zoneManager *zone.Manager

	// Set by ViewWithOverlay for in-modal bounds check when forwarding mouse
	lastOverlayLeft   int
	lastOverlayTop    int
	lastModalW        int
	lastOverlayHeight int
	lastViewWidth     int
	lastViewHeight    int

	// Extra Options passed to New when opening the modal (e.g. WithPresets, WithAutoDismiss).
	pickerOpts []Option
}

// Picker symbol shown to the right of the color square (indicates "click to open picker").
const swatchPickerSymbol = "▼"

// NewSwatchPicker returns a swatch that shows color and opens the modal picker on click.
// initialColor is hex (e.g. "#7E00AF"); label is optional (not shown in the minimal UI).
func NewSwatchPicker(initialColor, label string) *SwatchPicker {
	if initialColor == "" {
		initialColor = "#7E00AF"
	}
	return &SwatchPicker{
		color: initialColor,
		label: label,
	}
}

// SetPickerOptions configures extra bubblepicker.New options used whenever the modal opens.
// WithInitialColor is always applied from the swatch’s current color first; these append after it.
// Typical use: WithPresets, WithAutoDismiss.
func (s *SwatchPicker) SetPickerOptions(opts ...Option) {
	s.pickerOpts = append([]Option(nil), opts...)
}

func (s *SwatchPicker) newPickerModel() Model {
	opts := make([]Option, 0, 1+len(s.pickerOpts))
	opts = append(opts, WithInitialColor(s.color))
	opts = append(opts, s.pickerOpts...)
	p := New(opts...)
	if s.zoneManager != nil {
		p.SetZoneManager(s.zoneManager)
	}
	return p
}

// SwatchView returns the swatch as a single line: one cell of color plus the picker symbol (▼).
// No border. When focused (e.g. for keyboard nav), the arrow is highlighted.
func (s *SwatchPicker) SwatchView() string {
	colorBlock := lipgloss.NewStyle().Background(lipgloss.Color(s.color)).Render(" ")
	symbol := swatchPickerSymbol
	if s.focused {
		symbol = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true).Render(symbol)
	}
	return colorBlock + symbol
}

// SetFocused sets whether the swatch is focused (e.g. for keyboard-only navigation).
// When true, the picker arrow (▼) is rendered in a brighter color.
func (s *SwatchPicker) SetFocused(f bool) {
	s.focused = f
}

// Focused returns whether the swatch is currently focused.
func (s *SwatchPicker) Focused() bool {
	return s.focused
}

// Open returns whether the swatch's color picker modal is currently open.
// Use this so your app can route input only to the open swatch when a modal is active.
func (s *SwatchPicker) Open() bool {
	return s.open
}

// SetZoneManager sets the zone manager for the picker. When set, the picker uses
// zone-based mouse interaction (click hue bar / grid / Accept) and receives raw
// screen mouse events; the host must run zone.Scan() on the view that contains the overlay.
func (s *SwatchPicker) SetZoneManager(zm *zone.Manager) {
	s.zoneManager = zm
}

// Size returns the display size (width, height in cells) of the swatch: 2 wide (color + symbol), 1 high.
// Use this when building your layout and when calling SetBounds.
func (s *SwatchPicker) Size() (width, height int) {
	return 2, 1
}

// SetBounds sets where the swatch is drawn (0-based row, col) and its size (w, h).
// If w or h is 0, Size() is used for that dimension. Call this before ViewWithOverlay
// so the modal is centered on the swatch and clicks are detected correctly.
func (s *SwatchPicker) SetBounds(row, col, w, h int) {
	s.row = row
	s.col = col
	if w <= 0 || h <= 0 {
		pw, ph := s.Size()
		if w <= 0 {
			s.w = pw
		} else {
			s.w = w
		}
		if h <= 0 {
			s.h = ph
		} else {
			s.h = h
		}
	} else {
		s.w = w
		s.h = h
	}
}

// SetColor sets the current color (hex). Call this when you receive ColorChosenMsg.
func (s *SwatchPicker) SetColor(c string) {
	s.color = c
}

// Color returns the current color (hex).
func (s *SwatchPicker) Color() string {
	return s.color
}

// ViewWithOverlay returns the view to display. If the picker is open, it overlays
// the modal on mainView. It stores overlay position and dimensions on the receiver
// so Update can use them for mouse offset—no need to reassign the return value.
//
//	return app.swatch.ViewWithOverlay(mainView, width, height)
func (s *SwatchPicker) ViewWithOverlay(mainView string, viewWidth, viewHeight int) string {
	if !s.open {
		return mainView
	}
	modalContent := s.picker.View()
	modalW, overlayHeight := overlay.ModalCellSize(modalContent)
	centerRow := s.row + s.h/2
	centerCol := s.col + s.w/2
	topPad, leftPad := overlay.Fixed(centerRow-overlayHeight/2, centerCol-modalW/2).
		ClampedOrigin(modalW, overlayHeight, viewWidth, viewHeight)
	s.lastOverlayLeft = leftPad
	s.lastOverlayTop = topPad
	s.lastModalW = modalW
	s.lastOverlayHeight = overlayHeight
	s.lastViewWidth = viewWidth
	s.lastViewHeight = viewHeight
	return overlay.OverlayView(mainView, modalContent, viewWidth, viewHeight, topPad, leftPad)
}

// Update handles messages. Forward all tea.Msg to it. When the user picks a color,
// you'll receive ColorChosenMsg; call SetColor(msg.Color) and assign the returned
// model back. When the picker is open, Update forwards to the picker with the
// correct mouse offset.
func (s *SwatchPicker) Update(msg tea.Msg) (*SwatchPicker, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.lastViewWidth = m.Width
		s.lastViewHeight = m.Height
		if s.open {
			picker, cmd := s.picker.Update(tea.WindowSizeMsg{Width: 42, Height: 22})
			s.picker = picker.(Model)
			// Recompute overlay position so mouse coords stay correct after resize
			if s.lastOverlayHeight > 0 && s.lastModalW > 0 {
				centerRow := s.row + s.h/2
				centerCol := s.col + s.w/2
				topPad, leftPad := overlay.Fixed(centerRow-s.lastOverlayHeight/2, centerCol-s.lastModalW/2).
					ClampedOrigin(s.lastModalW, s.lastOverlayHeight, s.lastViewWidth, s.lastViewHeight)
				s.lastOverlayLeft = leftPad
				s.lastOverlayTop = topPad
			}
			return s, cmd
		}
		return s, nil

	case tea.KeyMsg:
		if s.open {
			updated, cmd := s.picker.Update(m)
			s.picker = updated.(Model)
			return s, cmd
		}
		return s, nil

	case tea.MouseMsg:
		if s.open {
			// Ignore the first left-button release after we opened (same click that opened the modal).
			// Otherwise the release is delivered to the picker, lands on the grid, and immediately confirms.
			if s.ignoreNextRelease && m.Action == tea.MouseActionRelease && m.Button == tea.MouseButtonLeft {
				next := *s
				next.ignoreNextRelease = false
				return &next, nil
			}
			leftPad := s.lastOverlayLeft
			topPad := s.lastOverlayTop
			if s.lastModalW <= 0 && s.lastViewWidth > 0 {
				leftPad = max((s.lastViewWidth-44)/2, 0)
			}
			if s.lastOverlayHeight <= 0 && s.lastViewHeight > 0 {
				topPad = max((s.lastViewHeight-22)/2, 0)
			}
			inModal := overlay.CellInModal(m.X, m.Y, topPad, leftPad, s.lastModalW, s.lastOverlayHeight)
			if !inModal {
				return s, nil
			}
			// When using zones, picker gets raw screen coords (zones are registered from full view).
			if s.zoneManager != nil {
				updated, cmd := s.picker.Update(m)
				s.picker = updated.(Model)
				return s, cmd
			}
			relX, relY := MouseToModalCoords(m.X, m.Y, leftPad, topPad)
			relMsg := tea.MouseMsg{
				X: relX, Y: relY,
				Button: m.Button, Action: m.Action, Alt: m.Alt, Ctrl: m.Ctrl, Shift: m.Shift,
			}
			updated, cmd := s.picker.Update(relMsg)
			s.picker = updated.(Model)
			return s, cmd
		}
		if m.Action == tea.MouseActionPress && m.Button == tea.MouseButtonLeft {
			// When zoneManager is set, the app only forwards to us when the zone was in bounds,
			// so we must not re-check bounds (zone covers e.g. "Color 1: ■▼", not just the 2-cell swatch).
			// When zoneManager is nil, use Bubble Tea 0-based cell coords (half-open ranges).
			inBounds := s.zoneManager != nil ||
				(m.X >= s.col && m.X < s.col+s.w && m.Y >= s.row && m.Y < s.row+s.h)
			if inBounds {
				next := *s
				next.picker = next.newPickerModel()
				picker, cmd := next.picker.Update(tea.WindowSizeMsg{Width: 42, Height: 22})
				next.picker = picker.(Model)
				next.open = true
				next.ignoreNextRelease = true // ignore same-click release so it doesn't confirm
				// Compute overlay position so first mouse event uses correct offset (no fallback)
				modalW, overlayHeight := next.picker.ViewSize()
				centerRow := next.row + next.h/2
				centerCol := next.col + next.w/2
				topPad, leftPad := overlay.Fixed(centerRow-overlayHeight/2, centerCol-modalW/2).
					ClampedOrigin(modalW, overlayHeight, next.lastViewWidth, next.lastViewHeight)
				next.lastOverlayLeft = leftPad
				next.lastOverlayTop = topPad
				next.lastModalW = modalW
				next.lastOverlayHeight = overlayHeight
				return &next, cmd
			}
		}
		return s, nil

	case ColorChosenMsg:
		if !s.open {
			return s, nil
		}
		next := *s
		next.color = m.Color
		next.open = false
		return &next, nil

	case ColorCanceledMsg:
		if !s.open {
			return s, nil
		}
		next := *s
		next.open = false
		return &next, nil
	}

	if s.open {
		updated, cmd := s.picker.Update(msg)
		s.picker = updated.(Model)
		return s, cmd
	}
	return s, nil
}

// MouseToModalCoords converts normalized Bubble Tea screen coords (0-based X and Y, same as
// overlay.CellInModal) to coordinates relative to the modal’s top-left cell, in the form the
// picker’s zone handlers expect (contentCol = relX-1, contentRow = relY-1 after Pos-style offsets).
func MouseToModalCoords(screenX, screenY, overlayLeft, overlayTop int) (relX, relY int) {
	relY = screenY - overlayTop + 1
	relX = screenX - overlayLeft + 2
	return relX, relY
}
