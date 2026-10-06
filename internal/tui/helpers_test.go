package tui

import (
	tea "charm.land/bubbletea/v2"
)

// keyMsg builds a key press from its string form, such as "j", "tab",
// "ctrl+k", "shift+tab" or "alt+>".
func keyMsg(s string) tea.KeyPressMsg {
	special := map[string]rune{
		"esc": tea.KeyEscape, "enter": tea.KeyEnter, "tab": tea.KeyTab, "space": tea.KeySpace,
		"up": tea.KeyUp, "down": tea.KeyDown, "left": tea.KeyLeft, "right": tea.KeyRight,
		"backspace": tea.KeyBackspace, "home": tea.KeyHome, "end": tea.KeyEnd,
		"pgup": tea.KeyPgUp, "pgdown": tea.KeyPgDown,
	}
	var mod tea.KeyMod
	for {
		switch {
		case len(s) > 5 && s[:5] == "ctrl+":
			mod |= tea.ModCtrl
			s = s[5:]
			continue
		case len(s) > 4 && s[:4] == "alt+":
			mod |= tea.ModAlt
			s = s[4:]
			continue
		case len(s) > 6 && s[:6] == "shift+":
			mod |= tea.ModShift
			s = s[6:]
			continue
		}
		break
	}
	if c, ok := special[s]; ok {
		msg := tea.KeyPressMsg{Code: c, Mod: mod}
		if c == tea.KeySpace {
			msg.Text = " "
		}
		return msg
	}
	r := []rune(s)
	msg := tea.KeyPressMsg{Code: r[0], Mod: mod}
	if mod == 0 {
		msg.Text = s
	}
	return msg
}

func send(m Model, keys ...string) Model {
	for _, k := range keys {
		mm, _ := m.Update(keyMsg(k))
		m = mm.(Model)
	}
	return m
}
