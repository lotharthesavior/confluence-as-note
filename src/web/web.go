package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	client2 "notes-app/src/client"
	config2 "notes-app/src/config"
)

func StartWebServer(client *http.Client, config *config2.Config) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		listHandler(client, config, w, r)
	})
	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		createHandler(client, config, w, r)
	})
	http.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		readHandler(client, config, w, r)
	})
	http.HandleFunc("/edit", func(w http.ResponseWriter, r *http.Request) {
		editHandler(client, config, w, r)
	})
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		deleteHandler(client, config, w, r)
	})

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}

func getAllPages(client *http.Client, config *config2.Config) ([]client2.Page, error) {
	var pages []client2.Page
	endpoint := fmt.Sprintf("/pages?space-id=%s&limit=25&body-format=storage", config.SpaceID)
	for endpoint != "" {
		resp, err := client2.MakeRequest(client, config, "GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		var res client2.PagesResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		pages = append(pages, res.Results...)

		endpoint = res.Links.Next
	}
	return pages, error(nil)
}

func listHandler(client *http.Client, config *config2.Config, w http.ResponseWriter, r *http.Request) {
	pages, err := getAllPages(client, config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Notes List</h1><a href=\"/create\">Create New Note</a><br><ul>")
	for _, p := range pages {
		fmt.Fprintf(w, "<li>%s <a href=\"/read?id=%s\">View</a> <a href=\"/edit?id=%s\">Edit</a> <form action=\"/delete\" method=\"post\" style=\"display:inline;\"><input type=\"hidden\" name=\"id\" value=\"%s\"/><button type=\"submit\">Delete</button></form></li>", p.Title, p.ID, p.ID, p.ID)
	}
	fmt.Fprint(w, "</ul>")
}

func createHandler(client *http.Client, config *config2.Config, w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<h1>Create Note</h1><form method="post">
            <label>Title: <input name="title"/></label><br>
            <label>Content: <textarea name="content"></textarea></label><br>
            <button type="submit">Create</button>
        </form>`)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		title := r.FormValue("title")
		content := r.FormValue("content")

		reqBody := client2.CreatePageRequest{
			SpaceID:  config.SpaceID,
			Status:   "current",
			Title:    title,
			ParentID: config.ParentPageID,
			Body: struct {
				Storage struct {
					Value          string `json:"value"`
					Representation string `json:"representation"`
				} `json:"storage"`
			}{
				Storage: struct {
					Value          string `json:"value"`
					Representation string `json:"representation"`
				}{
					Value:          content,
					Representation: "storage",
				},
			},
		}

		resp, err := client2.MakeRequest(client, config, "POST", "/pages", reqBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Body.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func readHandler(client *http.Client, config *config2.Config, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	resp, err := client2.MakeRequest(client, config, "GET", fmt.Sprintf("/pages/%s?body-format=storage", id), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var page client2.Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		resp.Body.Close()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Body.Close()

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div><a href=\"/\">Back</a>", page.Title, page.Body.Storage.Value)
}

func editHandler(client *http.Client, config *config2.Config, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	resp, err := client2.MakeRequest(client, config, "GET", fmt.Sprintf("/pages/%s?body-format=storage", id), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var page client2.Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		resp.Body.Close()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Body.Close()

	if r.Method == "GET" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<h1>Edit Note</h1><form method="post">
            <label>Title: <input name="title" value="%s"/></label><br>
            <label>Content: <textarea name="content">%s</textarea></label><br>
            <button type="submit">Update</button>
        </form>`, page.Title, page.Body.Storage.Value)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		title := r.FormValue("title")
		content := r.FormValue("content")

		reqBody := client2.UpdatePageRequest{
			ID:     id,
			Title:  title,
			Status: "current",
			Body: struct {
				Storage struct {
					Value          string `json:"value"`
					Representation string `json:"representation"`
				} `json:"storage"`
			}{
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
				Number:  page.Version.Number + 1,
				Message: "Updated via Web App",
			},
		}

		resp, err = client2.MakeRequest(client, config, "PUT", fmt.Sprintf("/pages/%s", id), reqBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Body.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func deleteHandler(client *http.Client, config *config2.Config, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	resp, err := client2.MakeRequest(client, config, "DELETE", fmt.Sprintf("/pages/%s", id), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Body.Close()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
