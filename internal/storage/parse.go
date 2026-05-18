package storage

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
	"time"

	"github.com/GuitarWag/clitasks/internal/model"
)

var (
	taskLineRe       = regexp.MustCompile(`^- \[.\] \[(.+?)\] \*\*(.+?)\*\*(.*)$`)
	priorityRe       = regexp.MustCompile("`priority:(\\w+)`")
	assigneeRe       = regexp.MustCompile("`assignee:([^`]+)`")
	tagsRe           = regexp.MustCompile("`tags:([^`]+)`")
	dueRe            = regexp.MustCompile("`due:([^`]+)`")
	createdUpdatedRe = regexp.MustCompile(`Created:\s*(\S+).*Updated:\s*(\S+)`)
	createdOnlyRe    = regexp.MustCompile(`Created:\s*(\S+)`)
	updatedOnlyRe    = regexp.MustCompile(`Updated:\s*(\S+)`)
	// Task metadata lines emitted by the renderer always look like
	//   "  > Created: <RFC3339Nano>[ | Updated: <RFC3339Nano>]"
	// so we anchor on the year prefix to avoid swallowing descriptions
	// that happen to contain the words "Created" or "Updated".
	taskMetaLineRe = regexp.MustCompile(`^\s*>\s*(?:Created|Updated):\s*\d{4}-\d{2}-\d{2}T`)
)

func parseMarkdown(data []byte, clock func() time.Time) *model.Board {
	now := clock().UTC()
	b := &model.Board{
		Name:      "My Board",
		Tasks:     []model.Task{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	var (
		section model.TaskStatus
		hasSec  bool
		cur     *model.Task
	)

	flush := func() {
		if cur != nil && cur.ID != "" {
			b.Tasks = append(b.Tasks, *cur)
			cur = nil
		}
	}

	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()

		switch {
		case strings.HasPrefix(line, "# Board:"):
			b.Name = strings.TrimSpace(strings.TrimPrefix(line, "# Board:"))
			continue
		case strings.HasPrefix(line, "> Description:"):
			b.Description = strings.TrimSpace(strings.TrimPrefix(line, "> Description:"))
			continue
		case strings.HasPrefix(line, "> Created:"):
			meta := strings.TrimSpace(strings.TrimPrefix(line, "> Created:"))
			if strings.Contains(meta, "|") {
				parts := strings.SplitN(meta, "|", 2)
				if t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(parts[0])); err == nil {
					b.CreatedAt = t
				}
				if strings.Contains(parts[1], "Updated:") {
					upd := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts[1]), "Updated:"))
					if t, err := time.Parse(time.RFC3339Nano, upd); err == nil {
						b.UpdatedAt = t
					}
				}
			} else {
				if t, err := time.Parse(time.RFC3339Nano, meta); err == nil {
					b.CreatedAt = t
				}
			}
			continue
		case strings.HasPrefix(line, "> Updated:"):
			upd := strings.TrimSpace(strings.TrimPrefix(line, "> Updated:"))
			if t, err := time.Parse(time.RFC3339Nano, upd); err == nil {
				b.UpdatedAt = t
			}
			continue
		}

		switch {
		case strings.HasPrefix(line, "## TODO"):
			flush()
			section, hasSec = model.StatusTodo, true
			continue
		case strings.HasPrefix(line, "## IN PROGRESS"):
			flush()
			section, hasSec = model.StatusInProgress, true
			continue
		case strings.HasPrefix(line, "## DONE"):
			flush()
			section, hasSec = model.StatusDone, true
			continue
		case strings.HasPrefix(line, "## BLOCKED"):
			flush()
			section, hasSec = model.StatusBlocked, true
			continue
		}

		if hasSec && (strings.HasPrefix(line, "- [ ]") ||
			strings.HasPrefix(line, "- [x]") ||
			strings.HasPrefix(line, "- [>]") ||
			strings.HasPrefix(line, "- [!]")) {
			flush()
			m := taskLineRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			t := &model.Task{
				ID:        m[1],
				Title:     m[2],
				Status:    section,
				Priority:  model.PriorityMedium,
				CreatedAt: now,
				UpdatedAt: now,
			}
			meta := m[3]
			if mm := priorityRe.FindStringSubmatch(meta); mm != nil {
				if p, err := model.ParsePriority(mm[1]); err == nil {
					t.Priority = p
				}
			}
			if mm := assigneeRe.FindStringSubmatch(meta); mm != nil {
				t.Assignee = mm[1]
			}
			if mm := tagsRe.FindStringSubmatch(meta); mm != nil {
				parts := strings.Split(mm[1], ",")
				tags := make([]string, 0, len(parts))
				for _, p := range parts {
					if v := strings.TrimSpace(p); v != "" {
						tags = append(tags, v)
					}
				}
				t.Tags = tags
			}
			if mm := dueRe.FindStringSubmatch(meta); mm != nil {
				t.DueDate = mm[1]
			}
			cur = t
			continue
		}

		if cur != nil && strings.HasPrefix(line, "  >") && !taskMetaLineRe.MatchString(line) {
			cur.Description = strings.TrimSpace(strings.TrimPrefix(line, "  >"))
			continue
		}

		if cur != nil && strings.Contains(line, "Created:") && strings.Contains(line, "Updated:") {
			if mm := createdUpdatedRe.FindStringSubmatch(line); mm != nil {
				if t, err := time.Parse(time.RFC3339Nano, mm[1]); err == nil {
					cur.CreatedAt = t
				}
				if t, err := time.Parse(time.RFC3339Nano, mm[2]); err == nil {
					cur.UpdatedAt = t
				}
			}
			continue
		}
		if cur != nil && strings.Contains(line, "Created:") {
			if mm := createdOnlyRe.FindStringSubmatch(line); mm != nil {
				if t, err := time.Parse(time.RFC3339Nano, mm[1]); err == nil {
					cur.CreatedAt = t
				}
			}
		}
		if cur != nil && strings.Contains(line, "Updated:") {
			if mm := updatedOnlyRe.FindStringSubmatch(line); mm != nil {
				if t, err := time.Parse(time.RFC3339Nano, mm[1]); err == nil {
					cur.UpdatedAt = t
				}
			}
		}
	}
	flush()
	return b
}
