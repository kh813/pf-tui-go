# `pf-tui-go` Implementation Plan

This document outlines the tasks required to implement the features described in `features.md` using Go and the `bubbletea` framework.

**Project Philosophy Note:** As per `features.md`, this project will maintain a simple and flat file structure. TUI components will be consolidated into a minimal number of files (e.g., a single `tui.go`) rather than being split into many separate files.

## Phase 1: Project Setup & Core Structures

-   [x] **Initialize Project:**
    -   [x] Set up `go.mod`.
    -   [x] Add dependencies: `bubbletea`, `bubbles`, `lipgloss`.

-   [x] **Define Core Data Structures (`firewall.go`):**
    -   [x] Create `FirewallRule` struct to represent a single filter rule.
    -   [x] Create `PortForwardingRule` struct for RDR rules.
    -   [x] Create a `Config` struct to hold both `FirewallRule` and `PortForwardingRule` slices.

-   [x] **Firewall Manager (`firewall.go`):**
    -   [x] Create `FirewallManager` struct.
    -   [x] Implement `LoadConfig(path string)` to read rules from JSON.
    -   [x] Implement `SaveConfig(path string)` to write rules to JSON.
    -   [x] Implement `GeneratePfConf()` to create the `pf.conf` string from the current rules.

## Phase 2: Backend `pfctl` Integration

-   [x] **Implement `pfctl` Commands (`pf.go`):**
    -   [x] `ApplyRules(rules string)`: Apply rules from a string.
    -   [x] `GetCurrentRules()`: Get currently loaded rules.
    -   [x] `GetPfStatus()`: Check if PF is enabled or disabled.
    -   [x] `EnablePf()`: Enable the firewall.
    -   [x] `DisablePf()`: Disable the firewall.
    -   [x] `GetPfInfo()`: Get detailed statistics.
    -   [x] `EnablePfOnStartup()`: Configure `launchd` to enable PF on boot.
    -   [x] `DisablePfOnStartup()`: Remove the `launchd` configuration.
    -   [x] **Sudo Handling:**
    -   [x] Implement a pre-flight check to validate `sudo` credentials.
    -   [x] If a password is required, temporarily pause the TUI and prompt for the password in the standard terminal.

## Phase 3: TUI Implementation (`tui.go`)

-   [x] **Main Model:**
    -   [x] Create the main `Model` struct that will manage the application state (e.g., different views/screens).
    -   [x] Implement the main `Init`, `Update`, and `View` methods.
    -   [x] The `Update` function will delegate to logic for the currently active view.
    -   [x] The `View` function will call the rendering logic for the active view.

-   [x] **Main Menu Layout:**
    -   [x] Add separators to the main menu for better visual organization.
    -   [x] Add a blank line at the top of the menu.
    -   [x] Display PF status at the top of the main menu.
    -   [x] Implement circular navigation for the main menu (up from top goes to bottom, down from bottom goes to top).

-   [x] **Add/Edit Rule Form:**
    -   [x] Create a model/struct for the form using `textinput.Model` for fields.
    -   [x] Implement logic for cycling through options (e.g., `block`/`pass`).
    -   [x] Pre-populate with default values when creating a new rule.
    -   [x] On save, validate and send a message to update the main state.
    -   [x] Update the layout to match the Python version.
    -   [x] Fix editing and display of values.
    -   [x] Show all available options and highlight the selected one.
    -   [x] Update 'Add/Edit Rule' form UI to match the Python version's layout, including inline options and highlighting.

-   [x] **Edit Rule List Screen:**
    -   [x] Use `list.Model` to display `FirewallRule` items.
    -   [x] Handle keys for adding (`a`), editing (`enter`), deleting (`d`), and reordering (`k`/`j`).
    -   [x] Handle saving the new order (`s`).
    -   [x] Display rules in a formatted table.
    -   [x] Display rules in a formatted table.
    -   [x] Add help text and re-ordering functionality.
    -   [x] Save and return to main menu on 's'.

-   [x] **Parse and Display Live Rules:**
    -   [x] Implement a parser for `pfctl -s rules` output.
    -   [x] Fetch and display live rules in the "Edit Rule" screen.

-   [x] **Port Forwarding Screens:**
    -   [x] Implement the form and list views for Port Forwarding rules, similar to the firewall rules screens.
    -   [x] Ensure required fields are validated.
    -   [x] Update 'Add/Edit Port Forwarding Rule' form UI to match the 'Add/Edit Rule' form's layout, using inline options and highlighting for the 'Protocol' field.

-   [x] **Informational Screens:**
    -   [x] Use a `viewport.Model` to display content.
    -   [x] For "Show Info", use a `tea.Tick` to refresh the content periodically.
    -   [x] Dynamically set title for "Show Current Rules" and "Show Info" screens.

-   [x] **Configuration Screens:**
    -   [x] **Export Configuration:** Use `textinput.Model` to get a file path, with confirmation for overwriting. Default to `~/.config/pf-tui/rules-export-YYYYMMDD-HHMMSS.json`.
    -   [x] **Import:** Use a custom file list to select a file. Implement backup of existing `rules.json`. Start file picker in `~/.config/pf-tui/`.

## Phase 4: Main Entrypoint & Finalization

-   [x] **Main Entrypoint (`main.go`):**
    -   [x] Initialize the `FirewallManager`.
    -   [x] Load rules from config at startup.
    -   [x] Initialize and run the `bubbletea` program.

-   [x] **Styling, Error Handling, and Help:**
    -   [x] Use `lipgloss` for styling.
    -   [x] Implement a status bar/dialogs for messages and errors.
    -   [x] Use `help.Model` for context-sensitive keybindings.
    -   [x] Implement global `q` and `esc` hotkeys.
        - [x] Add confirmation dialog on exit.
    -   [x] Pressing `esc` on main menu prompts for exit.
    -   [x] Pressing `esc` in the rule form prompts for exit.

-   [x] **Code Review & Refactoring:**
    -   [x] Clean up code, add comments, and verify all features are implemented.

## Phase 5: Configuration Handling Refactor

-   [x] **Save & Apply Configuration:**
    -   [x] Save the current rules to `~/.config/pf-tui/rules.json`.
    -   [x] Apply the rules to the live system firewall.
-   [x] **Refactor Rule Editing:**
    -   [x] Modify the "Add/Edit Rule" form's save logic to write changes directly to `~/.config/pf-tui/rules.json`.
    -   [x] After saving, reload the rules in the main model to reflect the change.
-   [x] **Refactor Rule List Management:**
    -   [x] Update the "Edit Rule List" screen's delete action (`d`) to remove the rule from `~/.config/pf-tui/rules.json` immediately.
    -   [x] Update the "Edit Rule List" screen's save order action (`s`) to rewrite the rules in the new order to `~/.config/pf-tui/rules.json`.
-   [x] **Refactor Port Forwarding Rule Management:**
    -   [x] Apply the same direct-to-JSON save/delete/reorder logic to the port forwarding rule screens.
-   [x] **Configuration Loading:**
    -   [x] Ensure that rule lists are reloaded from `~/.config/pf-tui/rules.json` whenever returning to the main menu or entering a rule editing screen, to ensure consistency.

## Phase 6: UI Implementation

-   [x] **Main Menu:**
    -   [x] Use a `list.Model` to display the main menu options.
    -   [x] Handle key presses to navigate the menu and select options.
-   [x] **Rule & Port Forwarding Lists:**
    -   [x] Use a `list.Model` to display the rules.
    -   [x] Implement custom list items to format the rule data into columns.
-   [x] **Rule & Port Forwarding Forms:**
    -   [x] Use a slice of `textinput.Model` to create the input fields for the forms.
    -   [x] Manage focus between the input fields.
-   [x] **Info & Current Rules Screens:**
    -   [x] Use a `viewport.Model` to display the output of the `pfctl` commands.
-   [x] **File Picker:**
    -   [x] Use a `filepicker.Model` for the "Import Configuration" screen.
-   [x] **Confirmation Dialogs:**
    -   [x] Use `promptkit/confirmation` to create confirmation dialogs for exiting, discarding changes, and overwriting files.

## Phase 7: Finalization

-   [x] **Makefile:**
    -   [x] Create a `Makefile` with targets for building, running, and cleaning the project.
-   [x] **README:**
    -   [x] Update the `README.md` file with instructions on how to build and run the application.
-   [x] **Testing:**
    -   [x] Write tests for the `firewall.go` and `pf.go` modules.
-   [x] **Release:**
    -   [x] Create a new release on GitHub with the compiled binaries.

## Phase 8: Fixes and Refinements

-   [x] **"Edit Rule" Screen:**
    -   [x] Restore the "Edit Rule" menu item.
    -   [x] Implement the functionality to edit existing rules.
    -   [x] Display the list of existing rules from `~/.config/pf-tui/rules.json`.
    -   [x] Ensure the table headers and layout match the Python version.
-   [x] **Main Menu:**
    -   [x] Add a title "Main menu" at the top of the main menu.
-   [x] **UI Refinements:**
    -   [x] Remove prompt character `>` from text input fields.
    -   [x] Fix indentation of text input fields in forms.
    -   [x] Fix rule list view to match Python version.
    -   [x] Increase rule list display capacity to 999 items.
    -   [x] Increase port forwarding rule list display capacity to 999 items.
-   [x] **Port Forwarding Form:**
    -   [x] Update 'Add/Edit Port Forwarding Rule' form UI to match the 'Add/Edit Rule' form, including inline options and highlighting for the 'Protocol' field.
-   [x] **Configuration Management:**
    -   [x] Verify that "Edit Rule" and "Add Rule" correctly modify `~/.config/pf-tui/rules.json`.
    -   [x] Verify that "Edit Forwarding Rule" and "Add Forwarding Rule" correctly modify `~/.config/pf-tui/rules.json`.
    -   [x] Verify that "Import config" and "Save Configuration" work as expected.
    -   [x] Verify that "Save & Apply Configuration" works as expected.
-   [x] **Documentation:**
    -   [x] Update `features.md` and `todo.md` to reflect the latest changes.
-   [x] **Testing:**
    -   [x] Add a `-test` flag to bypass `sudo` checks for UI testing.

-   [x] **Navigation:**
    -   [x] After saving a new firewall rule, navigate to the firewall rule list.
    -   [x] After saving a new port forwarding rule, navigate to the port forwarding rule list.

-   [x] **Form Interaction:**
    -   [x] Automatically focus text input fields when selected in the rule and port forwarding forms.
-   [x] Implement save logic for rule and port forwarding forms: 's' saves only after 'enter' finalizes text input fields.
-   [x] Implement text input field interaction: 'Enter' to toggle editing mode, 'Enter' again to finalize input and unfocus.

## Phase 9: Bug Fixes

-   [x] **Import Configuration:**
    -   [x] Replace `filepicker` with a custom file list that shows all `.json` files in `~/.config/pf-tui/`, excluding `rules.json`.
    -   [x] The list should be sorted by modification time, with the newest file at the top.
    -   [x] The newest file should be selected by default.
    -   [x] Remove empty lines from the list.

## Phase 10: Logging

-   [x] **Log Rotation:**
    -   [x] Rotate the log file daily at midnight, keeping up to 30 old log files.
-   [x] **Log Cleanup:**
    -   [x] Remove log files older than 90 days on application startup.

## Phase 11: Enhanced Logging

-   [x] **Structured Logging:**
    -   [x] Implement `INFO`, `WARN`, and `ERROR` log levels.
    -   [x] Add logging for key application events (startup, shutdown, config changes, etc.).
    -   [x] Add logging for potential issues (config not found, empty rules, etc.).
    -   [x] Add detailed logging for all critical errors (file I/O, `pfctl` commands, etc.).

## Phase 13: Bug Fixes

-   [x] **Fix `pf.conf` generation:**
    -   [x] Fix the logic for generating `pf.conf` from the JSON configuration.
    -   [x] Debug and fix remaining `pf.conf` syntax errors in RDR rules.
    -   [x] Fix syntax error in `pf.go` for printing generated `pf.conf`.
    -   [x] Remove unused variable `externalIPStr` in `firewall.go`.
    -   [x] Analyze generated `pf.conf` content to identify syntax errors.
    -   [x] Verify `pf.conf` generation by running the application and checking the output.
    -   [x] Verify `pf.conf` generation by running the application and checking the output.

## Phase 14: Firewall Rules Screen Enhancement

-   [x] **Highlighting:**
    -   [x] Ensure the selected item in the firewall rule list is clearly highlighted.
    -   [x] Fix highlighting after reordering items. (Reordering logic updated, but highlighting might still have issues)
-   [x] **Reordering:**
    -   [x] Implement moving items up (`k`) and down (`j`) in the list.

## Phase 15: Bug Fixes

-   [x] **"Show Info" screen:**
    -   [x] Fix "Show Info" screen to refresh every second only when PF is enabled.
