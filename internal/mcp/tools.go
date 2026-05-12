package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) registerTools() {
	s.mcp.AddTool(
		mcp.NewTool("list_recordings",
			mcp.WithDescription("List all TwinMind meeting recordings. Returns title, meeting ID, start/end time, and duration."),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of recordings to return (default: 20, max: 100)"),
			),
			mcp.WithNumber("offset",
				mcp.Description("Number of recordings to skip for pagination (default: 0)"),
			),
		),
		s.handleListRecordings,
	)

	s.mcp.AddTool(
		mcp.NewTool("get_recording",
			mcp.WithDescription("Get full details of a specific recording including summary, action items, transcript, and notes."),
			mcp.WithString("meeting_id",
				mcp.Required(),
				mcp.Description("The UUID meeting_id of the recording to retrieve"),
			),
			mcp.WithString("section",
				mcp.Description("Return only a specific section: 'transcript', 'summary', 'action', or 'notes'. Omit for full details."),
			),
		),
		s.handleGetRecording,
	)

	s.mcp.AddTool(
		mcp.NewTool("search_recordings",
			mcp.WithDescription("Search recordings by keyword. Searches across all recording titles."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Keyword or phrase to search for in recording titles"),
			),
		),
		s.handleSearchRecordings,
	)
}

func (s *Server) handleListRecordings(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 20
	offset := 0

	args := request.GetArguments()
	if v, ok := args["limit"].(float64); ok {
		limit = int(v)
	}
	if v, ok := args["offset"].(float64); ok {
		offset = int(v)
	}

	if limit < 1 || limit > 100 {
		limit = 20
	}

	recordings, err := s.api.ListRecordings(ctx, limit, offset)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to list recordings: %v", err)), nil
	}

	return mcp.NewToolResultText(FormatRecordingList(recordings)), nil
}

func (s *Server) handleGetRecording(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	meetingID, ok := args["meeting_id"].(string)
	if !ok || meetingID == "" {
		return errorResult("meeting_id is required"), nil
	}

	detail, err := s.api.GetRecording(ctx, meetingID)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to get recording: %v", err)), nil
	}

	section, _ := args["section"].(string)
	if section != "" {
		return mcp.NewToolResultText(FormatRecordingSection(detail, section)), nil
	}

	return mcp.NewToolResultText(FormatRecordingDetail(detail)), nil
}

func (s *Server) handleSearchRecordings(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.GetArguments()["query"].(string)
	if !ok || query == "" {
		return errorResult("query is required"), nil
	}

	recordings, err := s.api.SearchRecordings(ctx, query)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to search recordings: %v", err)), nil
	}

	return mcp.NewToolResultText(FormatRecordingList(recordings)), nil
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(msg),
		},
		IsError: true,
	}
}

// Ensure handler signatures match what mcp-go expects.
var (
	_ server.ToolHandlerFunc = (*Server)(nil).handleListRecordings
	_ server.ToolHandlerFunc = (*Server)(nil).handleGetRecording
	_ server.ToolHandlerFunc = (*Server)(nil).handleSearchRecordings
)
