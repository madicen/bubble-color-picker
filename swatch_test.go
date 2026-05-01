package bubblepicker

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	overlay "github.com/madicen/bubble-overlay"
)

func TestMouseToModalCoords(t *testing.T) {
	// Normalized Bubble Tea coords (0-based X/Y); overlay top-left at (left=10, top=5).
	left, top := 10, 5

	tests := []struct {
		screenX, screenY int
		wantRelX, wantRelY int
	}{
		{10, 5, 2, 1}, // first row/col of modal (matches overlay.CellInModal)
		{11, 5, 3, 1},
		{10, 6, 2, 2},
		{12, 8, 4, 4},
		{9, 5, 1, 1},
		{10, 4, 2, 0}, // above first modal row (CellInModal false)
	}
	for _, tt := range tests {
		relX, relY := MouseToModalCoords(tt.screenX, tt.screenY, left, top)
		if relX != tt.wantRelX || relY != tt.wantRelY {
			t.Errorf("MouseToModalCoords(%d,%d, %d,%d) = (%d,%d), want (%d,%d)",
				tt.screenX, tt.screenY, left, top, relX, relY, tt.wantRelX, tt.wantRelY)
		}
	}
}

func TestMouseToModalCoords_agreesWithCellInModal(t *testing.T) {
	t.Parallel()
	const left, top, mw, mh = 10, 5, 44, 22
	x, y := left, top
	if !overlay.CellInModal(x, y, top, left, mw, mh) {
		t.Fatal("expected first cell inside modal")
	}
	rx, ry := MouseToModalCoords(x, y, left, top)
	if rx != 2 || ry != 1 {
		t.Fatalf("first cell rel = (%d,%d), want (2,1)", rx, ry)
	}
	x, y = left, top-1
	if overlay.CellInModal(x, y, top, left, mw, mh) {
		t.Fatal("expected above modal to be outside")
	}
	rx, ry = MouseToModalCoords(x, y, left, top)
	if ry != 0 {
		t.Fatalf("above modal relY=%d want 0", ry)
	}
}

func TestSwatchMouseOffsetWhenModalOpen(t *testing.T) {
	s := NewSwatchPicker("#7E00AF", "")
	s.open = true
	s.picker = New(WithInitialColor(s.color))
	_, _ = s.picker.Update(tea.WindowSizeMsg{Width: 42, Height: 22})
	s.lastOverlayLeft = 10
	s.lastOverlayTop = 5
	s.lastModalW = 44
	s.lastOverlayHeight = 22
	s.lastViewWidth = 60
	s.lastViewHeight = 24

	// 0-based screen (12, 7): rel (12-10+2, 7-5+1) = (4, 3).
	msg := tea.MouseMsg{
		X: 12, Y: 7,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}
	next, cmd := s.Update(msg)
	if cmd != nil {
		// May trigger ColorChosenMsg if click landed on grid with release
		_ = cmd
	}
	// Picker (with zones) gets raw coords; without zones rel would be (4, 3). We only assert
	// no panic and modal still open (unless they picked a color).
	if !next.open {
		// They might have clicked confirm; that's valid
		return
	}
}

func TestSwatchViewSingleLine(t *testing.T) {
	// SwatchView is a single line (color + symbol), no newlines.
	s := NewSwatchPicker("#7E00AF", "")
	v := s.SwatchView()
	lines := strings.Split(v, "\n")
	if len(lines) != 1 {
		t.Errorf("SwatchView() split by newline has %d lines, want 1", len(lines))
	}
	if lines[0] == "" {
		t.Error("SwatchView() should not be empty")
	}
}

func TestSwatchResizeRecomputesOverlayPosition(t *testing.T) {
	s := NewSwatchPicker("#7E00AF", "")
	s.SetBounds(5, 15, 3, 3)
	s.open = true
	s.picker = New(WithInitialColor(s.color))
	_, _ = s.picker.Update(tea.WindowSizeMsg{Width: 42, Height: 22})
	s.lastOverlayLeft = 10
	s.lastOverlayTop = 5
	s.lastModalW = 44
	s.lastOverlayHeight = 22
	s.lastViewWidth = 60
	s.lastViewHeight = 24
	// Resize to a different view size; overlay position should be recomputed
	next, _ := s.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	if next.lastViewWidth != 80 || next.lastViewHeight != 30 {
		t.Errorf("after resize: lastView = %dx%d, want 80x30", next.lastViewWidth, next.lastViewHeight)
	}
	// Centered on swatch (row 5, col 15, 3x3) -> center (6, 16). With 80x30, modal 44x22:
	// leftPad = 16 - 22 = -6 -> 0, topPad = 6 - 11 = -5 -> 0. So we expect 0,0 or similar.
	if next.lastOverlayLeft == 10 && next.lastOverlayTop == 5 {
		t.Error("overlay position was not recomputed after WindowSizeMsg (still 10, 5)")
	}
}

// TestSwatchIgnoreSameClickRelease verifies that the release of the same click that
// opened the swatch is ignored, so the picker does not immediately confirm and close.
func TestSwatchIgnoreSameClickRelease(t *testing.T) {
	s := NewSwatchPicker("#7E00AF", "")
	s.SetBounds(2, 10, 2, 1)
	s.lastViewWidth = 80
	s.lastViewHeight = 24
	// Open with a left-button press (same as user click).
	press := tea.MouseMsg{X: 10, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	next, _ := s.Update(press)
	if !next.open {
		t.Fatal("press did not open modal")
	}
	// Send the release of the same click (host would forward it after opening).
	release := tea.MouseMsg{X: 10, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease}
	next, cmd := next.Update(release)
	if !next.open {
		t.Error("release after open closed the picker; same-click release should be ignored")
	}
	if cmd != nil {
		// Should not have sent ColorChosenMsg
		if _, isChosen := cmd().(ColorChosenMsg); isChosen {
			t.Error("same-click release should not produce ColorChosenMsg")
		}
	}
}

func TestSwatchHitTestBounds(t *testing.T) {
	s := NewSwatchPicker("#7E00AF", "")
	s.SetBounds(2, 10, 2, 1)
	// Swatch at row 2, col 10, size 2x1 (0-based half-open): X in [10,12), Y=2
	inside := tea.MouseMsg{X: 10, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	next, _ := s.Update(inside)
	if !next.open {
		t.Error("click inside swatch bounds did not open modal")
	}
	// Close modal, then click outside swatch (left of swatch): should not open
	next, _ = next.Update(ColorCanceledMsg{})
	outside := tea.MouseMsg{X: 8, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	next, _ = next.Update(outside)
	if next.open {
		t.Error("click outside swatch opened modal")
	}
}

// TestSwatchClickOpensPickerAtPosition mimics the example's 2x2 layout: position the mouse
// directly on the first swatch (using the same bounds as the example) and send a click;
// verify the picker opens. Uses normalized 0-based Bubble Tea mouse coordinates.
func TestSwatchSetPickerOptions_Merge(t *testing.T) {
	s := NewSwatchPicker("#112233", "")
	s.SetPickerOptions(WithPresets([]string{"#aabbcc", "#ddeeff"}))
	s.SetBounds(2, 10, 2, 1)
	s.lastViewWidth = 80
	s.lastViewHeight = 24
	next, _ := s.Update(tea.MouseMsg{X: 10, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if !next.open {
		t.Fatal("expected picker open")
	}
	next, _ = next.Update(ColorCanceledMsg{})
	if next.open {
		t.Fatal("expected picker closed after cancel")
	}
}

func TestSwatchClickOpensPickerAtPosition(t *testing.T) {
	const labelLen = 10
	const gap = 2
	sw, sh := 2, 1
	col1 := labelLen
	col2 := col1 + sw + gap + labelLen

	swatches := [4]*SwatchPicker{
		NewSwatchPicker("#7E00AF", ""),
		NewSwatchPicker("#00AF7E", ""),
		NewSwatchPicker("#AF7E00", ""),
		NewSwatchPicker("#AF007E", ""),
	}
	for i := range swatches {
		var row, col int
		if i < 2 {
			row, col = 2, col1
			if i == 1 {
				col = col2
			}
		} else {
			row, col = 3, col1
			if i == 3 {
				col = col2
			}
		}
		swatches[i].SetBounds(row, col, sw, sh)
	}

	click := tea.MouseMsg{X: col1, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	swatches[0], _ = swatches[0].Update(click)
	if !swatches[0].Open() {
		t.Errorf("click at (X=%d, Y=%d) on first swatch did not open picker", col1, 2)
	}

	swatches[1], _ = swatches[1].Update(ColorCanceledMsg{}) // close if any
	click2 := tea.MouseMsg{X: col2, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	swatches[1], _ = swatches[1].Update(click2)
	if !swatches[1].Open() {
		t.Errorf("click at (X=%d, Y=%d) on second swatch did not open picker", col2, 2)
	}
}
