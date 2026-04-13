package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func (c *Client) ListRecordings(ctx context.Context, limit, offset int) ([]RecordingTitle, error) {
	req := GetMemoryTitlesRequest{
		DistinctByMeeting: true,
		Limit:             limit,
		Offset:            offset,
	}

	data, err := c.do(ctx, "POST", "/api/v1/get_memory_titles", req)
	if err != nil {
		return nil, err
	}

	var resp TitlesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode recordings: %w", err)
	}

	return resp.Memories, nil
}

func (c *Client) GetRecording(ctx context.Context, meetingID string) (*RecordingDetail, error) {
	req := GetMemoryRequest{
		MeetingID: meetingID,
	}

	data, err := c.do(ctx, "POST", "/api/v1/get_memory", req)
	if err != nil {
		return nil, err
	}

	var resp DetailResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode recording: %w", err)
	}

	if len(resp.Memories) == 0 {
		return nil, fmt.Errorf("recording not found: %s", meetingID)
	}

	return &resp.Memories[0], nil
}

func (c *Client) SearchRecordings(ctx context.Context, query string) ([]RecordingTitle, error) {
	all, err := c.fetchAllTitles(ctx)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []RecordingTitle
	for _, r := range all {
		if strings.Contains(strings.ToLower(r.Title), query) {
			results = append(results, r)
		}
	}

	return results, nil
}

func (c *Client) fetchAllTitles(ctx context.Context) ([]RecordingTitle, error) {
	const pageSize = 100
	var all []RecordingTitle

	for offset := 0; ; offset += pageSize {
		page, err := c.ListRecordings(ctx, pageSize, offset)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if len(page) < pageSize {
			break
		}
	}

	return all, nil
}
