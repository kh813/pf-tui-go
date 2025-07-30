package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	
)

var testMode bool

func main() {
	// Logging is initialized in logging.go's init() function

	if err := EnsureConfigDirExists(); err != nil {
				LogError(fmt.Sprintf("Error creating config directory: %v", err))
		fmt.Printf("Error creating config directory: %v\n", err)
		os.Exit(1)
	}

		flag.BoolVar(&testMode, "test", false, "Enable test mode to bypass sudo checks")
	flag.Parse()

	if testMode {
		os.Setenv("TERM", "dumb")
	}

	LogInfo(fmt.Sprintf("Test mode: %t", testMode))

	// Check for sudo credentials before starting the TUI
	if !testMode {
		if err := checkSudo(); err != nil {
			LogError(fmt.Sprintf("Error with sudo: %v", err))
			fmt.Printf("Error with sudo: %v\n", err)
			os.Exit(1)
		}
	}

	// Initialize the firewall manager
	fm := NewFirewallManager()

	// Load the initial configuration
	if err := fm.LoadConfig(); err != nil {
		LogWarn(fmt.Sprintf("Error loading initial config: %v", err))
	}

	// Initialize the Bubble Tea program
	programOpts := []tea.ProgramOption{}
	if !testMode {
		programOpts = append(programOpts, tea.WithAltScreen())
	} else {
		programOpts = append(programOpts, tea.WithoutRenderer())
	}
		p := tea.NewProgram(NewModel(fm), programOpts...)

	LogInfo("Attempting to run the Bubble Tea program.")

	// Run the program
	if !testMode {
		if _, err := p.Run(); err != nil {
			LogError(fmt.Sprintf("Bubble Tea program exited with error: %v", err))
			LogError(fmt.Sprintf("Alas, there's been an error: %v", err))
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	} else {
		LogInfo("Skipping Bubble Tea program execution in test mode.")
	}
}

func checkSudo() error {
	// -n, --non-interactive
	// Avoid prompting the user for a password.  If a password is required for the command to run, sudo will display an error message and exit.
	cmd := exec.Command("sudo", "-n", "true")
	if err := cmd.Run(); err != nil {
		// If the command fails, it's likely because a password is required.
		// Prompt the user for their password in the terminal.
		fmt.Println("Sudo credentials required. Please enter your password.")
		cmd := exec.Command("sudo", "-v")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}
