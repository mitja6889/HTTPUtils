package PrintHTTP

import (
	"encoding/json"
	"net/http"

	"github.com/getsentry/sentry-go"
)

type RawMap map[string]any

func WriteInternalError(w http.ResponseWriter, err error) {

	WriteJSON(w, http.StatusInternalServerError, RawMap{"error": err.Error()})
}

func WriteErrorMessage(w http.ResponseWriter, statusCode int, message string) {

	WriteJSON(w, statusCode, RawMap{"error": message})
}

func WriteJSON(w http.ResponseWriter, statusCode int, data any) {

	if statusCode == 0 {

		statusCode = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {

		sentry.CaptureException(err)
	}
}

func WriteStatus(w http.ResponseWriter, statusCode int) {

	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
}

func WriteString(w http.ResponseWriter, statusCode int, data string) {

	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(data)); err != nil {

		sentry.CaptureException(err)
	}
}
