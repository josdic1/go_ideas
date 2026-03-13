package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Idea struct {
	ID   int    `json:"id"`
	Idea string `json:"idea"`
}

var db *sql.DB

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./ideas.db")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS ideas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		idea TEXT NOT NULL
	)
	`)
	if err != nil {
		panic(err)
	}

http.Handle("/", logRequest(http.FileServer(http.Dir("./static"))))
http.Handle("/ideas", logRequest(http.HandlerFunc(ideasHandler)))
http.Handle("/ideas/", logRequest(http.HandlerFunc(ideaByIDHandler)))

	fmt.Println("Server running at http://localhost:8001")
	http.ListenAndServe(":8001", nil)
}

func ideasHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		rows, err := db.Query("SELECT id, idea FROM ideas ORDER BY id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		ideas := []Idea{}
		for rows.Next() {
			var idea Idea
			if err := rows.Scan(&idea.ID, &idea.Idea); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ideas = append(ideas, idea)
		}

		json.NewEncoder(w).Encode(ideas)
		return
	}

	if r.Method == "POST" {
		var newIdea Idea

		if err := json.NewDecoder(r.Body).Decode(&newIdea); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := db.Exec("INSERT INTO ideas (idea) VALUES (?)", newIdea.Idea)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newIdea.ID = int(id)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newIdea)
		return
	}



	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"error":"method not allowed"}`))
}

func ideaByIDHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/ideas/"):]

	if r.Method == "GET" {

		row := db.QueryRow("SELECT id, idea FROM ideas WHERE id = ?", id)

		var idea Idea

		err := row.Scan(&idea.ID, &idea.Idea)
		if err != nil {
			http.Error(w, "idea not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(idea)
		return
	}

	if r.Method == "DELETE" {

		result, err := db.Exec("DELETE FROM ideas WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			http.Error(w, "idea not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"status": "deleted",
		})
		return
	}

	if r.Method == "PATCH" {

	var updated Idea

	err := json.NewDecoder(r.Body).Decode(&updated)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec(
		"UPDATE ideas SET idea = ? WHERE id = ?",
		updated.Idea,
		id,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "idea not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "updated",
	})
	return
}

	w.WriteHeader(http.StatusMethodNotAllowed)
}