package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jlentink/twinmind-mcp/internal/api"
)

type clearFlashMsg struct{}

type recordingsLoadedMsg struct {
	recordings []api.RecordingTitle
	err        error
}

type recordingDetailLoadedMsg struct {
	detail    *api.RecordingDetail
	meetingID string
	err       error
}

func fetchRecordingsCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		recordings, err := client.ListRecordings(context.Background(), 50, 0)
		return recordingsLoadedMsg{recordings: recordings, err: err}
	}
}

func fetchDetailCmd(client *api.Client, meetingID string) tea.Cmd {
	return func() tea.Msg {
		detail, err := client.GetRecording(context.Background(), meetingID)
		return recordingDetailLoadedMsg{detail: detail, meetingID: meetingID, err: err}
	}
}

func clearFlashAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearFlashMsg{}
	})
}
