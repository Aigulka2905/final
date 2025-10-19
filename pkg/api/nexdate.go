package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", nil
	}

	start, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %v", err)
	}

	parts := strings.SplitN(repeat, " ", 2)
	rule := parts[0]

	switch {
	case rule == "y":

		start = start.AddDate(1, 0, 0)

		for !start.After(now) {
			start = start.AddDate(1, 0, 0)
		}
	case strings.HasPrefix(rule, "d"):
		if len(parts) < 2 {
			return "", fmt.Errorf("missing interval for 'd' rule")
		}
		interval, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || interval < 1 || interval > 400 {
			return "", fmt.Errorf("invalid interval: must be 1-400")
		}
		start = start.AddDate(0, 0, interval)
		for !start.After(now) {
			start = start.AddDate(0, 0, interval)
		}
	default:
		return "", fmt.Errorf("unsupported repeat rule: %s", rule)
	}

	return start.Format(dateFormat), nil
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	now := time.Now()
	if nowStr != "" {
		parsedNow, err := time.Parse(dateFormat, nowStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid now date format: %v", err), http.StatusBadRequest)
			return
		}
		now = parsedNow
	}

	if dateStr == "" {
		http.Error(w, "date parameter is required", http.StatusBadRequest)
		return
	}

	next, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, next)
}
