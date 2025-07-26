# Simple Notes App on top of Confluence

A simple command-line application to manage notes using Confluence as a backend. This app allows you to create, read, update, and delete notes stored in Confluence.

It also has a Web interface to view and manage notes.

# Quick Start

Start the app:

```bash
go mod init notes-app
go get
```

Commands:
```bash
# List all notes
go run main.go -action list

# Create a new note
go run main.go -action create -title "My Note" -content "<p>This is my note content</p>"

# Read a specific note
go run main.go -action read -id 123456789

# Update a note
go run main.go -action update -id 123456789 -title "Updated Note" -content "<p>Updated content</p>"

# Delete a note
go run main.go -action delete -id 123456789
```
