package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/jlentink/twinmind-mcp/internal/api"
	"github.com/jlentink/twinmind-mcp/internal/auth"
	"github.com/jlentink/twinmind-mcp/internal/config"
	"github.com/jlentink/twinmind-mcp/internal/tui"
	"github.com/spf13/cobra"
)

var Version = "dev"

var jsonOutput bool
var outputOnExit bool

func main() {
	rootCmd := &cobra.Command{
		Use:     "twinmind-cli",
		Short:   "TwinMind CLI - manage your meeting recordings",
		Version: Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			client := api.NewClient(cfg)
			return tui.Run(client, outputOnExit)
		},
	}

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	rootCmd.Flags().BoolVar(&outputOnExit, "output-on-exit", false, "print viewed content to stdout on exit")

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}

	authCmd.AddCommand(authLoginCmd())
	authCmd.AddCommand(authStatusCmd())

	recordingsCmd := &cobra.Command{
		Use:     "recordings",
		Aliases: []string{"rec"},
		Short:   "Recording commands",
	}

	recordingsCmd.AddCommand(recordingsListCmd())
	recordingsCmd.AddCommand(recordingsGetCmd())
	recordingsCmd.AddCommand(recordingsSearchCmd())

	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(recordingsCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func authLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with TwinMind via browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := auth.Login()
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			cfg.Auth.IDToken = token.IDToken
			cfg.Auth.RefreshToken = token.RefreshToken
			cfg.Auth.IDTokenExpiry = token.ExpiresAt

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			fmt.Println("Successfully authenticated.")
			return nil
		},
	}
}

func authStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if cfg.Auth.RefreshToken == "" {
				fmt.Println("Not authenticated. Run: twinmind-cli auth login")
				return nil
			}

			token := &auth.TokenPair{
				IDToken:   cfg.Auth.IDToken,
				ExpiresAt: cfg.Auth.IDTokenExpiry,
			}

			if token.IsExpired() {
				fmt.Println("Authenticated (token expired, will auto-refresh on next request)")
			} else {
				fmt.Printf("Authenticated. Token expires: %s\n", cfg.Auth.IDTokenExpiry.Format(time.RFC3339))
			}

			fmt.Printf("Config: %s\n", config.ConfigFile())
			return nil
		},
	}
}

func recordingsListCmd() *cobra.Command {
	var limit, offset int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recordings",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client := api.NewClient(cfg)
			recordings, err := client.ListRecordings(context.Background(), limit, offset)
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(recordings)
			}

			printRecordingsTable(recordings)
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "maximum number of recordings to return")
	cmd.Flags().IntVar(&offset, "offset", 0, "number of recordings to skip")
	return cmd
}

func recordingsGetCmd() *cobra.Command {
	var only string

	cmd := &cobra.Command{
		Use:   "get <meeting_id>",
		Short: "Get recording details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client := api.NewClient(cfg)
			detail, err := client.GetRecording(context.Background(), args[0])
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(detail)
			}

			if only != "" {
				printRecordingSection(detail, only)
			} else {
				printRecordingDetail(detail)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&only, "only", "", "output only a specific section: transcript, summary, action, notes")
	return cmd
}

func recordingsSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search recordings by title",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client := api.NewClient(cfg)
			recordings, err := client.SearchRecordings(context.Background(), args[0])
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(recordings)
			}

			if len(recordings) == 0 {
				fmt.Println("No recordings found.")
				return nil
			}

			printRecordingsTable(recordings)
			return nil
		},
	}
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printRecordingsTable(recordings []api.RecordingTitle) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "MEETING ID\tTITLE\tSTART TIME\tDURATION")
	fmt.Fprintln(w, "----------\t-----\t----------\t--------")
	for _, r := range recordings {
		duration := formatDuration(r.Metadata.DurationSeconds)
		title := truncate(r.Title, 60)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			r.MeetingID,
			title,
			r.StartTime.Local().Format("2006-01-02 15:04"),
			duration,
		)
	}
	w.Flush()
}

func printRecordingDetail(d *api.RecordingDetail) {
	s := d.Summary
	fmt.Printf("Title:      %s\n", s.MeetingTitle)
	fmt.Printf("Meeting ID: %s\n", s.MeetingID)
	fmt.Printf("Start:      %s\n", s.StartTime.Local().Format("2006-01-02 15:04"))
	fmt.Printf("End:        %s\n", s.EndTime.Local().Format("2006-01-02 15:04"))
	fmt.Printf("Status:     %s\n", s.Status)

	if s.Summary != "" {
		fmt.Println("\n--- Summary ---")
		fmt.Println(s.Summary)
	}

	if s.Action != "" {
		fmt.Println("\n--- Action Items ---")
		fmt.Println(s.Action)
	}

	if d.Notes.Notes != "" && d.Notes.Notes != "notes: " {
		fmt.Println("\n--- Notes ---")
		fmt.Println(d.Notes.Notes)
	}

	if s.Transcript != "" {
		fmt.Println("\n--- Transcript ---")
		fmt.Println(s.Transcript)
	}
}

func printRecordingSection(d *api.RecordingDetail, section string) {
	switch section {
	case "transcript":
		fmt.Println(d.Summary.Transcript)
	case "summary":
		fmt.Println(d.Summary.Summary)
	case "action":
		fmt.Println(d.Summary.Action)
	case "notes":
		fmt.Println(d.Notes.Notes)
	default:
		fmt.Fprintf(os.Stderr, "unknown section %q, valid options: transcript, summary, action, notes\n", section)
	}
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
