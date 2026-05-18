package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func installSkill(tool string, global bool, out io.Writer) error {
	var dir string
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(home, "."+tool, "skills", "tasks-cli")
	} else {
		dir = filepath.Join("."+tool, "skills", "tasks-cli")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	target := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(target, skillContent, 0o644); err != nil {
		return err
	}

	scope := "local"
	if global {
		scope = "global"
	}
	fmt.Fprintln(out, styleSuccess.Render(fmt.Sprintf("✓ SKILL.md installed to %s %s skills:", scope, tool)))
	fmt.Fprintln(out, "  "+styleCyan.Render(target))
	if global {
		fmt.Fprintln(out, styleDim.Render(fmt.Sprintf("\nThe skill is now available globally for all %s sessions.", tool)))
	} else {
		fmt.Fprintln(out, styleDim.Render("\nThe skill is available for this project only."))
	}
	return nil
}
