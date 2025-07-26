package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	config2 "notes-app/src/config"
)

type Body struct {
	Storage struct {
		Value          string `json:"value"`
		Representation string `json:"representation"`
	} `json:"storage"`
}

type Page struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	SpaceID string `json:"spaceId"`
	Body    Body   `json:"body"`
	Version struct {
		Number int `json:"number"`
	}
}

type CreatePageRequest struct {
	SpaceID  string `json:"spaceId"`
	Status   string `json:"status"`
	Title    string `json:"title"`
	Body     Body   `json:"body"`
	ParentID string `json:"parentId"`
}

type PagesResponse struct {
	Results []Page `json:"results"`
	Links   struct {
		Next string `json:"next"`
	} `json:"_links"`
}

type UpdatePageRequest struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Body    Body   `json:"body"`
	Version struct {
		Number  int    `json:"number"`
		Message string `json:"message"`
	} `json:"version"`
}

func MakeRequest(client *http.Client, config *config2.Config, method, endpoint string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("https://%s/wiki/api/v2%s", config.Domain, endpoint)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshalling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+BasicAuth(config.Email, config.APIToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
