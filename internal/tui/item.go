package tui

import (
	"fmt"

	"github.com/jlentink/twinmind-mcp/internal/api"
)

type recordingItem struct {
	recording api.RecordingTitle
}

func (r recordingItem) FilterValue() string { return r.recording.Title }
func (r recordingItem) Title() string       { return r.recording.Title }
func (r recordingItem) Description() string {
	duration := formatDuration(r.recording.Metadata.DurationSeconds)
	return r.recording.StartTime.Local().Format("2006-01-02 15:04") + "  " + duration
}

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "-"
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
