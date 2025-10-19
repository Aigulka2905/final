package db

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// Task представляет структуру задачи
type Task struct {
	ID      string `db:"id"`
	Date    string `db:"date"`
	Title   string `db:"title"`
	Comment string `db:"comment"`
	Repeat  string `db:"repeat"`
}

// NextDateFunc определяет сигнатуру функции NextDate
type NextDateFunc func(time.Time, string, string) (string, error)

// GetTask возвращает задачу по ID
func GetTask(db *sqlx.DB, id string) (Task, error) {
	var task Task
	err := db.Get(&task, `SELECT id, date, title, COALESCE(comment, '') AS comment, COALESCE(repeat, '') AS repeat FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return task, err
	}
	return task, nil
}

// GetAllTasks возвращает все задачи из базы данных
func GetAllTasks(db *sqlx.DB) ([]Task, error) {
	var tasks []Task
	err := db.Select(&tasks, `
        SELECT id, date, title, COALESCE(comment, '') AS comment, COALESCE(repeat, '') AS repeat 
        FROM scheduler 
        ORDER BY date
    `)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// AddTask добавляет новую задачу и возвращает её ID
func AddTask(db *sqlx.DB, task Task) (string, error) {
	var id string
	err := db.QueryRowx(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?) RETURNING id`,
		task.Date, task.Title, task.Comment, task.Repeat,
	).Scan(&id)
	return id, err
}

// UpdateTask обновляет существующую задачу
func UpdateTask(db *sqlx.DB, task Task) error {
	result, err := db.NamedExec(
		`UPDATE scheduler SET date=:date, title=:title, comment=:comment, repeat=:repeat WHERE id=:id`,
		task,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteTask удаляет задачу по ID
func DeleteTask(db *sqlx.DB, id string) error {
	result, err := db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DoneTask обрабатывает выполнение задачи (обновляет дату или удаляет)
func DoneTask(db *sqlx.DB, id string, nextDate NextDateFunc) error {
	var task Task
	err := db.Get(&task, `SELECT id, date, repeat FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return err
	}

	if task.Repeat == "" {
		result, err := db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			return sql.ErrNoRows
		}
		return nil
	}

	now := time.Now()
	next, err := nextDate(now, task.Date, task.Repeat) // Используем nextDate как параметр
	if err != nil {
		return err
	}

	result, err := db.NamedExec(
		`UPDATE scheduler SET date=:date WHERE id=:id`,
		map[string]interface{}{"date": next, "id": id},
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
