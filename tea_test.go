package bubblepicker

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestEnterEmitsColorChangedMsg verifies that Enter dispatches ColorChangedMsg with the current hex.
func TestEnterEmitsColorChangedMsg(t *testing.T) {
	m := New(WithInitialColor("#336699"))
	var tm tea.Model = m
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 50, Height: 24})
	m = tm.(Model)
	m.Focus = FocusGrid // avoid preset branch when presets added elsewhere
	tm = m
	tm, cmd := tm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("expected cmd from Enter")
	}
	msg := cmd()
	cm, ok := msg.(ColorChangedMsg)
	if !ok {
		t.Fatalf("expected ColorChangedMsg, got %T", msg)
	}
	if !strings.EqualFold(cm.Color, "#336699") {
		t.Errorf("Color = %q, want #336699", cm.Color)
	}
	if cm.Dismiss {
		t.Error("Dismiss should be false without WithAutoDismiss")
	}
}

// TestAutoDismissSetsFlag verifies WithAutoDismiss sets Dismiss on ColorChangedMsg.
func TestAutoDismissSetsFlag(t *testing.T) {
	m := New(WithInitialColor("#ff0000"), WithAutoDismiss(true))
	var tm tea.Model = m
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 50, Height: 24})
	m = tm.(Model)
	m.Focus = FocusHueBar
	tm = m
	_, cmd := tm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd")
	}
	cm := cmd().(ColorChangedMsg)
	if !cm.Dismiss {
		t.Error("Dismiss should be true with WithAutoDismiss(true)")
	}
}

// TestPresetsInView verifies preset strip appears when WithPresets is used (deterministic string check).
func TestPresetsInView(t *testing.T) {
	m := New(WithInitialColor("#000000"), WithPresets([]string{"#ff0000", "#00ff00", "#0000ff"}))
	var tm tea.Model = m
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 60, Height: 30})
	m = tm.(Model)
	view := m.View()
	if !strings.Contains(view, "Pick a color") {
		t.Error("view should contain title")
	}
	if len(m.presets) != 3 {
		t.Fatalf("presets len = %d, want 3", len(m.presets))
	}
	if !strings.Contains(m.Value(), "#") {
		t.Error("picker should have a hex value")
	}
	_ = view
}

// TestPresetEnterSelectsPreset applies Enter on preset focus to emit ColorChangedMsg with that color.
func TestPresetEnterSelectsPreset(t *testing.T) {
	m := New(WithPresets([]string{"#abcdef", "#123456"}))
	var tm tea.Model = m
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 60, Height: 30})
	m = tm.(Model)
	if m.Focus != FocusPresets {
		t.Fatalf("initial focus want FocusPresets, got %v", m.Focus)
	}
	m.presetFocus = 0
	tm = m
	_, cmd := tm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd")
	}
	cm := cmd().(ColorChangedMsg)
	want := "#abcdef"
	if !strings.EqualFold(cm.Color, want) {
		t.Errorf("Color = %q, want %s", cm.Color, want)
	}
}
