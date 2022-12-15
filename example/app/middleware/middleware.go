package middleware

import (
	"net/http"
	"time"
)

func GetTime(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("Current time is " + time.Now().Format("15:04:05"))); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
