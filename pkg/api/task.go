package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"final/pkg/db"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Task struct {
	ID      string `json:"id" db:"id"`
	Date    string `json:"date" db:"date"`
	Title   string `json:"title" db:"title"`
	Comment string `json:"comment" db:"comment"`
	Repeat  string `json:"repeat" db:"repeat"`
}

type DeleteRequest struct {
	ID string `json:"id"`
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		id := r.URL.Query().Get("id")
		if id == "" {
			fmt.Println("GET task: missing id parameter")
			writeError(w, "id is required", http.StatusBadRequest)
			return
		}

		var task Task
		err := db.DB.Get(&task, `SELECT id, date, title, COALESCE(comment, '') AS comment, COALESCE(repeat, '') AS repeat FROM scheduler WHERE id = ?`, id)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Printf("GET task not found: id=%s\n", id)
				writeError(w, "task not found", http.StatusNotFound)
				return
			}
			fmt.Printf("GET task DB error: %v\n", err)
			writeError(w, "failed to fetch task", http.StatusInternalServerError)
			return
		}

		fmt.Printf("GET task response: %+v\n", task)

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(task); err != nil {
			fmt.Printf("GET task encode error: %v\n", err)
			writeError(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("POST read body error: %v\n", err)
			writeError(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		fmt.Printf("POST request body: %s\n", string(body))

		if len(body) == 0 {
			fmt.Println("POST empty request body")
			writeError(w, "request body is empty", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			fmt.Printf("POST JSON decode error: %v\n", err)
			writeError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		fmt.Printf("POST Task: %+v\n", task)

		if task.Title == "" {
			fmt.Println("POST missing title")
			writeError(w, "title is required", http.StatusBadRequest)
			return
		}

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		if task.Date == "" {
			task.Date = today.Format(dateFormat)
		}

		date, err := time.Parse(dateFormat, task.Date)
		if err != nil {
			fmt.Printf("POST date parse error: %v\n", err)
			writeError(w, "invalid date format", http.StatusBadRequest)
			return
		}

		if date.Before(today) {
			task.Date = today.Format(dateFormat)
		}

		if task.Repeat != "" && date.Before(today) {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				fmt.Printf("POST NextDate error: %v\n", err)
				writeError(w, err.Error(), http.StatusBadRequest)
				return
			}
			task.Date = next
		}

		result, err := db.DB.NamedExec(
			`INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)`,
			task,
		)
		if err != nil {
			fmt.Printf("POST DB error: %v\n", err)
			writeError(w, "failed to save task", http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			fmt.Printf("POST LastInsertId error: %v\n", err)
			writeError(w, "failed to get task ID", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		response := map[string]string{"id": fmt.Sprintf("%d", id)}
		fmt.Printf("POST response: %+v\n", response)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			fmt.Printf("POST encode error: %v\n", err)
			writeError(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("PUT read body error: %v\n", err)
			writeError(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		fmt.Printf("PUT request body: %s\n", string(body))

		if len(body) == 0 {
			fmt.Println("PUT empty request body")
			writeError(w, "request body is empty", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			fmt.Printf("PUT JSON decode error: %v\n", err)
			writeError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		fmt.Printf("PUT Task: %+v\n", task)

		if task.ID == "" {
			fmt.Println("PUT missing id")
			writeError(w, "id is required", http.StatusBadRequest)
			return
		}

		if task.Title == "" {
			fmt.Println("PUT missing title")
			writeError(w, "title is required", http.StatusBadRequest)
			return
		}

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		if task.Date == "" {
			task.Date = today.Format(dateFormat)
		}

		date, err := time.Parse(dateFormat, task.Date)
		if err != nil {
			fmt.Printf("PUT date parse error: %v\n", err)
			writeError(w, "invalid date format", http.StatusBadRequest)
			return
		}

		if date.Before(today) {
			task.Date = today.Format(dateFormat)
		}

		if task.Repeat != "" && date.Before(today) {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				fmt.Printf("PUT NextDate error: %v\n", err)
				writeError(w, err.Error(), http.StatusBadRequest)
				return
			}
			task.Date = next
		}

		result, err := db.DB.NamedExec(
			`UPDATE scheduler SET date=:date, title=:title, comment=:comment, repeat=:repeat WHERE id=:id`,
			task,
		)
		if err != nil {
			fmt.Printf("PUT DB error: %v\n", err)
			writeError(w, "failed to update task", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			fmt.Printf("PUT rowsAffected error: %v\n", err)
			writeError(w, "failed to check update result", http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			fmt.Printf("PUT task not found: id=%s\n", task.ID)
			writeError(w, "task not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{}); err != nil {
			fmt.Printf("PUT encode error: %v\n", err)
			writeError(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	case http.MethodDelete:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("DELETE read body error: %v\n", err)
			writeError(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		fmt.Printf("DELETE request body: %s\n", string(body))

		var id string
		if len(body) > 0 {
			var req DeleteRequest
			if err := json.Unmarshal(body, &req); err != nil {
				fmt.Printf("DELETE JSON unmarshal error: %v\n", err)
				writeError(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			id = req.ID
		} else {
			id = r.URL.Query().Get("id")
			if id == "" {
				fmt.Println("DELETE missing id in body or query")
				writeError(w, "id is required", http.StatusBadRequest)
				return
			}
		}

		fmt.Printf("Delete ID: %s\n", id)

		if id == "" {
			fmt.Println("DELETE missing id")
			writeError(w, "id is required", http.StatusBadRequest)
			return
		}

		result, err := db.DB.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
		if err != nil {
			fmt.Printf("DELETE DB error: %v\n", err)
			writeError(w, "failed to delete task", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			fmt.Printf("DELETE rowsAffected error: %v\n", err)
			writeError(w, "failed to check delete result", http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			fmt.Printf("DELETE task not found: id=%s\n", id)
			writeError(w, "task not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{}); err != nil {
			fmt.Printf("DELETE encode error: %v\n", err)
			writeError(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

	default:
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func doneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("DONE read body error: %v\n", err)
		writeError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	fmt.Printf("DONE request body: %s\n", string(body))

	var id string
	if len(body) > 0 {
		var req DeleteRequest
		if err := json.Unmarshal(body, &req); err != nil {
			fmt.Printf("DONE JSON unmarshal error: %v\n", err)
			writeError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		id = req.ID
	} else {
		id = r.URL.Query().Get("id")
		if id == "" {
			fmt.Println("DONE missing id in body or query")
			writeError(w, "id is required", http.StatusBadRequest)
			return
		}
	}

	fmt.Printf("Done ID: %s\n", id)

	if id == "" {
		writeError(w, "id is required", http.StatusBadRequest)
		return
	}

	var task Task
	err = db.DB.Get(&task, `SELECT id, date, title, COALESCE(comment, '') AS comment, COALESCE(repeat, '') AS repeat FROM scheduler WHERE id = ?`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("DONE task not found: id=%s\n", id)
			writeError(w, "task not found", http.StatusNotFound)
			return
		}
		fmt.Printf("DONE DB error: %v\n", err)
		writeError(w, "failed to fetch task", http.StatusInternalServerError)
		return
	}

	if task.Repeat == "" {
		result, err := db.DB.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
		if err != nil {
			fmt.Printf("DONE delete DB error: %v\n", err)
			writeError(w, "failed to delete task", http.StatusInternalServerError)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			fmt.Printf("DONE delete task not found: id=%s\n", id)
			writeError(w, "task not found", http.StatusNotFound)
			return
		}
	} else {
		now := time.Now()
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			fmt.Printf("DONE NextDate error: %v\n", err)
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := db.DB.NamedExec(
			`UPDATE scheduler SET date=:date WHERE id=:id`,
			map[string]interface{}{"date": next, "id": id},
		)
		if err != nil {
			fmt.Printf("DONE update DB error: %v\n", err)
			writeError(w, "failed to update task", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			fmt.Printf("DONE update task not found: id=%s\n", id)
			writeError(w, "task not found", http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{})
}
