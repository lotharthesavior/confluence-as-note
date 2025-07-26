package main

import (
	"flag"
	"fmt"
	"net/http"
	"notes-app/src/cli"
	config2 "notes-app/src/config"
	"notes-app/src/web"
	"os"
	"time"
)

func main() {
	configFile := flag.String("config", "config.json", "Path to configuration file")
	action := flag.String("action", "list", "Action to perform (list, create, read, update, delete)")
	title := flag.String("title", "", "Note title")
	content := flag.String("content", "", "Note content")
	id := flag.String("id", "", "Note ID for read/update")
	flag.Parse()

	// Load configuration
	config, err := config2.LoadConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	switch *action {
	case "web":
		web.StartWebServer(client, config)
	default:
		cli.RouteCall(*action, client, config, *title, *content, *id)
	}
}
