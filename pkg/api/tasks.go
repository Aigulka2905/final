package api

import (
	"encoding/json"
	"final/pkg/db"
	"fmt"
	"net/http"
)

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tasks []Task
	err := db.DB.Select(&tasks, `SELECT id, date, title, COALESCE(comment, '') AS comment, COALESCE(repeat, '') AS repeat FROM scheduler ORDER BY date`)
	if err != nil {
		fmt.Printf("GET tasks DB error: %v\n", err)
		writeError(w, "failed to fetch tasks", http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []Task{}
	}

	fmt.Printf("GET tasks response: %+v\n", tasks)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string][]Task{"tasks": tasks})
}
