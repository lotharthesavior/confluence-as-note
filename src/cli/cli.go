package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	client2 "notes-app/src/client"
	config2 "notes-app/src/config"
	"os"
)

func RouteCall(action string, client *http.Client, config *config2.Config, title, content, id string) {
	switch action {
	case "list":
		ListNotes(client, config)
	case "create":
		if title == "" || content == "" {
			fmt.Println("Title and content are required for create action")
			os.Exit(1)
		}
		CreateNote(client, config, title, content)
	case "read":
		if id == "" {
			fmt.Println("ID is required for read action")
			os.Exit(1)
		}
		ReadNote(client, config, id)
	case "update":
		if id == "" || title == "" || content == "" {
			fmt.Println("ID, title, and content are required for update action")
			os.Exit(1)
		}
		UpdateNote(client, config, id, title, content)
	case "delete":
		if id == "" {
			fmt.Println("ID is required for delete action")
			os.Exit(1)
		}
		DeleteNote(client, config, id)
	default:
		fmt.Println("Invalid action. Supported actions: list, create, read, update, delete")
		os.Exit(1)
	}
}

func ListNotes(client *http.Client, config *config2.Config) {
	endpoint := fmt.Sprintf("/pages?space-id=%s&limit=25", config.SpaceID)
	for {
		resp, err := client2.MakeRequest(client, config, "GET", endpoint, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing notes: %v\n", err)
			return
		}

		var pages client2.PagesResponse
		if err := json.NewDecoder(resp.Body).Decode(&pages); err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
			resp.Body.Close()
			return
		}
		resp.Body.Close()

		for _, page := range pages.Results {
			fmt.Printf("ID: %s, Title: %s, Status: %s\n", page.ID, page.Title, page.Status)
		}

		if pages.Links.Next == "" {
			break
		}
		endpoint = pages.Links.Next
	}
}

func CreateNote(client *http.Client, config *config2.Config, title, content string) {
	body := client2.CreatePageRequest{
		SpaceID:  config.SpaceID,
		Status:   "current",
		Title:    title,
		ParentID: config.ParentPageID,
		Body: client2.Body{
			Storage: struct {
				Value          string `json:"value"`
				Representation string `json:"representation"`
			}{
				Value:          content,
				Representation: "storage",
			},
		},
	}

	resp, err := client2.MakeRequest(client, config, "POST", "/pages", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating note: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var page client2.Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		return
	}

	fmt.Printf("Note created successfully. ID: %s, Title: %s\n", page.ID, page.Title)
}

func ReadNote(client *http.Client, config *config2.Config, id string) {
	resp, err := client2.MakeRequest(client, config, "GET", fmt.Sprintf("/pages/%s?body-format=storage", id), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading note: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var page client2.Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\nTitle: %s\nStatus: %s\nContent: %s\n", page.ID, page.Title, page.Status, page.Body.Storage.Value)
}

func UpdateNote(client *http.Client, config *config2.Config, id, title, content string) {
	// First, get current version number
	resp, err := client2.MakeRequest(client, config, "GET", fmt.Sprintf("/pages/%s", id), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting note: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var currentPage client2.Page
	if err := json.NewDecoder(resp.Body).Decode(&currentPage); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		return
	}

	body := client2.UpdatePageRequest{
		ID:     id,
		Title:  title,
		Status: "current",
		Body: client2.Body{
			Storage: struct {
				Value          string `json:"value"`
				Representation string `json:"representation"`
			}{
				Value:          content,
				Representation: "storage",
			},
		},
		Version: struct {
			Number  int    `json:"number"`
			Message string `json:"message"`
		}{
			Number:  currentPage.Version.Number + 1,
			Message: "Updated via Notes App",
		},
	}

	resp, err = client2.MakeRequest(client, config, "PUT", fmt.Sprintf("/pages/%s", id), body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating note: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Note updated successfully. ID: %s\n", id)
}

func DeleteNote(client *http.Client, config *config2.Config, id string) {
	resp, err := client2.MakeRequest(client, config, "DELETE", fmt.Sprintf("/pages/%s", id), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting note: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Note deleted successfully. ID: %s\n", id)
}
