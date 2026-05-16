package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var idRe = regexp.MustCompile(`T-[0-9A-Z]+-[0-9A-Z]{3}`)

func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd("test")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func withBoardFile(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "board.md")
	t.Setenv("TASK_BOARD_FILE", p)
	t.Cleanup(func() { flagFile = "" })
	return p
}

func TestInit(t *testing.T) {
	p := withBoardFile(t)
	out, err := runCmd(t, "init", "-n", "Smoke", "-d", "desc")
	require.NoError(t, err)
	assert.Contains(t, out, "Board initialized")

	data, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Contains(t, string(data), "# Board: Smoke")
	assert.Contains(t, string(data), "Description: desc")
}

func TestAdd_persists(t *testing.T) {
	p := withBoardFile(t)
	out, err := runCmd(t, "add", "first", "-p", "high", "-t", "a,b", "-a", "alice", "--due", "2026-06-01")
	require.NoError(t, err)
	assert.Contains(t, out, "Task added")

	data, _ := os.ReadFile(p)
	s := string(data)
	assert.Contains(t, s, "**first**")
	assert.Contains(t, s, "`priority:high`")
	assert.Contains(t, s, "`assignee:alice`")
	assert.Contains(t, s, "`tags:a,b`")
	assert.Contains(t, s, "`due:2026-06-01`")
}

func TestAdd_invalidPriority(t *testing.T) {
	_ = withBoardFile(t)
	_, err := runCmd(t, "add", "x", "-p", "bogus")
	assert.Error(t, err)
}

func TestList_filters(t *testing.T) {
	_ = withBoardFile(t)
	_, err := runCmd(t, "add", "a", "-a", "alice")
	require.NoError(t, err)
	_, err = runCmd(t, "add", "b", "-a", "bob")
	require.NoError(t, err)

	out, err := runCmd(t, "list", "-a", "alice")
	require.NoError(t, err)
	assert.Contains(t, out, "Found 1 task")
	assert.Contains(t, out, "a")

	out, err = runCmd(t, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "Found 2 task")
}

func TestList_noResults(t *testing.T) {
	_ = withBoardFile(t)
	out, err := runCmd(t, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "No tasks found")
}

func TestBoardView(t *testing.T) {
	_ = withBoardFile(t)
	_, _ = runCmd(t, "add", "one")
	out, err := runCmd(t, "board")
	require.NoError(t, err)
	assert.Contains(t, out, "TODO (1)")
	assert.Contains(t, out, "IN PROGRESS (0)")
	assert.Contains(t, out, "DONE (0)")
	assert.Contains(t, out, "BLOCKED (0)")
}

func TestShow_andNotFound(t *testing.T) {
	_ = withBoardFile(t)
	_, _ = runCmd(t, "add", "task one")

	out, _ := runCmd(t, "list")
	id := idRe.FindString(out)
	require.NotEmpty(t, id)

	showOut, err := runCmd(t, "show", id)
	require.NoError(t, err)
	assert.Contains(t, showOut, "task one")

	missingOut, err := runCmd(t, "show", "T-NOPE-XXX")
	require.NoError(t, err)
	assert.Contains(t, missingOut, "Task not found")
}

func extractID(t *testing.T, listOut string) string {
	t.Helper()
	id := idRe.FindString(listOut)
	require.NotEmpty(t, id, "no ID in output")
	return id
}

var _ = strings.Contains

func TestUpdate_and_NotFound(t *testing.T) {
	_ = withBoardFile(t)
	_, _ = runCmd(t, "add", "x")
	listOut, _ := runCmd(t, "list")
	id := extractID(t, listOut)

	out, err := runCmd(t, "update", id, "-t", "renamed", "-p", "critical")
	require.NoError(t, err)
	assert.Contains(t, out, "Task updated")
	assert.Contains(t, out, "renamed")

	out, err = runCmd(t, "update", "T-NOPE-XXX", "-t", "y")
	require.NoError(t, err)
	assert.Contains(t, out, "Task not found")
}

func TestMoveShortcuts(t *testing.T) {
	_ = withBoardFile(t)
	_, _ = runCmd(t, "add", "x")
	id := extractID(t, mustList(t))

	for _, tc := range []struct {
		cmd, expect string
	}{
		{"start", "Started"},
		{"complete", "completed"},
		{"block", "blocked"},
	} {
		out, err := runCmd(t, tc.cmd, id)
		require.NoError(t, err)
		assert.Contains(t, out, tc.expect)
	}

	out, err := runCmd(t, "move", id, "todo")
	require.NoError(t, err)
	assert.Contains(t, out, "moved to todo")

	out, err = runCmd(t, "move", id, "bogus")
	require.NoError(t, err)
	assert.Contains(t, out, "invalid status")
}

func mustList(t *testing.T) string {
	t.Helper()
	out, err := runCmd(t, "list")
	require.NoError(t, err)
	return out
}

func TestDelete(t *testing.T) {
	_ = withBoardFile(t)
	_, _ = runCmd(t, "add", "x")
	id := extractID(t, mustList(t))

	out, err := runCmd(t, "delete", id)
	require.NoError(t, err)
	assert.Contains(t, out, "Task deleted")

	out, err = runCmd(t, "delete", "T-NOPE-XXX")
	require.NoError(t, err)
	assert.Contains(t, out, "Task not found")
}

func TestInfo(t *testing.T) {
	_ = withBoardFile(t)
	_, _ = runCmd(t, "init", "-n", "Board One")
	_, _ = runCmd(t, "add", "x")

	out, err := runCmd(t, "info")
	require.NoError(t, err)
	assert.Contains(t, out, "Board One")
	assert.Contains(t, out, "Total tasks: ")
}

func TestStats(t *testing.T) {
	_ = withBoardFile(t)
	_, _ = runCmd(t, "add", "x", "-p", "high", "-a", "alice")
	_, _ = runCmd(t, "add", "y", "-p", "low")

	out, err := runCmd(t, "stats")
	require.NoError(t, err)
	assert.Contains(t, out, "Statistics")
	assert.Contains(t, out, "Status Breakdown")
	assert.Contains(t, out, "Priority Breakdown")
	assert.Contains(t, out, "Assignee Breakdown")
	assert.Contains(t, out, "alice")
	assert.Contains(t, out, "Completion Rate")
}

func TestExport_stdoutJSON(t *testing.T) {
	_ = withBoardFile(t)
	_, _ = runCmd(t, "add", "x")
	out, err := runCmd(t, "export", "--format", "json")
	require.NoError(t, err)

	startIdx := strings.Index(out, "{")
	require.GreaterOrEqual(t, startIdx, 0)
	endIdx := strings.LastIndex(out, "}")
	require.Greater(t, endIdx, startIdx)
	jsonStr := out[startIdx : endIdx+1]

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &got))
	tasks, ok := got["tasks"].([]any)
	require.True(t, ok)
	assert.Len(t, tasks, 1)
}

func TestExport_toFile(t *testing.T) {
	_ = withBoardFile(t)
	_, _ = runCmd(t, "add", "x")
	dst := filepath.Join(t.TempDir(), "out.csv")
	out, err := runCmd(t, "export", "--format", "csv", "-o", dst)
	require.NoError(t, err)
	assert.Contains(t, out, "Exported to "+dst)

	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Contains(t, string(data), "ID,Title")
}

func TestExport_invalidFormat(t *testing.T) {
	_ = withBoardFile(t)
	out, err := runCmd(t, "export", "--format", "xml")
	require.NoError(t, err)
	assert.Contains(t, out, "invalid format")
}

func TestSkillInstall_local(t *testing.T) {
	dir := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	require.NoError(t, os.Chdir(dir))

	out, err := runCmd(t, "claude")
	require.NoError(t, err)
	assert.Contains(t, out, "SKILL.md installed")

	target := filepath.Join(dir, ".claude", "skills", "SKILL.md")
	info, err := os.Stat(target)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))

	out, err = runCmd(t, "codex")
	require.NoError(t, err)
	assert.Contains(t, out, "SKILL.md installed")
	_, err = os.Stat(filepath.Join(dir, ".codex", "skills", "SKILL.md"))
	require.NoError(t, err)
}

func TestGlobalFileFlag_directPath(t *testing.T) {
	t.Setenv("TASK_BOARD_FILE", "")
	t.Cleanup(func() { flagFile = "" })
	p := filepath.Join(t.TempDir(), "explicit.md")

	out, err := runCmd(t, "-f", p, "init", "-n", "Explicit")
	require.NoError(t, err)
	assert.Contains(t, out, "Board initialized")

	_, err = os.Stat(p)
	require.NoError(t, err, "the explicit file should have been created")
}

func TestList_invalidStatusFlag(t *testing.T) {
	_ = withBoardFile(t)
	_, err := runCmd(t, "list", "-s", "bogus")
	assert.Error(t, err, "invalid status flag should error")
}

func TestList_combinedFilters(t *testing.T) {
	_ = withBoardFile(t)
	_, _ = runCmd(t, "add", "alpha", "-p", "high", "-a", "alice", "-t", "x")
	_, _ = runCmd(t, "add", "beta", "-p", "high", "-a", "bob", "-t", "x")
	_, _ = runCmd(t, "add", "gamma", "-p", "low", "-a", "alice", "-t", "x")

	out, err := runCmd(t, "list", "-p", "high", "-a", "alice")
	require.NoError(t, err)
	assert.Contains(t, out, "Found 1 task")
	assert.Contains(t, out, "alpha")
	assert.NotContains(t, out, "beta")
	assert.NotContains(t, out, "gamma")
}

func TestList_detailed(t *testing.T) {
	_ = withBoardFile(t)
	_, _ = runCmd(t, "add", "x", "-d", "the description body")
	out, err := runCmd(t, "list", "--detailed")
	require.NoError(t, err)
	assert.Contains(t, out, "the description body")
	assert.Contains(t, out, "Created:")
}

func TestSkillInstall_global(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	require.NoError(t, os.Chdir(dir))

	out, err := runCmd(t, "claude", "-g")
	require.NoError(t, err)
	assert.Contains(t, out, "SKILL.md installed")

	_, err = os.Stat(filepath.Join(home, ".claude", "skills", "tasks-cli", "SKILL.md"))
	require.NoError(t, err)
}
