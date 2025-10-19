package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"final/pkg/db"
)

type DeleteRequest struct {
	ID string `json:"id"`
}

const dateFormat = "20060102"

// taskHandler маршрутизирует запросы к /api/task
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPost:
		postTaskHandler(w, r)
	case http.MethodPut:
		putTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// getTaskHandler обрабатывает GET /api/task?id=<id>
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		fmt.Println("GET task: missing id parameter")
		writeError(w, "id is required", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(db.DB, id)
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
	}
}

// postTaskHandler обрабатывает POST /api/task
func postTaskHandler(w http.ResponseWriter, r *http.Request) {
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
	var task db.Task // Используем db.Task
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

	id, err := db.AddTask(db.DB, task) // Теперь task типа db.Task
	if err != nil {
		fmt.Printf("POST DB error: %v\n", err)
		writeError(w, "failed to save task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"id": id}
	fmt.Printf("POST response: %+v\n", response)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("POST encode error: %v\n", err)
		writeError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// putTaskHandler обрабатывает PUT /api/task
func putTaskHandler(w http.ResponseWriter, r *http.Request) {
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
	var task db.Task // Используем db.Task
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

	err = db.UpdateTask(db.DB, task) // Теперь task типа db.Task
	if err != nil {
		fmt.Printf("PUT DB error: %v\n", err)
		writeError(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{}); err != nil {
		fmt.Printf("PUT encode error: %v\n", err)
		writeError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// deleteTaskHandler обрабатывает DELETE /api/task
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	err = db.DeleteTask(db.DB, id)
	if err != nil {
		fmt.Printf("DELETE DB error: %v\n", err)
		writeError(w, "failed to delete task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{}); err != nil {
		fmt.Printf("DELETE encode error: %v\n", err)
		writeError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// doneHandler обрабатывает выполнение задачи (обновление даты или удаление)
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

	err = db.DoneTask(db.DB, id, NextDate)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("DONE task not found: id=%s\n", id)
			writeError(w, "task not found", http.StatusNotFound)
			return
		}
		fmt.Printf("DONE DB error: %v\n", err)
		writeError(w, "failed to process task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{})
}
