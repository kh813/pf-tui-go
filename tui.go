package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
)

// Styles
var (
	appStyle          = lipgloss.NewStyle().Padding(1, 2)
	titleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1)
	statusStyle       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"})
	selectedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	focusedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Underline(true)
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
)

// Views
type view int

const (
	mainView view = iota
	ruleListView
	ruleFormView
	portForwardingListView
	portForwardingFormView
	infoView
	saveConfigView
	importConfigView
	confirmationView
)

// Model
type model struct {
	list                 list.Model
	ruleList             list.Model
	portForwardingList   list.Model
	fileList             list.Model
	viewport             viewport.Model
	textinput            textinput.Model
	confirmationMessage  string
	confirming           bool
	firewallManager      *FirewallManager
	statusMessage        string
	pfStatus             string
	startupStatus        string
	currentView          view
	previousView         view
	form                 ruleForm
	portForwardingForm   portForwardingForm
	infoContent          string
	infoViewTitle        string // New field for dynamic title
	showConfirm          bool
	help                 help.Model
	keys                 keyMap
	width, height        int
}

// Messages
type pfStatusMsg string
type pfStartupStatusMsg string
type pfInfoMsg string
type currentRulesMsg string
type firewallRuleSavedMsg string
type portForwardingRuleSavedMsg string
type configLoadedMsg string
type configSavedAndBackToMainMsg string
type configExportedMsg string
type fileListMsg []list.Item
type errMsg struct{ err error }
type infoRefreshMsg struct{}

func (e errMsg) Error() string { return e.err.Error() }

// keyMap defines a set of keybindings.
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Help   key.Binding
	Quit   key.Binding
	Select key.Binding
	Back   key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("→/l", "move right"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right}, // first column
		{k.Help, k.Quit},                // second column
	}
}

// Helper function to render options
func renderOptions(label string, options []string, selected string, isFocused bool) string {
	var parts []string
	for _, opt := range options {
		if opt == selected {
			style := selectedStyle
			if isFocused {
				style = focusedStyle
			}
			parts = append(parts, style.Render(fmt.Sprintf(" %s ", opt)))
		} else {
			parts = append(parts, fmt.Sprintf(" %s ", opt))
		}
	}
	labelPart := fmt.Sprintf("    %-15s:", label)
	if isFocused {
		labelPart = focusedStyle.Render(labelPart)
	}
	return fmt.Sprintf("%s %s\n", labelPart, strings.Join(parts, " "))
}

// Helper function to render text input
func renderInput(label string, input textinput.Model, isFocused bool, activeTextInputIndex int, currentFieldIndex int, fieldLabel string) string {
	if isFocused && activeTextInputIndex == currentFieldIndex {
		input.Focus()
	} else {
		input.Blur()
	}
	labelPart := fmt.Sprintf("    %-15s:", label)
	if isFocused {
		labelPart = focusedStyle.Render(labelPart)
	}

	hint := ""
	if isFocused && activeTextInputIndex == -1 { // Only show hint if focused and not actively editing
		if (fieldLabel == "Interface" || fieldLabel == "Source" || fieldLabel == "Destination" || fieldLabel == "Port") && input.Value() == "any" {
			hint = "  <-- Press Enter to specify"
		} else if fieldLabel == "Description" && input.Value() == "" {
			hint = "  <-- Press Enter to specify"
		} else if fieldLabel == "External Port" && input.Value() == "" {
			hint = "  <-- Press Enter to specify"
		} else if fieldLabel == "Internal Port" && input.Value() == "" {
			hint = "  <-- Press Enter to specify"
		} else if fieldLabel == "External IP" && input.Value() == "any" {
			hint = "  <-- Press Enter to specify"
		} else if fieldLabel == "Internal IP" && input.Value() == "127.0.0.1" {
			hint = "  <-- Press Enter to specify"
		}
	}

	return fmt.Sprintf("%s  %s%s\n", labelPart, input.View(), hint)
}

// Commands

func checkPfStatus() tea.Msg {
	status, err := GetPfStatus()
	if err != nil {
		return errMsg{err}
	}
	return pfStatusMsg(status)
}

func checkPfStartupStatus() tea.Msg {
	status, err := CheckPfStartupStatus()
	if err != nil {
		return errMsg{err}
	}
	return pfStartupStatusMsg(status)
}

func getPfInfo() tea.Msg {
	info, err := GetPfInfo()
	if err != nil {
		return errMsg{err}
	}
	return pfInfoMsg(info)
}

func getCurrentRules() tea.Msg {
	rules, err := GetCurrentRules()
	if err != nil {
		return errMsg{err}
	}
	return currentRulesMsg(rules)
}

func enablePf() tea.Msg {
	_, err := EnablePf()
	if err != nil {
		return errMsg{err}
	}
	return checkPfStatus()
}

func disablePf() tea.Msg {
	_, err := DisablePf()
	if err != nil {
		return errMsg{err}
	}
	return checkPfStatus()
}

func enablePfOnStartup() tea.Msg {
	_, err := EnablePfOnStartup()
	if err != nil {
		return errMsg{err}
	}
	return checkPfStartupStatus()
}

func disablePfOnStartup() tea.Msg {
	_, err := DisablePfOnStartup()
	if err != nil {
		return errMsg{err}
	}
	return checkPfStartupStatus()
}

func saveConfigAs(fm *FirewallManager, path string) tea.Cmd {
	return func() tea.Msg {
		if err := fm.SaveConfigAs(path); err != nil {
			return errMsg{err}
		}
		return configExportedMsg(fmt.Sprintf("Configuration exported to %s", path))
	}
}

func importConfig(fm *FirewallManager, path string) tea.Cmd {
	return func() tea.Msg {
		LogInfo(fmt.Sprintf("Importing config from: %s", path))
		if err := fm.ImportConfigFile(path); err != nil {
			LogError(fmt.Sprintf("Error loading config: %v", err))
			return errMsg{err}
		}
		LogInfo("Config imported successfully")
		return configLoadedMsg("Configuration imported successfully.")
	}
}

func saveAndApplyRules(fm *FirewallManager) tea.Cmd {
	return func() tea.Msg {
		// Ensure pf.conf is set up correctly
		if err := SetupPfConf(); err != nil {
			return errMsg{err}
		}

		// Save the configuration
		if err := fm.SaveConfig(); err != nil {
			return errMsg{err}
		}

		// Apply the rules
		pfConf := fm.GeneratePfConf()
		output, err := ApplyRules(pfConf)
		if err != nil {
			return errMsg{fmt.Errorf("failed to apply rules: %w, output: %s", err, output)}
		}

		return configSavedAndBackToMainMsg("Configuration saved and applied to the system.")
	}
}

// item represents a list item.
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// ruleForm represents the form for adding/editing a rule.

type ruleForm struct {
	focused          int
	activeTextInput  int // -1 if no text input is active, otherwise the index of the active text input
	isNew            bool
	ruleIndex        int
	action           string
	direction        string
	quick            string
	protocol         string
	keepState        string
	interfaceInput   textinput.Model
	sourceInput      textinput.Model
	destinationInput textinput.Model
	portInput        textinput.Model
	descriptionInput textinput.Model
}

func newRuleForm() ruleForm {
	interfaceInput := textinput.New()
	interfaceInput.SetValue("any")
	interfaceInput.Prompt = ""
	interfaceInput.Blur()
	sourceInput := textinput.New()
	sourceInput.SetValue("any")
	sourceInput.Prompt = ""
	sourceInput.Blur()
	destinationInput := textinput.New()
	destinationInput.SetValue("any")
	destinationInput.Prompt = ""
	destinationInput.Blur()
	portInput := textinput.New()
	portInput.SetValue("any")
	portInput.Prompt = ""
	portInput.Blur()
	descriptionInput := textinput.New()
	descriptionInput.Prompt = ""
	descriptionInput.Blur()

	return ruleForm{
		focused:          0,
		activeTextInput:  -1,
		action:           "block",
		direction:        "in",
		quick:            "No",
		protocol:         "any",
		keepState:        "No",
		interfaceInput:   interfaceInput,
		sourceInput:      sourceInput,
		destinationInput: destinationInput,
		portInput:        portInput,
		descriptionInput: descriptionInput,
	}
}

func newPortForwardingForm() portForwardingForm {
	interfaceInput := textinput.New()
	interfaceInput.SetValue("any")
	interfaceInput.Prompt = ""
	interfaceInput.Blur()
	externalIPInput := textinput.New()
	externalIPInput.SetValue("any")
	externalIPInput.Prompt = ""
	externalIPInput.Blur()
	externalPortInput := textinput.New()
	externalPortInput.Prompt = ""
	externalPortInput.Blur()
	internalIPInput := textinput.New()
	internalIPInput.SetValue("127.0.0.1")
	internalIPInput.Prompt = ""
	internalIPInput.Blur()
	internalPortInput := textinput.New()
	internalPortInput.Prompt = ""
	internalPortInput.Blur()
	descriptionInput := textinput.New()
	descriptionInput.Prompt = ""
	descriptionInput.Blur()

	return portForwardingForm{
		focused:           0,
		activeTextInput:   -1,
		protocol:          "tcp",
		interfaceInput:    interfaceInput,
		externalIPInput:   externalIPInput,
		externalPortInput: externalPortInput,
		internalIPInput:   internalIPInput,
		internalPortInput: internalPortInput,
		descriptionInput:  descriptionInput,
	}
}

func NewModel(fm *FirewallManager) *model {
	m := model{
		firewallManager:    fm,
		pfStatus:           "Checking...",
		startupStatus:      "Unknown",
		currentView:        mainView,
		form:               newRuleForm(),
		portForwardingForm: newPortForwardingForm(),
		viewport:           viewport.New(80, 24),
		textinput:          textinput.New(),
		help:               help.New(),
		keys:               DefaultKeyMap(),
	}

	// Main menu list
	items := []list.Item{
		//item{title: ""},
		item{title: "Edit Firewall Rule"},
		item{title: "Add New Firewall Rule"},
		item{title: "Edit Port Forwarding Rule"},
		item{title: "Add Port Forwarding Rule"},
		item{title: "---"},
		item{title: "Save & Apply Configuration"},
		item{title: "Export Configuration"},
		item{title: "Import Configuration"},
		item{title: "---"},
		item{title: "Show Current Rules"},
		item{title: "Show Info"},
		item{title: "---"},
		item{title: "Enable PF"},
		item{title: "Disable PF"},
		item{title: "Enable PF on Startup"},
		item{title: "Disable PF on Startup"},
		item{title: "---"},
		item{title: "Exit"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Padding(0, 0, 0, 2)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Padding(0, 0, 0, 1)
	l := list.New(items, delegate, 0, 0)
	l.Title = "Main menu"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.SetShowHelp(false)
	l.SetShowPagination(true)
	l.SetShowStatusBar(false)
	m.list = l

	// Rule list
	ruleListDelegate := list.NewDefaultDelegate()
	ruleListDelegate.ShowDescription = false
	ruleListDelegate.SetHeight(1)
	ruleListDelegate.Styles.NormalTitle = lipgloss.NewStyle().Padding(0, 0, 0, 2)
	ruleListDelegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		Padding(0, 0, 0, 1)
	ruleListDelegate.SetSpacing(0)

	m.ruleList = list.New([]list.Item{}, ruleListDelegate, 0, 0)
	m.ruleList.Title = "Firewall Rules"
	m.ruleList.SetShowStatusBar(false)
	m.ruleList.SetFilteringEnabled(false)
	m.ruleList.SetShowHelp(false)
	m.ruleList.SetShowTitle(false)

	// Port forwarding list
	portForwardingListDelegate := list.NewDefaultDelegate()
	portForwardingListDelegate.ShowDescription = false
	portForwardingListDelegate.SetHeight(1)
	portForwardingListDelegate.SetSpacing(0)
	m.portForwardingList = list.New([]list.Item{}, portForwardingListDelegate, 0, 0)
	m.portForwardingList.Title = "Port Forwarding Rules"
	m.portForwardingList.SetShowStatusBar(false)
	m.portForwardingList.SetFilteringEnabled(false)
	m.portForwardingList.SetShowTitle(false)
	m.portForwardingList.SetShowHelp(false)

	// File list
	fileListDelegate := list.NewDefaultDelegate()
	fileListDelegate.ShowDescription = true
	fileListDelegate.SetHeight(2)
	fileListDelegate.SetSpacing(0)
	fileListDelegate.Styles.NormalTitle = lipgloss.NewStyle().Padding(0, 0, 0, 2)
	fileListDelegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		Padding(0, 0, 0, 1)

	m.fileList = list.New([]list.Item{}, fileListDelegate, 0, 0)
	m.fileList.Title = "Select a file to import"
	m.fileList.SetShowStatusBar(false)
	m.fileList.SetFilteringEnabled(false)
	m.fileList.SetShowTitle(true)
	m.fileList.SetShowHelp(false)

	return &m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		checkPfStatus,
		checkPfStartupStatus,
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.currentView == mainView {
				m.previousView = m.currentView
				m.currentView = confirmationView
				m.confirming = true
				m.confirmationMessage = "Are you sure you want to exit?"
				return m, nil
			} else if m.currentView != confirmationView {
				m.currentView = mainView
				return m, nil
			}
		}
		if m.currentView == confirmationView {
			switch msg.String() {
			case "y":
				if m.confirming {
					m.confirming = false
					if m.previousView == mainView {
						return m, tea.Quit
					} else if m.previousView == ruleFormView {
						m.currentView = mainView
						return m, nil
					} else if m.previousView == saveConfigView {
						path := m.textinput.Value()
						return m, saveConfigAs(m.firewallManager, path)
					}
				}
			case "n":
				if m.confirming {
					m.confirming = false
					m.currentView = m.previousView
				}
			}
		}

		switch m.currentView {
		case mainView:
			switch msg.String() {
			case "up", "k":
				if m.list.Index() == 0 {
					m.list.Select(len(m.list.Items()) - 1)
				} else {
					m.list.Select(m.list.Index() - 1)
				}
				return m, nil
			case "down", "j":
				if m.list.Index() == len(m.list.Items()) - 1 {
					m.list.Select(0)
				} else {
					m.list.Select(m.list.Index() + 1)
				}
				return m, nil
					
			case "enter":
				selectedItem, ok := m.list.SelectedItem().(item)
				if !ok {
					return m, nil
				}
				switch selectedItem.title {
				case " ", "---":
					// Do nothing for separators and empty space

				case "Add New Firewall Rule":
					m.currentView = ruleFormView
					m.form = newRuleForm()
					m.form.isNew = true
					m.focusRuleForm()
				case "Edit Firewall Rule":
					m.currentView = ruleListView
					return m, m.updateRuleList()

				case "Add Port Forwarding Rule":
					m.currentView = portForwardingFormView
					m.portForwardingForm = newPortForwardingForm()
					m.portForwardingForm.isNew = true
					m.focusPortForwardingForm()
				case "Edit Port Forwarding Rule":
					m.currentView = portForwardingListView
					m.updatePortForwardingList()
				case "Show Info":
					m.currentView = infoView
					m.infoViewTitle = "Live PF Info"
					m.viewport.SetContent("Loading...")
					return m, tea.Batch(getPfInfo, func() tea.Msg { return infoRefreshMsg{} })
				case "Show Current Rules":
					m.currentView = infoView
					m.infoViewTitle = "Current Live PF Rules"
					m.viewport.SetContent("Loading...")
					return m, getCurrentRules
				case "Enable PF":
					return m, enablePf
				case "Disable PF":
					return m, disablePf
				case "Enable PF on Startup":
					return m, enablePfOnStartup
				case "Disable PF on Startup":
					return m, disablePfOnStartup
				case "Save & Apply Configuration":
					return m, saveAndApplyRules(m.firewallManager)
				case "Export Configuration":
					m.currentView = saveConfigView
					configPath, _ := GetConfigPath()
					timestamp := time.Now().Format("20060102-150405")
					filename := fmt.Sprintf("rules-export-%s.json", timestamp)
					m.textinput.SetValue(filepath.Join(configPath, filename))
					m.textinput.Focus()
				case "Import Configuration":
					m.currentView = importConfigView
					return m, m.updateFileList()
				case "Exit":
					m.previousView = m.currentView
					m.currentView = confirmationView
					m.confirming = true
					m.confirmationMessage = "Are you sure you want to exit?"

					return m, nil
				}
			}
				case ruleListView:
			// Handle key presses for reordering
			switch msg.String() {
			case "k":
				idx := m.ruleList.Index()
				if idx > 0 {
					m.firewallManager.MoveFirewallRule(idx, idx-1)
					m.ruleList.SetItems(m.getRuleListItems())
					m.ruleList.Select(idx - 1) // Select the moved item
				}
				return m, nil
			case "j":
				idx := m.ruleList.Index()
				if idx < len(m.firewallManager.Config.FirewallRules)-1 {
					m.firewallManager.MoveFirewallRule(idx, idx+1)
					m.ruleList.SetItems(m.getRuleListItems())
					m.ruleList.Select(idx + 1) // Select the moved item
				}
				return m, nil
			}

			// Let the list model handle its own updates for other keys
			m.ruleList, cmd = m.ruleList.Update(msg)

			// Handle other specific key presses for this view
			switch msg.String() {
			case "esc":
				m.currentView = mainView
			case "a": // Add new rule
				m.currentView = ruleFormView
				m.form = newRuleForm()
				m.form.isNew = true
				m.focusRuleForm()
			case "enter":
				selectedItem, ok := m.ruleList.SelectedItem().(ruleListItem)
				if ok {
					m.currentView = ruleFormView
					m.form = newRuleForm()
					m.form.isNew = false
					m.form.ruleIndex = selectedItem.index
					rule := m.firewallManager.Config.FirewallRules[selectedItem.index]
					m.form.action = rule.Action
					m.form.direction = rule.Direction
					m.form.quick = map[bool]string{true: "Yes", false: "No"}[rule.Quick]
					m.form.interfaceInput.SetValue(rule.Interface)
					m.form.protocol = rule.Protocol
					m.form.sourceInput.SetValue(rule.Source)
					m.form.destinationInput.SetValue(rule.Destination)
					m.form.portInput.SetValue(rule.Port)
					m.form.keepState = map[bool]string{true: "Yes", false: "No"}[rule.KeepState]
					m.form.descriptionInput.SetValue(rule.Description)
					m.focusRuleForm()
				}
			case "d":
				selectedItem, ok := m.ruleList.SelectedItem().(ruleListItem)
				if ok {
					cmd = func() tea.Msg {
						if err := m.firewallManager.DeleteFirewallRule(selectedItem.index); err != nil {
							return errMsg{err}
						}
						return firewallRuleSavedMsg("Rule deleted successfully.")
					}
					return m, tea.Sequence(cmd, m.updateRuleList())
				}
			case "s":
				return m, func() tea.Msg {
					if err := m.firewallManager.SaveConfig(); err != nil {
						return errMsg{err}
					}
					return configSavedAndBackToMainMsg("Rule order saved.")
				}
			}
				case ruleFormView:
			// If a text input is active, let it handle the key presses
			if m.form.activeTextInput != -1 {
				var cmd tea.Cmd
				switch m.form.activeTextInput {
				case 3:
					m.form.interfaceInput, cmd = m.form.interfaceInput.Update(msg)
				case 5:
					m.form.sourceInput, cmd = m.form.sourceInput.Update(msg)
				case 6:
					m.form.destinationInput, cmd = m.form.destinationInput.Update(msg)
				case 7:
					m.form.portInput, cmd = m.form.portInput.Update(msg)
				case 9:
					m.form.descriptionInput, cmd = m.form.descriptionInput.Update(msg)
				}

				if msg.String() == "enter" {
					// Finalize input and unfocus
					m.form.activeTextInput = -1
					m.focusRuleForm() // Blur all text inputs
					return m, nil
				}
				return m, cmd
			}

			// Handle navigation and option changes when no text input is active
			switch msg.String() {
			case "esc":
				m.currentView = ruleListView
			case "s":
				// Only save if no text input is active
				if m.form.activeTextInput == -1 {
					return m, m.saveRule()
				}
			case "enter":
				// If the current field is a text input, enter editing mode
				if m.form.focused == 3 || m.form.focused == 5 || m.form.focused == 6 || m.form.focused == 7 || m.form.focused == 9 {
					m.form.activeTextInput = m.form.focused
					m.focusRuleForm() // Focus the active text input
					return m, nil
				}
			case "up":
				m.form.focused = (m.form.focused - 1 + 10) % 10
				m.focusRuleForm()
			case "down":
				m.form.focused = (m.form.focused + 1) % 10
				m.focusRuleForm()
			case "left":
				switch m.form.focused {
				case 0: // Action
					if m.form.action == "pass" {
						m.form.action = "block"
					} else {
						m.form.action = "pass"
					}
				case 1: // Direction
					if m.form.direction == "out" {
						m.form.direction = "in"
					} else {
						m.form.direction = "out"
					}
				case 2: // Quick
					if m.form.quick == "No" {
						m.form.quick = "Yes"
					} else {
						m.form.quick = "No"
					}
				case 4: // Protocol
					options := []string{"tcp", "udp", "tcp,udp", "icmp", "any"}
					for i, opt := range options {
						if opt == m.form.protocol {
							m.form.protocol = options[(i-1+len(options))%len(options)]
							break
						}
					}
				case 8: // Keep State
					if m.form.keepState == "No" {
						m.form.keepState = "Yes"
					} else {
						m.form.keepState = "No"
					}
				}
			case "right":
				switch m.form.focused {
				case 0: // Action
					if m.form.action == "block" {
						m.form.action = "pass"
					} else {
						m.form.action = "block"
					}
				case 1: // Direction
					if m.form.direction == "in" {
						m.form.direction = "out"
					} else {
						m.form.direction = "in"
					}
				case 2: // Quick
					if m.form.quick == "Yes" {
						m.form.quick = "No"
					} else {
						m.form.quick = "Yes"
					}
				case 4: // Protocol
					options := []string{"tcp", "udp", "tcp,udp", "icmp", "any"}
					for i, opt := range options {
						if opt == m.form.protocol {
							m.form.protocol = options[(i+1)%len(options)]
							break
						}
					}
				case 8: // Keep State
					if m.form.keepState == "Yes" {
						m.form.keepState = "No"
					} else {
						m.form.keepState = "Yes"
					}
				}
			}
			return m, nil
		case portForwardingListView:
			m.portForwardingList, cmd = m.portForwardingList.Update(msg)
			switch msg.String() {
			case "esc":
				m.currentView = mainView
			case "a": // Add new port forwarding rule
				m.currentView = portForwardingFormView
				m.portForwardingForm = newPortForwardingForm()
				m.portForwardingForm.isNew = true
				m.focusPortForwardingForm()
			case "enter":
				selectedItem, ok := m.portForwardingList.SelectedItem().(portForwardingListItem)
				if ok {
					m.currentView = portForwardingFormView
					m.portForwardingForm = newPortForwardingForm()
					m.portForwardingForm.isNew = false
					m.portForwardingForm.ruleIndex = selectedItem.index
					rule := m.firewallManager.Config.PortForwardingRules[selectedItem.index]
					m.portForwardingForm.interfaceInput.SetValue(rule.Interface)
					m.portForwardingForm.protocol = rule.Protocol
					m.portForwardingForm.externalIPInput.SetValue(rule.ExternalIP)
					m.portForwardingForm.externalPortInput.SetValue(rule.ExternalPort)
					m.portForwardingForm.internalIPInput.SetValue(rule.InternalIP)
					m.portForwardingForm.internalPortInput.SetValue(rule.InternalPort)
					m.portForwardingForm.descriptionInput.SetValue(rule.Description)
					m.focusPortForwardingForm()
				}
			case "d":
				selectedItem, ok := m.portForwardingList.SelectedItem().(portForwardingListItem)
				if ok {
					cmd = func() tea.Msg {
						if err := m.firewallManager.DeletePortForwardingRule(selectedItem.index); err != nil {
							return errMsg{err}
						}
						return firewallRuleSavedMsg("Port forwarding rule deleted successfully.")
					}
					return m, tea.Sequence(cmd, func() tea.Msg {
						m.updatePortForwardingList()
						return nil
					})
				}
			case "k":
				selectedItem, ok := m.portForwardingList.SelectedItem().(portForwardingListItem)
				if ok {
					m.firewallManager.MovePortForwardingRule(selectedItem.index, selectedItem.index-1)
					m.updatePortForwardingList()
				}
			case "j":
				selectedItem, ok := m.portForwardingList.SelectedItem().(portForwardingListItem)
				if ok {
					m.firewallManager.MovePortForwardingRule(selectedItem.index, selectedItem.index+1)
					m.updatePortForwardingList()
				}
			case "s":
				return m, func() tea.Msg {
					if err := m.firewallManager.SaveConfig(); err != nil {
						return errMsg{err}
					}
					return configSavedAndBackToMainMsg("Rule order saved.")
				}
			}
		case portForwardingFormView:
			// If a text input is active, let it handle the key presses
			if m.portForwardingForm.activeTextInput != -1 {
				var cmd tea.Cmd
				switch m.portForwardingForm.activeTextInput {
				case 0:
					m.portForwardingForm.interfaceInput, cmd = m.portForwardingForm.interfaceInput.Update(msg)
				case 2:
					m.portForwardingForm.externalIPInput, cmd = m.portForwardingForm.externalIPInput.Update(msg)
				case 3:
					m.portForwardingForm.externalPortInput, cmd = m.portForwardingForm.externalPortInput.Update(msg)
				case 4:
					m.portForwardingForm.internalIPInput, cmd = m.portForwardingForm.internalIPInput.Update(msg)
				case 5:
					m.portForwardingForm.internalPortInput, cmd = m.portForwardingForm.internalPortInput.Update(msg)
				case 6:
					m.portForwardingForm.descriptionInput, cmd = m.portForwardingForm.descriptionInput.Update(msg)
				}

				if msg.String() == "enter" {
					// Finalize input and unfocus
					m.portForwardingForm.activeTextInput = -1
					m.focusPortForwardingForm() // Blur all text inputs
					return m, nil
				}
				return m, cmd
			}

			switch msg.String() {
			case "esc":
				m.currentView = mainView
			case "s":
				// Only save if no text input is active
				if m.portForwardingForm.activeTextInput == -1 {
					return m, m.savePortForwardingRule()
				}
			case "enter":
				// If the current field is a text input, enter editing mode
				if m.portForwardingForm.focused == 0 || m.portForwardingForm.focused == 2 || m.portForwardingForm.focused == 3 || m.portForwardingForm.focused == 4 || m.portForwardingForm.focused == 5 || m.portForwardingForm.focused == 6 {
					m.portForwardingForm.activeTextInput = m.portForwardingForm.focused
					m.focusPortForwardingForm() // Focus the active text input
					return m, nil
				}
				// Otherwise, move to the next field (for option fields)
				m.portForwardingForm.focused = (m.portForwardingForm.focused + 1) % 7
				m.focusPortForwardingForm()
			case "up":
				m.portForwardingForm.focused = (m.portForwardingForm.focused - 1 + 7) % 7
				m.focusPortForwardingForm()
			case "down":
				m.portForwardingForm.focused = (m.portForwardingForm.focused + 1) % 7
				m.focusPortForwardingForm()
			case "left", "right":
				if m.portForwardingForm.focused == 1 { // Protocol
					if m.portForwardingForm.protocol == "tcp" {
						m.portForwardingForm.protocol = "udp"
					} else {
						m.portForwardingForm.protocol = "tcp"
					}
				}
			}
			return m, nil
		case infoView:
			m.viewport, cmd = m.viewport.Update(msg)
			switch msg.String() {
			case "esc", "q":
				m.currentView = mainView
				return m, nil
			}
		case saveConfigView:
			m.textinput, cmd = m.textinput.Update(msg)
			switch msg.String() {
			case "esc":
				m.currentView = mainView
			case "enter":
				path := m.textinput.Value()
				if path != "" {
					// Check if file exists
					if _, err := os.Stat(path); err == nil {
						m.previousView = saveConfigView
						m.currentView = confirmationView
						m.confirming = true
						m.confirmationMessage = fmt.Sprintf("File '%s' already exists. Overwrite?", path)
						return m, nil
					}
					return m, saveConfigAs(m.firewallManager, path)
				}
			}
		case importConfigView:
			m.fileList, cmd = m.fileList.Update(msg)
			switch msg.String() {
			case "enter":
				selectedItem, ok := m.fileList.SelectedItem().(fileInfo)
				if ok {
					configPath, _ := GetConfigPath()
					path := filepath.Join(configPath, selectedItem.name)
					return m, importConfig(m.firewallManager, path)
				}
			case "esc":
				m.currentView = mainView
			}
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-4)
		m.ruleList.SetSize(msg.Width-h, msg.Height-v-4)
		m.portForwardingList.SetSize(msg.Width-h, msg.Height-v-4)
		m.fileList.SetSize(msg.Width-h, msg.Height-v-4)
		m.viewport.Width = msg.Width - h
		m.viewport.Height = msg.Height - v - 4
		m.help.Width = msg.Width

	case pfStatusMsg:
		m.pfStatus = string(msg)
		return m, nil

	case pfStartupStatusMsg:
		m.startupStatus = string(msg)
		return m, nil

	case pfInfoMsg:
		m.infoContent = string(msg)
		m.viewport.SetContent(m.infoContent)
		return m, nil

	case infoRefreshMsg:
		if m.currentView == infoView && m.pfStatus == "Enabled" {
			return m, tea.Batch(
				getPfInfo,
				tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return infoRefreshMsg{}
				}),
			)
		}
		return m, nil

	case currentRulesMsg:
		m.infoContent = string(msg)
		m.viewport.SetContent(m.infoContent)
		return m, nil

	case firewallRuleSavedMsg:
		m.statusMessage = string(msg)
		m.currentView = ruleListView
		return m, m.updateRuleList()

	case portForwardingRuleSavedMsg:
		m.statusMessage = string(msg)
		m.currentView = portForwardingListView
		m.updatePortForwardingList()
		return m, nil

		case configLoadedMsg:
		m.statusMessage = string(msg)
		m.currentView = mainView
		return m, tea.Batch(m.updateRuleList(), func() tea.Msg { m.updatePortForwardingList(); return nil })

	case configExportedMsg:
		m.statusMessage = string(msg)
		m.currentView = mainView
		return m, nil

	case configSavedAndBackToMainMsg:
		m.statusMessage = string(msg)
		m.currentView = mainView
		return m, nil

	case fileListMsg:
		m.fileList.SetItems(msg)
		return m, nil

	case errMsg:
		m.statusMessage = msg.Error()
		return m, nil
	}

	return m, cmd
}

func (m *model) View() string {
	switch m.currentView {
	case confirmationView:
		return m.confirmationView()
	case mainView:
		return m.mainView()
	case ruleListView:
		return m.ruleListView()
	case ruleFormView:
		return m.ruleFormView()
	case portForwardingListView:
		return m.portForwardingListView()
	case portForwardingFormView:
		return m.portForwardingFormView()
	case infoView:
		return m.infoView()
	case saveConfigView:
		return m.saveConfigView()
	case importConfigView:
		return m.importConfigView()
	default:
		return "Unknown view"
	}
}

func (m *model) mainView() string {
	var s strings.Builder
	status := fmt.Sprintf("PF Status: %s | Startup: %s", m.pfStatus, m.startupStatus)
	s.WriteString(statusStyle.Render(status))
	s.WriteString("\n\n")
	s.WriteString(m.list.View())
	s.WriteString("\n")
	s.WriteString(m.statusMessage)
	return appStyle.Render(s.String())
}

func (m *model) confirmationView() string {
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.confirmationMessage,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render("(y/n)"),
		),
	)
}

func (m *model) ruleListView() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Firewall Rules"))
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Padding(0, 1).Render("  #   Action  Dir   Q   Proto   Source          Dest            Port       S   Description"))
	s.WriteString("\n")
	m.ruleList.SetItems(m.getRuleListItems())
	s.WriteString(m.ruleList.View())
	s.WriteString(`
  Arrows: Navigate | a: Add | Enter: Edit | d: Delete | k/j: Move Up/Down | s: Save order | Esc: Cancel`)
	return appStyle.Render(s.String())
}

func (m *model) ruleFormView() string {
	var b strings.Builder
	b.WriteString("  Add/Edit Firewall Rule\n\n")

	fields := []struct {
		label    string
		isInput  bool
		options  []string
		selected string
		input    *textinput.Model
	}{
		{"Action", false, []string{"block", "pass"}, m.form.action, nil},
		{"Direction", false, []string{"in", "out"}, m.form.direction, nil},
		{"Quick", false, []string{"Yes", "No"}, m.form.quick, nil},
		{"Interface", true, nil, "", &m.form.interfaceInput},
		{"Protocol", false, []string{"tcp", "udp", "tcp,udp", "icmp", "any"}, m.form.protocol, nil},
		{"Source", true, nil, "", &m.form.sourceInput},
		{"Destination", true, nil, "", &m.form.destinationInput},
		{"Port", true, nil, "", &m.form.portInput},
		{"Keep State", false, []string{"Yes", "No"}, m.form.keepState, nil},
		{"Description", true, nil, "", &m.form.descriptionInput},
	}

	for i, field := range fields {
		isFocused := m.form.focused == i
		if field.isInput {
			b.WriteString(renderInput(field.label, *field.input, isFocused, m.form.activeTextInput, i, field.label))
		} else {
			b.WriteString(renderOptions(field.label, field.options, field.selected, isFocused))
		}
	}

	b.WriteString("\n\n    Instructions:\n")
	b.WriteString("    Up/Down: Navigate fields\n")
	b.WriteString("    Left/Right: Change value for fields with options\n")
	b.WriteString("    Enter: Toggle text input edit mode\n")
	b.WriteString("    's': Save rule | Esc: Cancel\n")

	return appStyle.Render(b.String())
}

func (m *model) portForwardingListView() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Port Forwarding Rules"))
	s.WriteString("\n")
		
	s.WriteString("\n")
	s.WriteString(m.portForwardingList.View())
	s.WriteString(`
  Arrows: Navigate | a: Add | Enter: Edit | d: Delete | k/j: Move Up/Down | s: Save order | Esc: Cancel`)
	return appStyle.Render(s.String())
}

type portForwardingForm struct {
	focused           int
	activeTextInput   int // -1 if no text input is active, otherwise the index of the active text input
	isNew             bool
	ruleIndex         int
	protocol          string
	interfaceInput    textinput.Model
	externalIPInput   textinput.Model
	externalPortInput textinput.Model
	internalIPInput   textinput.Model
	internalPortInput textinput.Model
	descriptionInput  textinput.Model
}

func (m *model) portForwardingFormView() string {
	var b strings.Builder
	b.WriteString("  Add/Edit Port Forwarding Rule\n\n")

	fields := []struct {
		label    string
		isInput  bool
		options  []string
		selected string
		input    *textinput.Model
	}{
		{"Interface", true, nil, "", &m.portForwardingForm.interfaceInput},
		{"Protocol", false, []string{"tcp", "udp"}, m.portForwardingForm.protocol, nil},
		{"External IP", true, nil, "", &m.portForwardingForm.externalIPInput},
		{"External Port", true, nil, "", &m.portForwardingForm.externalPortInput},
		{"Internal IP", true, nil, "", &m.portForwardingForm.internalIPInput},
		{"Internal Port", true, nil, "", &m.portForwardingForm.internalPortInput},
		{"Description", true, nil, "", &m.portForwardingForm.descriptionInput},
	}

	for i, field := range fields {
		isFocused := m.portForwardingForm.focused == i
		if field.isInput {
			b.WriteString(renderInput(field.label, *field.input, isFocused, m.portForwardingForm.activeTextInput, i, field.label))
		} else {
			b.WriteString(renderOptions(field.label, field.options, field.selected, isFocused))
		}
	}

	b.WriteString("\n\n    Instructions:\n")
	b.WriteString("    Up/Down: Navigate fields\n")
	b.WriteString("    Left/Right: Change value for fields with options (e.g., Protocol)\n")
	b.WriteString("    Enter: Toggle text input edit mode\n")
	b.WriteString("    's': Save rule | Esc: Cancel\n")

	return appStyle.Render(b.String())
}

func (m *model) focusRuleForm() {
	// Blur all text inputs first
	m.form.interfaceInput.Blur()
	m.form.sourceInput.Blur()
	m.form.destinationInput.Blur()
	m.form.portInput.Blur()
	m.form.descriptionInput.Blur()

	// If a text input is active, focus only that one
	if m.form.activeTextInput != -1 {
		switch m.form.activeTextInput {
		case 3:
			m.form.interfaceInput.Focus()
		case 5:
			m.form.sourceInput.Focus()
		case 6:
			m.form.destinationInput.Focus()
		case 7:
			m.form.portInput.Focus()
		case 9:
			m.form.descriptionInput.Focus()
		}
	} else { // Otherwise, ensure no text input is focused
		m.form.interfaceInput.Blur()
		m.form.sourceInput.Blur()
		m.form.destinationInput.Blur()
		m.form.portInput.Blur()
		m.form.descriptionInput.Blur()
	}
}

func (m *model) focusPortForwardingForm() {
	// Blur all text inputs first
	m.portForwardingForm.interfaceInput.Blur()
	m.portForwardingForm.externalIPInput.Blur()
	m.portForwardingForm.externalPortInput.Blur()
	m.portForwardingForm.internalIPInput.Blur()
	m.portForwardingForm.internalPortInput.Blur()
	m.portForwardingForm.descriptionInput.Blur()

	// If a text input is active, focus only that one
	if m.portForwardingForm.activeTextInput != -1 {
		switch m.portForwardingForm.activeTextInput {
		case 0:
			m.portForwardingForm.interfaceInput.Focus()
		case 2:
			m.portForwardingForm.externalIPInput.Focus()
		case 3:
			m.portForwardingForm.externalPortInput.Focus()
		case 4:
			m.portForwardingForm.internalIPInput.Focus()
		case 5:
			m.portForwardingForm.internalPortInput.Focus()
		case 6:
			m.portForwardingForm.descriptionInput.Focus()
		}
	} else { // Otherwise, ensure no text input is focused
		m.portForwardingForm.interfaceInput.Blur()
		m.portForwardingForm.externalIPInput.Blur()
		m.portForwardingForm.externalPortInput.Blur()
	m.portForwardingForm.internalIPInput.Blur()
	m.portForwardingForm.internalPortInput.Blur()
	m.portForwardingForm.descriptionInput.Blur()
	}
}

func (m *model) infoView() string {
	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
						titleStyle.Render(m.infoViewTitle),
			m.viewport.View(),
		),
	)
}

func (m *model) saveConfigView() string {
	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			"Export Configuration As...",
			m.textinput.View(),
			"(Enter to save, Esc to cancel)",
		),
	)
}

func (m *model) importConfigView() string {
	return appStyle.Render(m.fileList.View())
}

type fileInfo struct {
	name    string
	modTime time.Time
}

func (i fileInfo) Title() string       { return i.name }
func (i fileInfo) Description() string { return i.modTime.Format("2006-01-02 15:04:05") }
func (i fileInfo) FilterValue() string { return i.name }


type ruleListItem struct {
	rule  FirewallRule
	index int
}

func (i ruleListItem) Title() string {
	quick := ""
	if i.rule.Quick {
		quick = "Y"
	}
	keepState := ""
	if i.rule.KeepState {
		keepState = "Y"
	}

	return fmt.Sprintf("%3d  %-7s %-5s %-3s %-7s %-15s %-15s %-10s %-3s %s",
		i.index+1,
		i.rule.Action,
		i.rule.Direction,
		quick,
		i.rule.Protocol,
		i.rule.Source,
		i.rule.Destination,
		i.rule.Port,
		keepState,
		i.rule.Description,
	)
}
func (i ruleListItem) Description() string { return "" }
func (i ruleListItem) FilterValue() string { return i.rule.Description }

type portForwardingListItem struct {
	rule  PortForwardingRule
	index int
}

func (i portForwardingListItem) Title() string {
	return fmt.Sprintf("%3d  %-15s %-7s %-15s:%-5s -> %-15s:%-5s %s",
		i.index+1,
		i.rule.Interface,
		i.rule.Protocol,
		i.rule.ExternalIP,
		i.rule.ExternalPort,
		i.rule.InternalIP,
		i.rule.InternalPort,
		i.rule.Description,
	)
}

func (i portForwardingListItem) Description() string { return "" }
func (i portForwardingListItem) FilterValue() string { return i.rule.Description }

func (m *model) getRuleListItems() []list.Item {
	items := []list.Item{}
	for i, rule := range m.firewallManager.Config.FirewallRules {
		items = append(items, ruleListItem{rule: rule, index: i})
	}
	return items
}

func (m *model) updateRuleList() tea.Cmd {
	items := []list.Item{}
	for i, rule := range m.firewallManager.Config.FirewallRules {
		items = append(items, ruleListItem{rule: rule, index: i})
	}
	m.ruleList.SetItems(items)
	return nil
}

func (m *model) updateFileList() tea.Cmd {
	return func() tea.Msg {
		configPath, _ := GetConfigPath()
		LogInfo(fmt.Sprintf("Reading files from: %s", configPath))
		files, err := os.ReadDir(configPath)
		if err != nil {
			LogError(fmt.Sprintf("Error reading config directory: %v", err))
			return errMsg{err}
		}

		var fileInfos []fileInfo
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") && file.Name() != "rules.json" {
				info, err := file.Info()
				if err == nil {
					fileInfos = append(fileInfos, fileInfo{name: file.Name(), modTime: info.ModTime()})
				} else {
					LogError(fmt.Sprintf("Error getting file info: %v", err))
				}
			}
		}

		sort.Slice(fileInfos, func(i, j int) bool {
			return fileInfos[i].modTime.After(fileInfos[j].modTime)
		})

		items := make([]list.Item, len(fileInfos))
		for i, fi := range fileInfos {
			items[i] = fi
		}

		LogInfo(fmt.Sprintf("Found %d JSON files", len(items)))
		return fileListMsg(items)
	}
}

func (m *model) updatePortForwardingList() {
	items := []list.Item{}
	for i, rule := range m.firewallManager.Config.PortForwardingRules {
		items = append(items, portForwardingListItem{rule: rule, index: i})
	}
	m.portForwardingList.SetItems(items)
}

func (m *model) saveRule() tea.Cmd {
	rule := FirewallRule{
		Action:      m.form.action,
		Direction:   m.form.direction,
		Quick:       m.form.quick == "Yes",
		Interface:   m.form.interfaceInput.Value(),
		Protocol:    m.form.protocol,
		Source:      m.form.sourceInput.Value(),
		Destination: m.form.destinationInput.Value(),
		Port:        m.form.portInput.Value(),
		KeepState:   m.form.keepState == "Yes",
		Description: m.form.descriptionInput.Value(),
	}

	var cmd tea.Cmd
	if m.form.isNew {
		cmd = func() tea.Msg {
			if err := m.firewallManager.AddFirewallRule(rule); err != nil {
				return errMsg{err}
			}
			return firewallRuleSavedMsg("Rule added successfully.")
		}
	} else {
		cmd = func() tea.Msg {
			if err := m.firewallManager.UpdateFirewallRule(m.form.ruleIndex, rule); err != nil {
				return errMsg{err}
			}
			return firewallRuleSavedMsg("Rule updated successfully.")
		}
	}

	return cmd
}

func (m *model) savePortForwardingRule() tea.Cmd {
	rule := PortForwardingRule{
		Interface:    m.portForwardingForm.interfaceInput.Value(),
		Protocol:     m.portForwardingForm.protocol,
		ExternalIP:   m.portForwardingForm.externalIPInput.Value(),
		ExternalPort: m.portForwardingForm.externalPortInput.Value(),
		InternalIP:   m.portForwardingForm.internalIPInput.Value(),
		InternalPort: m.portForwardingForm.internalPortInput.Value(),
		Description:  m.portForwardingForm.descriptionInput.Value(),
	}

	var cmd tea.Cmd
	if m.portForwardingForm.isNew {
		cmd = func() tea.Msg {
			if err := m.firewallManager.AddPortForwardingRule(rule); err != nil {
				return errMsg{err}
			}
			return portForwardingRuleSavedMsg("Port forwarding rule added successfully.")
		}
	} else {
		cmd = func() tea.Msg {
			if err := m.firewallManager.UpdatePortForwardingRule(m.portForwardingForm.ruleIndex, rule); err != nil {
				return errMsg{err}
			}
			return portForwardingRuleSavedMsg("Port forwarding rule updated successfully.")
		}
	}

	return cmd
}
