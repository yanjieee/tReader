# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

tReader is a stealth novel reader designed to look like a legitimate terminal-based system monitoring tool. It displays fake system logs while embedding novel content in barely visible gray text, allowing users to read novels discretely during work hours.

## Architecture

The application is built as a single Go binary using the tview terminal UI library. Key architectural components:

- **App struct**: Central application state containing UI components, reading state, and novel content
- **Mixed Display Mode**: Novel content is interspersed with fake system logs using different color schemes
- **Dynamic Layout**: Uses tview.Flex to dynamically show/hide search input overlay
- **Encoding Detection**: Smart UTF-8/GBK detection for Chinese text files
- **State Management**: Tracks reading position, display mode, opacity levels, and search state

## Development Commands

### Build and Run
```bash
go build -o tReader                    # Build the binary
./tReader                              # Run with default novel.txt
./tReader path/to/novel.txt           # Run with custom file
```

### Dependencies
```bash
go mod tidy                           # Update dependencies
go mod download                       # Download dependencies
```

## Code Organization

The single `main.go` file is organized into clearly marked sections:

- **Application Initialization**: App struct and setup
- **File Loading**: Text encoding detection and content parsing
- **Search Functionality**: Text search and navigation
- **String Processing**: UTF-8 safe text truncation and line splitting
- **UI Management**: Display rendering and fake log generation
- **Keyboard Controls**: Input handling and mode switching

## Key Features Implementation

### Stealth Mode
- Fake system logs are generated with realistic timestamps and technical messages
- Novel content appears as extremely faint gray text between log lines
- Boss key ('h') instantly switches between stealth and reading modes

### Text Processing
- Long lines are automatically split at 70-character boundaries with UTF-8 safety
- Supports both UTF-8 and GBK encoding with automatic detection
- Novel content is loaded into memory and indexed for navigation

### Search System
- Search input overlay appears/disappears dynamically
- Case-insensitive text matching with instant navigation
- Search results automatically adjust reading position

## File Format Support

Only plain text (.txt) files are supported with UTF-8 and GBK encoding detection. The encoding detection logic:
1. First attempts UTF-8 validation on raw bytes
2. If UTF-8 fails, attempts GBK decoding
3. Falls back to UTF-8 with replacement characters if both fail

## UI Controls

The application uses vim-style navigation:
- h: Toggle reading mode (boss key)
- j/k or ↑↓: Navigate text
- [/]: Adjust opacity (10 levels)
- /: Search functionality
- PgUp/PgDn: Fast navigation

## Development Notes

- The application maintains a 25-line fake log buffer that continuously updates
- Display refresh happens on mode changes, navigation, and periodically for fake logs
- All text rendering uses tview's markup system for color control
- The search functionality uses a separate input field that overlays the main view