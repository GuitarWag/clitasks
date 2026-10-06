// Package clitasks holds the files at the repository root that the binary
// embeds. go:embed cannot reach outside a package's directory, so the
// canonical SKILL.md is embedded here, where it lives.
package clitasks

import _ "embed"

// Skill is SKILL.md, the instructions that `tasks claude` and `tasks codex`
// install for AI agents.
//
//go:embed SKILL.md
var Skill []byte
