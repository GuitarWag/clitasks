// Package icons holds the glyph sets of the UI. Every icon has a Nerd Font,
// a Unicode and an ASCII form, and status and priority icons are always shown
// with a word, so no meaning depends on the glyph alone.
package icons

import (
	"fmt"
	"strings"

	"github.com/GuitarWag/clitasks/internal/model"
)

type Set struct {
	Name string

	Todo, InProgress, Blocked, Done          string
	Critical, High, Medium, Low              string
	Due, Start, After, Assignee, Tag, Warn   string
	Board, Timeline, Search, Today, Selected string
	// Bullet separates items on one line, such as chips on a card.
	Bullet string
}

const (
	NameAuto    = "auto"
	NameNerd    = "nerd"
	NameUnicode = "unicode"
	NameASCII   = "ascii"
)

// Codepoints are from nerd-fonts glyphnames.json (md-* Material Design names).
var Nerd = Set{
	Name: NameNerd,
	Todo: "\U000F0766", InProgress: "\U000F0996", Blocked: "\U000F073A", Done: "\U000F05E0",
	Critical: "\U000F0026", High: "\U000F013F", Medium: "\U000F01FC", Low: "\U000F0140",
	Due: "\U000F00ED", Start: "\U000F040A", After: "\U000F0339", Assignee: "\U000F0004",
	Tag: "\U000F04F9", Warn: "\U000F0028",
	Board: "\U000F056D", Timeline: "\U000F066C", Search: "\U000F0349", Today: "\U000F00F6",
	Selected: "▌", Bullet: "·",
}

var Unicode = Set{
	Name: NameUnicode,
	Todo: "○", InProgress: "◐", Blocked: "⊘", Done: "●",
	Critical: "▲▲", High: "▲", Medium: "■", Low: "▼",
	Due: "◷", Start: "▸", After: "↳", Assignee: "@", Tag: "#", Warn: "⚠",
	Board: "▦", Timeline: "▤", Search: "⌕", Today: "◆",
	Selected: "▌", Bullet: "·",
}

var ASCII = Set{
	Name: NameASCII,
	Todo: "[ ]", InProgress: "[>]", Blocked: "[!]", Done: "[x]",
	Critical: "!!", High: "!", Medium: "=", Low: ".",
	Due: "due", Start: "start", After: "after", Assignee: "@", Tag: "#", Warn: "!",
	Board: "", Timeline: "", Search: "/", Today: "*",
	Selected: ">", Bullet: "|",
}

// nerdTerminals ship Nerd Font symbols built in, so the glyphs render even
// when the user's font has none.
var nerdTerminals = []string{"wezterm", "ghostty"}

// Resolve returns the set for a name. On "" and "auto" it uses Nerd glyphs in
// terminals that bundle them and Unicode elsewhere, because a missing glyph
// shows as a box.
func Resolve(name, termProgram string) (Set, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", NameAuto:
		tp := strings.ToLower(termProgram)
		for _, t := range nerdTerminals {
			if strings.Contains(tp, t) {
				return Nerd, nil
			}
		}
		return Unicode, nil
	case NameNerd:
		return Nerd, nil
	case NameUnicode:
		return Unicode, nil
	case NameASCII:
		return ASCII, nil
	}
	return Set{}, fmt.Errorf("unknown icon set %q (want auto|nerd|unicode|ascii)", name)
}

func (s Set) Status(st model.TaskStatus) string {
	switch st {
	case model.StatusInProgress:
		return s.InProgress
	case model.StatusBlocked:
		return s.Blocked
	case model.StatusDone:
		return s.Done
	default:
		return s.Todo
	}
}

func (s Set) Priority(p model.TaskPriority) string {
	switch p {
	case model.PriorityCritical:
		return s.Critical
	case model.PriorityHigh:
		return s.High
	case model.PriorityLow:
		return s.Low
	default:
		return s.Medium
	}
}
