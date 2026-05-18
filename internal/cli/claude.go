package cli

import "github.com/spf13/cobra"

func newClaudeCmd() *cobra.Command {
	var global bool
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Install SKILL.md to Claude skills directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return installSkill("claude", global, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false,
		"Install to ~/.claude/skills/tasks-cli instead of local .claude/skills/tasks-cli")
	return cmd
}
