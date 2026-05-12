package mcp

import (
	"fmt"
	"strings"

	"github.com/jlentink/twinmind-mcp/internal/api"
)

func FormatRecordingList(recordings []api.RecordingTitle) string {
	if len(recordings) == 0 {
		return "No recordings found."
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d recording(s):\n\n", len(recordings)))

	for i, r := range recordings {
		duration := formatDuration(r.Metadata.DurationSeconds)
		b.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, r.Title))
		b.WriteString(fmt.Sprintf("   - Meeting ID: `%s`\n", r.MeetingID))
		b.WriteString(fmt.Sprintf("   - Start: %s\n", r.StartTime.Local().Format("2006-01-02 15:04")))
		b.WriteString(fmt.Sprintf("   - Duration: %s\n", duration))
		b.WriteString(fmt.Sprintf("   - Device: %s\n\n", r.Metadata.DeviceType))
	}

	return b.String()
}

func FormatRecordingDetail(d *api.RecordingDetail) string {
	s := d.Summary
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", s.MeetingTitle))
	b.WriteString(fmt.Sprintf("- **Meeting ID:** `%s`\n", s.MeetingID))
	b.WriteString(fmt.Sprintf("- **Start:** %s\n", s.StartTime.Local().Format("2006-01-02 15:04")))
	b.WriteString(fmt.Sprintf("- **End:** %s\n", s.EndTime.Local().Format("2006-01-02 15:04")))
	b.WriteString(fmt.Sprintf("- **Status:** %s\n\n", s.Status))

	if s.Summary != "" {
		b.WriteString("## Summary\n\n")
		b.WriteString(s.Summary)
		b.WriteString("\n\n")
	}

	if s.Action != "" {
		b.WriteString("## Action Items\n\n")
		b.WriteString(s.Action)
		b.WriteString("\n\n")
	}

	if d.Notes.Notes != "" && d.Notes.Notes != "notes: " {
		b.WriteString("## Notes\n\n")
		b.WriteString(d.Notes.Notes)
		b.WriteString("\n\n")
	}

	if s.Transcript != "" {
		b.WriteString("## Transcript\n\n")
		b.WriteString(s.Transcript)
		b.WriteString("\n")
	}

	return b.String()
}

func FormatRecordingSection(d *api.RecordingDetail, section string) string {
	switch section {
	case "transcript":
		return d.Summary.Transcript
	case "summary":
		return d.Summary.Summary
	case "action":
		return d.Summary.Action
	case "notes":
		return d.Notes.Notes
	default:
		return fmt.Sprintf("Unknown section %q. Valid options: transcript, summary, action, notes.", section)
	}
}

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "unknown"
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
