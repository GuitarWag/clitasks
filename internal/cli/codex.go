package cli

import "github.com/spf13/cobra"

func newCodexCmd() *cobra.Command {
	var global bool
	cmd := &cobra.Command{
		Use:   "codex",
		Short: "Install SKILL.md to Codex CLI skills directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return installSkill("codex", global, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false,
		"Install to ~/.codex/skills/tasks-cli instead of local .codex/skills")
	return cmd
}
