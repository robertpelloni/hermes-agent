package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	skinpkg "github.com/robertpelloni/hermes-agent/hermes_cli/skin"
	"github.com/robertpelloni/hermes-agent/pkg/agent"
)

// transcriptItem implements list.Item – each line of the conversation is a separate item.
type transcriptItem struct {
	content string
}

func (t transcriptItem) Title() string       { return t.content }
func (t transcriptItem) Description() string { return "" }
func (t transcriptItem) FilterValue() string { return t.content }

// Model holds the TUI state.
type Model struct {
	transcriptList   list.Model
	input            textinput.Model
	agent            *agent.Agent
	thinking         bool
	spin             spinner.Model
	quit             bool
	program          *tea.Program // reference to running program (for async messages)
	activeToolName   string

	// Slash command palette
	showPalette    bool
	paletteInput   textinput.Model
	paletteList    list.Model

	// The active UI skin
	skin           *skinpkg.Skin
}

// streamMsg wraps an agent StreamEvent into a Bubble Tea message.
type streamMsg struct {
	ev agent.StreamEvent
}

// nilMsg is returned by commands that have no immediate message to deliver
// (they schedule work via goroutines instead).
type nilMsg struct{}

// SetProgram attaches the running *tea.Program to the Model.
// This is needed so that long-running background goroutines (e.g. the agent
// stream consumer) can inject messages back into the Bubble Tea event loop.
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

// commandItem wraps CommandDef for the list.
type commandItem struct {
	cmd CommandDef
}

func (c commandItem) Title() string       { return "/" + c.cmd.Name }
func (c commandItem) Description() string { return c.cmd.Description }
func (c commandItem) FilterValue() string { return "/" + c.cmd.Name + " " + c.cmd.Description }

// NewModel builds a fresh Model wired to an Hermes agent.
func NewModel() *Model {
	ti := textinput.New()
	ti.Placeholder = "Enter a prompt…"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 80

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle()

	items := []list.Item{}
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	l := list.NewModel(items, delegate, 0, 0)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	// Palette input
	pi := textinput.New()
	pi.Placeholder = "Type command…"
	pi.Focus()
	pi.CharLimit = 50
	pi.Width = 40

	// Palette list
	plItems := []list.Item{}
	for _, cmd := range CommandRegistry {
		plItems = append(plItems, commandItem{cmd: cmd})
	}
	plDelegate := list.NewDefaultDelegate()
	plDelegate.SetHeight(1)
	pl := list.NewModel(plItems, plDelegate, 50, 10)
	pl.SetShowStatusBar(false)
	pl.SetShowHelp(false)
	pl.DisableQuitKeybindings()

	ag := agent.New(agent.DefaultConfig())

	// Load skin name from $HERMES_HOME/config.yaml (display.skin: <name>)
	skinName := skinpkg.DefaultName()
	if home := os.Getenv("HERMES_HOME"); home != "" {
		cfgPath := filepath.Join(home, "config.yaml")
		if data, err := os.ReadFile(cfgPath); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "display.skin:") {
					skinName = strings.TrimSpace(strings.TrimPrefix(line, "display.skin:"))
					break
				}
			}
		}
	}
	loadedSkin, err := skinpkg.Load(skinName)
	if err != nil {
		loadedSkin, _ = skinpkg.Load(skinpkg.DefaultName())
	}

	m := &Model{
		transcriptList: l,
		input:          ti,
		agent:          ag,
		spin:           sp,
		paletteInput:   pi,
		paletteList:    pl,
		skin:           loadedSkin,
	}
	m.applySkin(loadedSkin)
	return m}

// applySkin updates spinner style, colour and tool prefix according to the active skin.
func (m *Model) applySkin(s *skinpkg.Skin) {
	if s == nil {
		return
	}
	// Update spinner faces if the skin defines them.
	if len(s.Spinner.ThinkingFaces) > 0 {
		sp := spinner.New()
		sp.Spinner = spinner.Spinner{Frames: s.Spinner.ThinkingFaces, FPS: 30}
		if s.Spinner.Style != "" {
			sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(s.Spinner.Style))
		}
		m.spin = sp
	}
	// Store the skin for later rendering (e.g., tool prefix).
	m.skin = s
}

// Init returns the initial Bubble Tea command.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spin.Tick)
}

// addLine appends a markdown-rendered line to the transcript list.
func (m *Model) addLine(line string) {
	rendered, err := glamour.Render(line, "dark")
	if err != nil {
		rendered = line
	}
	m.transcriptList.InsertItem(len(m.transcriptList.Items()), transcriptItem{content: rendered})
}

// updateThinkingLine returns the spinner + "thinking…" line, or empty when idle.
func (m *Model) updateThinkingLine() string {
	if !m.thinking {
		return ""
	}
	label := "*Assistant: thinking…*"
	if m.activeToolName != "" {
		label = fmt.Sprintf("*Running tool:* %s", m.activeToolName)
	}
	return fmt.Sprintf("%s  %s ", m.spin.View(), label)
}

// Update handles incoming Bubble Tea messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showPalette {
			// Palette mode – handle keys for palette navigation / filtering.
			switch msg.String() {
			case "esc":
				// Hide palette and return to normal mode.
				m.showPalette = false
				return m, nil
			case "enter":
				// Execute the selected command.
				if sel, ok := m.paletteList.SelectedItem().(commandItem); ok {
					cmdDef := sel.cmd
					// Run the command's handler and capture output.
					output := cmdDef.Handler("")
					// Special handling for built‑in commands that affect UI.
					if cmdDef.Name == "quit" || cmdDef.Name == "exit" || cmdDef.Name == "q" {
						m.quit = true
						return m, tea.Quit
					}
					if cmdDef.Name == "clear" {
						// Clear transcript.
						m.transcriptList = list.NewModel([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
						m.addLine(output)
						m.showPalette = false
						return m, nil
					}
					if cmdDef.Name == "new" {
						// Reset everything – clear transcript and reset state.
						m.transcriptList = list.NewModel([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
						m.addLine(output)
						m.showPalette = false
						return m, nil
					}
					// Default: treat output as a normal assistant message.
					m.addLine(output)
					m.showPalette = false
					return m, nil
				}
				return m, nil
			case "up", "down", "ctrl+k", "ctrl+j":
				// Forward navigation keys to the list.
				var cmd tea.Cmd
				m.paletteList, cmd = m.paletteList.Update(msg)
				return m, cmd
			default:
				// Update the palette input (filter) and refresh the list.
				var cmd tea.Cmd
				m.paletteInput, cmd = m.paletteInput.Update(msg)
				// Filter commands based on the current input value.
				filtered := FilterCommands(m.paletteInput.Value())
				items := []list.Item{}
				for _, c := range filtered {
					items = append(items, commandItem{cmd: c})
				}
				m.paletteList.SetItems(items)
				return m, cmd
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quit = true
			return m, tea.Quit

		case "enter":
			if m.thinking {
				return m, nil
			}
			user := strings.TrimSpace(m.input.Value())
			if user == "" {
				return m, nil
			}
			m.addLine(fmt.Sprintf("**You:** %s", user))
			m.input.SetValue("")
			m.thinking = true
			m.activeToolName = ""
			// Start spinner ticks AND run the agent streaming call.
			return m, tea.Batch(m.spin.Tick, m.callAgent(user))

		case "/":
			// Open slash command palette.
			m.showPalette = true
			// Reset palette input.
			m.paletteInput.SetValue("")
			// Show all commands initially.
			items := []list.Item{}
			for _, c := range CommandRegistry {
				items = append(items, commandItem{cmd: c})
			}
			m.paletteList.SetItems(items)
			return m, nil

		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			m.transcriptList, _ = m.transcriptList.Update(msg)
			return m, cmd
		}

	case streamMsg:
		// (unchanged – same as previous version)
		 ev := msg.ev
		 switch ev.Type {
		 case agent.EventText:
		 	 m.addLine(fmt.Sprintf("**Assistant:** %s", ev.Text))
		 case agent.EventToolStart:
		 	 m.activeToolName = ev.ToolName
		 	 m.addLine(fmt.Sprintf("*Tool start:* `%s`", ev.ToolName))
		 case agent.EventToolDone:
		 	 if ev.ToolError != nil {
		 		 m.addLine(fmt.Sprintf("**Tool %s error:** %v", ev.ToolName, ev.ToolError))
		 	 } else {
		 		 m.addLine(fmt.Sprintf("**Tool %s result:** %s", ev.ToolName, ev.ToolResult))
		 	 }
		 	 m.activeToolName = ""
		 case agent.EventOutcome:
		 	 m.addLine(fmt.Sprintf("**Assistant (final):** %s", ev.Outcome))
		 case agent.EventDone:
		 	 m.thinking = false
		 	 m.activeToolName = ""
		 case agent.EventError:
		 	 m.thinking = false
		 	 m.activeToolName = ""
		 	 m.addLine(fmt.Sprintf("**Error:** %v", ev.Error))
		 }
		 return m, nil

	case nilMsg:
		return m, nil

	case spinner.TickMsg:
		if m.thinking {
			var cmd tea.Cmd
			m.spin, cmd = m.spin.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Fallback – let the input field process any other messages.
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// callAgent kicks off a streaming conversation in a background goroutine
// and forwards every StreamEvent into the Bubble Tea loop via program.Send.
func (m *Model) callAgent(user string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		ch := m.agent.HandleMessageStream(ctx, "tui", "user", user)
		go func() {
			for ev := range ch {
				if m.program != nil {
					m.program.Send(streamMsg{ev: ev})
				}
			}
		}()
		return nilMsg{}
	}
}

// View renders the current UI.
func (m *Model) View() string {
	if m.quit {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(m.transcriptList.View())
	sb.WriteRune('\n')
	if thinkingLine := m.updateThinkingLine(); thinkingLine != "" {
		sb.WriteString(thinkingLine + "\n")
	}
	sb.WriteString("> ")
	sb.WriteString(m.input.View())
	sb.WriteString("\n\nPress Ctrl‑C or q to quit.")
	return sb.String()
}
