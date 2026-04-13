package api

import "time"

type GetMemoryTitlesRequest struct {
	DistinctByMeeting bool `json:"distinctByMeeting"`
	Limit             int  `json:"limit"`
	Offset            int  `json:"offset"`
}

type GetMemoryRequest struct {
	MeetingID string `json:"meeting_id"`
}

type RecordingMetadata struct {
	DurationSeconds int    `json:"durationSeconds"`
	DeviceType      string `json:"deviceType"`
}

type RecordingTitle struct {
	Title     string            `json:"title"`
	MeetingID string            `json:"meeting_id"`
	SummaryID string            `json:"summary_id"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Metadata  RecordingMetadata `json:"metadata"`
}

type RecordingSummary struct {
	MeetingTitle string    `json:"meeting_title"`
	Summary      string    `json:"summary"`
	Action       string    `json:"action"`
	Transcript   string    `json:"transcript"`
	Keywords     string    `json:"keywords"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Status       string    `json:"status"`
	MeetingID    string    `json:"meeting_id"`
}

type RecordingNotes struct {
	MeetingID string `json:"meeting_id"`
	Notes     string `json:"notes"`
}

type RecordingDetail struct {
	Summary RecordingSummary `json:"summary"`
	Notes   RecordingNotes   `json:"notes"`
}

type TitlesResponse struct {
	Memories []RecordingTitle `json:"memories"`
}

type DetailResponse struct {
	Memories []RecordingDetail `json:"memories"`
}
