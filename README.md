# pf-tui-go

A terminal user interface for managing the PF firewall on macOS.

![screenshot](https://user-images.githubusercontent.com/1234567/123456789-abcdef.png)

## Features

- **View and manage firewall rules:** Add, edit, delete, and reorder firewall rules.
- **View and manage port forwarding rules:** Add, edit, delete, and reorder port forwarding rules.
- **Enable and disable PF:** Easily enable or disable the PF firewall.
- **Enable and disable PF on startup:** Configure PF to start automatically on system boot.
- **Live status information:** View live information and statistics from the PF firewall.
- **Import and export rules:** Easily back up and restore your firewall configuration.
- **Sudo password prompt handling:** Automatically pauses the TUI to allow for password entry in the terminal, preventing UI conflicts.
- **Test mode:** Run the application without requiring `sudo` privileges for UI testing.

## Installation

### Prerequisites

- Go 1.16 or later
- macOS

### Building from source

1.  Clone the repository:

    ```bash
    git clone https://github.com/kh813/pf-tui-go.git
    cd pf-tui-go
    ```

2.  Build the application:

    ```bash
    make native
    ```

3.  Run the application:

    ```bash
    ./pf-tui
    ```

## Usage

The application is controlled using keyboard shortcuts. The available shortcuts are displayed at the bottom of each screen.

### Global Hotkeys

-   **`Esc`**: Cancel the current operation and return to the previous screen.
-   **`q`**: Quit the application.

## Configuration

On macOS, the configuration file is located at `~/.config/pf-tui/rules.json`. This file contains all of your firewall and port forwarding rules.

The application also keeps a log file at `~/.config/pf-tui/pf-tui.log`, which can be useful for troubleshooting.

## Development

### Building

To build the application, run:

```bash
make native
```

This will create a binary for your current platform in the project directory.

### Testing

To run the application in test mode, use the `-test` flag:

```bash
go run main.go -test
```

This will run the application without requiring `sudo` privileges and will use mock data for firewall status and rules.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
