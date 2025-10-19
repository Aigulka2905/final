package api

import (
	"encoding/json"
	"net/http"

	"final/pkg/db"
)

// tasksHandler обрабатывает GET /api/tasks, возвращая список всех задач
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем список задач через функцию из пакета db
	tasks, err := db.GetAllTasks(db.DB)
	if err != nil {
		writeError(w, "failed to fetch tasks", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок и отправляем ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string][]db.Task{"tasks": tasks}); err != nil {
		writeError(w, "failed to encode response", http.StatusInternalServerError)
	}
}
