# pf-tui Go Features

This document outlines the features and user interface of the `pf-tui` Go implementation.

## Project Philosophy

This is a small project and I'd like to keep the files and structure as simple and flat as possible.

## Global Hotkeys

- **`Esc`**: In most screens, this key cancels the current operation (e.g., editing a rule, browsing files) and returns to the previous screen or main menu. In a text input field, it cancels the edit. From the main menu, it will show a confirmation dialog to exit the application.
- **`q`**: From the main menu or informational screens, this key will show a confirmation dialog to quit the application.

## Main Screen

The initial screen provides a central menu for all major operations.

- **Status Display:** Shows the current status of the PF firewall (Enabled/Disabled) and whether it's enabled on startup. This is displayed at the top of the screen.
- **Navigation:** Use arrow keys to navigate the menu. Navigation is circular, meaning pressing up from the top item goes to the bottom, and pressing down from the bottom item goes to the top. This has been implemented.
- **Actions:**

    - **Edit Firewall Rule**
    - **Add New Firewall Rule**
    - **Edit Port Forwarding Rule**
    - **Add Port Forwarding Rule**
    - **Save & Apply Configuration**
    - **Export Configuration**
    - **Import Configuration**
    - **Show Current Rules**
    - **Show Info**
    - **Enable PF**
    - **Disable PF**
    - **Enable PF on Startup**
    - **Disable PF on Startup**
    - **Exit**

## Firewall Rule Screens

### Add/Edit Rule Screen

This screen provides a form to create or modify a firewall filter rule. When adding a new rule, fields are pre-populated with default values.

- **Fields (Default Value):**
    - **Action:** `block` or `pass` (Select with left/right arrows). (Default: `block`)
    - **Direction:** `in` or `out` (Select with left/right arrows). (Default: `in`)
    - **Quick:** `Yes` or `No` (Select with left/right arrows). (Default: `No`)
    - **Interface:** Network interface (e.g., `en0`) or `any` (Text input). (Default: `any`)
    - **Protocol:** `tcp`, `udp`, `tcp,udp`, `icmp`, or `any` (Select with left/right arrows). (Default: `any`)
    - **Source:** Source IP address, subnet, or `any` (Text input). (Default: `any`)
    - **Destination:** Destination IP address, subnet, or `any` (Text input). (Default: `any`)
    - **Port:** Port number, range (`-`), list (`,`), or `any` (Text input). For multiple ports or ranges, they will be enclosed in curly braces `{}` in the generated `pf.conf`. (Default: `any`)
    - **Keep State:** `Yes` or `No` (Select with left/right arrows). (Default: `No`)
    - **Description:** A brief description of the rule (Text input). (Default: empty)
- **Interaction:**
    - **Navigate:** Use up/down arrow keys to move between fields.
    - **Edit:** Press `Enter` to enter editing mode for text fields. Press `Enter` again to finalize input and exit editing mode.
    - **Save:** Press `'s'` to save the rule to `~/.config/pf-tui/rules.json`. If a text input field is active, press `Enter` to finalize the input before pressing `'s'` to save.
    - **Cancel:** Press `Esc` to show a confirmation dialog. Press `Enter` to confirm and return to the main menu.

### Edit Rule List Screen

This screen lists all configured firewall rules and allows for reordering and deletion.

- **Display:** Shows a list of all filter rules with their details in the following columns: `#`, `Action`, `Dir`, `Q`, `Proto`, `Source`, `Dest`, `Port`, `S`, `Description`. The list is capable of displaying up to 999 items.
- **Interaction:**
    - **Navigate:** Use up/down arrow keys to select a rule. The selected rule is highlighted.
    - **Add:** Press `'a'` to add a new rule.
    - **Edit:** Press `Enter` to open the selected rule in the "Add/Edit Rule Screen".
    - **Delete:** Press `'d'` to delete the selected rule from `~/.config/pf-tui/rules.json` (with confirmation).
    - **Move:** Use `k` (up) and `j` (down) to reorder rules.
    - **Save Order:** Press `'s'` to save the new rule order to `~/.config/pf-tui/rules.json`.

## Port Forwarding Rule Screens

### Add/Edit Port Forwarding Rule Screen

This screen provides a form to create or modify a port forwarding (RDR) rule. When adding a new rule, fields are pre-populated with default values.

- **Fields (Default Value):**
    - **Interface:** Network interface (e.g., `en0`) or `any` (Text input). (Default: `any`)
    - **Protocol:** `tcp` or `udp` (Select with left/right arrows). (Default: `tcp`)
    - **External IP:** The public-facing IP address (Text input). (Default: `any`)
    - **External Port:** The public-facing port (Text input). (Default: empty) **(Required)**
    - **Internal IP:** The internal IP address to forward to (Text input). (Default: `127.0.0.1`)
    - **Internal Port:** The internal port to forward to (Text input). (Default: empty) **(Required)**
    - **Description:** A brief description of the rule (Text input). (Default: empty)
- **Interaction:**
    - **Navigate:** Use up/down arrow keys to move between fields.
    - **Edit:** Press `Enter` to enter editing mode for text fields. Press `Enter` again to finalize input and exit editing mode.
    - **Save:** Press `'s'` to save the rule to `~/.config/pf-tui/rules.json`. If a text input field is active, press `Enter` to finalize the input before pressing `'s'` to save.
    - **Cancel:** Press `Esc` to show a confirmation dialog. Press `Enter` to confirm and return to the main menu.

**Note:** Fields marked as **(Required)** cannot be empty.


### Edit Port Forwarding Rule List Screen

This screen lists all configured RDR rules.

- **Display:** Shows a list of all port forwarding rules, capable of displaying up to 999 items.
- **Interaction:**
    - **Navigate:** Use up/down arrow keys.
    - **Add:** Press `'a'` to add a new rule.
    - **Edit:** Press `Enter` to edit the selected rule.
    - **Delete:** Press `'d'` to delete the selected rule from `~/.config/pf-tui/rules.json` (with confirmation).
    - **Move:** Press `'k'` (up) and `'j'` (down) to reorder.
    - **Save Order:** Press `'s'` to save the new order to `~/.config/pf-tui/rules.json`.

## Configuration Screens

### Export Configuration Screen

- **Action:** Prompts for a file path to save a copy of the current rule configuration. After saving, it returns to the main menu.
- **Default Value:** Defaults to `~/.config/pf-tui/rules-export-YYYYMMDD-HHMMSS.json`. The user can edit the path and filename.
- **Overwrite Confirmation:** Asks for confirmation if the specified file already exists.

### Import Configuration Screen

- **File Selector:** Opens a TUI file selector showing all `.json` files in the default configuration directory (`~/.config/pf-tui/`), excluding the default `rules.json` file.
- **Sorting:** The list of files is sorted by modification date, with the newest file at the top and selected by default.
- **Action:** Allows the user to select a JSON file to replace `~/.config/pf-tui/rules.json`. The existing file is backed up to `~/.config/pf-tui/rules.json.bak`.
- **Confirmation:** Shows a dialog with the result of the import operation.

## Informational Screens

### Show Current Rules Screen

- **Title:** "Current Live PF Rules"
- **Content:** Displays the output of `pfctl -s rules`, showing the rules currently active in the system's firewall. **Note: "ALTQ" related messages are filtered out.**
- **Interaction:** Read-only view. Press `Esc` or `'q'` to return to the main menu.

### Show Info Screen

- **Title:** "Live PF Info"
- **Content:** Displays the output of `pfctl -s info`, showing live, detailed statistics and status information from the `pf` firewall. If PF is enabled, the content is refreshed automatically every second. If PF is disabled, the content is not refreshed.
- **Interaction:** Read-only view. Press `Esc` or `'q'` to return to the main menu.

## Golang Tweaks

### Sudo Password Prompt Handling

-   **Problem:** When running the application, the `sudo` password prompt would conflict with the `bubbletea` TUI, causing the UI to render before the user could enter their password. This made the password prompt inaccessible.
-   **Solution:** To resolve this, the application performs a pre-flight check to validate `sudo` credentials. If a password is required, the TUI is temporarily paused, and the user is prompted for their password in the standard terminal. Once authenticated, the TUI resumes. This ensures a clean separation between the application's UI and system-level authentication.

### Test Mode

- **Flag:** `-test`
- **Purpose:** Allows running the application without requiring `sudo` privileges. When in test mode, the application will not execute any `pfctl` commands and will return mock data for firewall status and rules. This is useful for testing the UI and other non-sudo features.


## Go Implementation Details

The Go version of `pf-tui` is built with the `bubbletea` framework and follows the Model-View-Update (MVU) architecture.

### Core Data Models (`firewall.go`)

The application's logic is centered around a few core data structures:

-   **`FirewallRule`**: Represents a single firewall filter rule, containing fields like `Action`, `Direction`, `Protocol`, `Source`, `Destination`, etc. This struct is used for both in-memory representation and JSON serialization.
-   **`PortForwardingRule`**: Represents a single port forwarding (RDR) rule with fields for `Interface`, `Protocol`, `ExternalPort`, `InternalIP`, etc.
-   **`Config`**: A container struct that holds slices of `FirewallRule` and `PortForwardingRule`. This entire structure is what gets saved to and loaded from the `rules.json` configuration file.
-   **`FirewallManager`**: A manager struct that handles all operations related to the configuration, including loading from, saving to, and modifying the `rules.json` file. It also generates the `pf.conf` content from the current rules.

### TUI Model (`tui.go`)

### Logging

The application logs events to `pf-tui.log` to provide a clear history of its operations, which is useful for troubleshooting. The logs are categorized by severity: `INFO`, `WARN`, and `ERROR`.

-   **Log File Location:** `~/.config/pf-tui/pf-tui.log`
-   **Log Rotation:** The log file is automatically rotated daily at midnight, keeping up to 30 old log files.
-   **Log Retention:** Log files older than 90 days are automatically deleted at startup.

#### Logged Events:

-   **INFO:**
    -   Application startup and shutdown.
    -   Loading, saving, importing, and exporting configurations.
    -   Enabling/disabling PF and startup services.
    -   Log rotation and cleanup actions.
-   **WARN:**
    -   Configuration file not found.
    -   No rules found in the configuration.
    -   Non-critical warnings from `pfctl`.
-   **ERROR:**
    -   Failures in reading/writing/backing up configuration files.
    -   Failures in executing `pfctl` commands.
    -   Failures in parsing JSON configuration.
    -   Failures in log file operations.

The user interface is managed by a central `model` struct in `tui.go`. This struct holds the entire state of the application, including:

-   The current view being displayed (e.g., main menu, rule editor, info screen).
-   The `list.Model`, `viewport.Model`, `textinput.Model`, etc., from the `bubbles` library that manage different UI components.
-   The `FirewallManager` instance to interact with the firewall configuration.
-   Status information, such as the current PF status and any messages to be displayed to the user.
-   The application's main `Update` function acts as a state machine, handling messages (like user key presses or data loading) and updating the model's state accordingly. The `View` function then renders the UI based on the current state.


## Main menu
The main menu have separators, and it would look like this.

Main menu

  | Edit Rule
    Add New Rule
    Add Port Forwarding Rule
    Edit Port Forwarding Rule
    ---