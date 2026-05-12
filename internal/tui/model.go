package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jlentink/twinmind-mcp/internal/api"
)

type section int

const (
	sectionSummary section = iota
	sectionActions
	sectionTranscript
)

func (s section) String() string {
	switch s {
	case sectionSummary:
		return "Summary"
	case sectionActions:
		return "Actions"
	case sectionTranscript:
		return "Transcript"
	}
	return ""
}

type panel int

const (
	panelList panel = iota
	panelDetail
)

type appState int

const (
	stateLoading appState = iota
	stateReady
	stateError
)

type Model struct {
	client *api.Client

	list     list.Model
	viewport viewport.Model

	detail        map[string]*api.RecordingDetail
	activeID      string
	activeSection section
	loadingDetail bool
	focusedPanel  panel
	appState      appState
	errMsg        string

	width  int
	height int

	exitContent  string
	flashMessage string
}

func New(client *api.Client) Model {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.Title = "Recordings"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	vp := viewport.New(0, 0)
	vp.SetContent("Loading recordings...")

	return Model{
		client:   client,
		list:     l,
		viewport: vp,
		detail:   make(map[string]*api.RecordingDetail),
		appState: stateLoading,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchRecordingsCmd(m.client)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizePanels()
		return m, nil

	case recordingsLoadedMsg:
		if msg.err != nil {
			m.appState = stateError
			m.errMsg = msg.err.Error()
			return m, nil
		}
		items := make([]list.Item, len(msg.recordings))
		for i, r := range msg.recordings {
			items[i] = recordingItem{recording: r}
		}
		m.list.SetItems(items)
		m.appState = stateReady
		if len(items) > 0 {
			m.viewport.SetContent("Press Enter to view recording details.")
		} else {
			m.viewport.SetContent("No recordings found.")
		}
		return m, nil

	case recordingDetailLoadedMsg:
		if msg.err != nil {
			if msg.meetingID == m.activeID {
				m.loadingDetail = false
				m.viewport.SetContent("Error loading detail: " + msg.err.Error())
			}
			return m, nil
		}
		m.detail[msg.meetingID] = msg.detail
		if msg.meetingID == m.activeID {
			m.loadingDetail = false
			m.refreshViewport()
		}
		return m, nil

	case clearFlashMsg:
		m.flashMessage = ""
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys
	switch key {
	case "q", "x":
		m.exitContent = m.currentViewportContent()
		return m, tea.Quit

	case "tab":
		if m.focusedPanel == panelList {
			m.focusedPanel = panelDetail
		} else {
			m.focusedPanel = panelList
		}
		return m, nil

	case "r":
		m.appState = stateLoading
		m.viewport.SetContent("Refreshing...")
		return m, fetchRecordingsCmd(m.client)

	case "s":
		if m.activeID != "" && m.detail[m.activeID] != nil {
			m.activeSection = sectionSummary
			m.refreshViewport()
		}
		return m, nil

	case "a":
		if m.activeID != "" && m.detail[m.activeID] != nil {
			m.activeSection = sectionActions
			m.refreshViewport()
		}
		return m, nil

	case "t":
		if m.activeID != "" && m.detail[m.activeID] != nil {
			m.activeSection = sectionTranscript
			m.refreshViewport()
		}
		return m, nil

	case "c":
		content := m.currentViewportContent()
		if content != "" {
			if err := clipboard.WriteAll(content); err != nil {
				m.flashMessage = "Clipboard error: " + err.Error()
			} else {
				m.flashMessage = "Copied to clipboard!"
			}
			return m, clearFlashAfter(2 * time.Second)
		}
		return m, nil

	case "enter":
		if m.focusedPanel == panelList {
			if item, ok := m.list.SelectedItem().(recordingItem); ok {
				m.activeID = item.recording.MeetingID
				if _, cached := m.detail[m.activeID]; cached {
					m.refreshViewport()
					return m, nil
				}
				m.loadingDetail = true
				m.viewport.SetContent("Loading...")
				return m, fetchDetailCmd(m.client, m.activeID)
			}
		}
		return m, nil
	}

	// Route navigation keys based on focused panel
	if m.focusedPanel == panelDetail {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	// List panel is focused
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	if m.appState == stateError {
		return fmt.Sprintf("Error: %s\n\nPress r to retry, q to quit.\n", m.errMsg)
	}

	leftWidth := m.width / 3
	if leftWidth < 30 {
		leftWidth = 30
	}
	rightWidth := m.width - leftWidth

	contentHeight := m.height - 2 // status bar

	// Left panel
	m.list.SetSize(leftWidth-2, contentHeight-2)
	leftBorder := panelBorder
	if m.focusedPanel == panelList {
		leftBorder = activePanelBorder
	}
	leftPanel := leftBorder.
		Width(leftWidth - 2).
		Height(contentHeight - 2).
		Render(m.list.View())

	// Right panel
	tabs := m.renderTabs()
	tabHeight := lipgloss.Height(tabs)

	vpHeight := contentHeight - tabHeight - 4
	if vpHeight < 1 {
		vpHeight = 1
	}
	m.viewport.Width = rightWidth - 4
	m.viewport.Height = vpHeight

	detailTitle := ""
	if m.activeID != "" && m.detail[m.activeID] != nil {
		detailTitle = titleStyle.Render(m.detail[m.activeID].Summary.MeetingTitle)
	}

	rightContent := lipgloss.JoinVertical(lipgloss.Left, detailTitle, tabs, m.viewport.View())
	rightBorder := panelBorder
	if m.focusedPanel == panelDetail {
		rightBorder = activePanelBorder
	}
	rightPanel := rightBorder.
		Width(rightWidth - 4).
		Height(contentHeight - 2).
		Render(rightContent)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	statusText := "tab:switch panel  enter:select  s:summary  a:actions  t:transcript  c:copy  r:refresh  q:quit"
	if m.flashMessage != "" {
		statusText = m.flashMessage
	}
	status := statusBarStyle.Width(m.width).Render(statusText)

	return lipgloss.JoinVertical(lipgloss.Left, body, status)
}

func (m *Model) renderTabs() string {
	if m.activeID == "" {
		return ""
	}

	sections := []section{sectionSummary, sectionActions, sectionTranscript}
	keys := []string{"s", "a", "t"}
	var tabs []string

	for i, s := range sections {
		label := fmt.Sprintf("[%s] %s", keys[i], s.String())
		if s == m.activeSection {
			tabs = append(tabs, tabActive.Render(label))
		} else {
			tabs = append(tabs, tabInactive.Render(label))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m *Model) refreshViewport() {
	if m.activeID == "" {
		m.viewport.SetContent("Select a recording to view details.")
		return
	}
	detail, ok := m.detail[m.activeID]
	if !ok {
		m.viewport.SetContent("Loading...")
		return
	}

	var content string
	switch m.activeSection {
	case sectionSummary:
		content = detail.Summary.Summary
	case sectionActions:
		content = detail.Summary.Action
	case sectionTranscript:
		content = detail.Summary.Transcript
	}

	if strings.TrimSpace(content) == "" {
		content = "(No content available)"
	}

	m.viewport.SetContent(content)
	m.viewport.GotoTop()
}

func (m Model) currentViewportContent() string {
	if m.activeID == "" {
		return ""
	}
	detail, ok := m.detail[m.activeID]
	if !ok {
		return ""
	}
	switch m.activeSection {
	case sectionSummary:
		return detail.Summary.Summary
	case sectionActions:
		return detail.Summary.Action
	case sectionTranscript:
		return detail.Summary.Transcript
	}
	return ""
}

func (m *Model) resizePanels() {
	leftWidth := m.width / 3
	if leftWidth < 30 {
		leftWidth = 30
	}
	rightWidth := m.width - leftWidth
	contentHeight := m.height - 2

	m.list.SetSize(leftWidth-2, contentHeight-2)
	m.viewport.Width = rightWidth - 4
	m.viewport.Height = contentHeight - 6
}

func Run(client *api.Client, outputOnExit bool) error {
	m := New(client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	if outputOnExit {
		if fm, ok := finalModel.(Model); ok && fm.exitContent != "" {
			fmt.Println(fm.exitContent)
		}
	}
	return nil
}
