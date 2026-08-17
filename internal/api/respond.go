package api

import (
	"net/http"

	printhttp "github.com/mitja6889/HTTPUtils/PrintHTTP"
)

func printhttpWriteJSON(w http.ResponseWriter, status int, data any) {
	printhttp.WriteJSON(w, status, data)
}

func writeBadRequest(w http.ResponseWriter, message string) {
	printhttp.WriteErrorMessage(w, http.StatusBadRequest, message)
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	printhttp.WriteErrorMessage(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeNoContent(w http.ResponseWriter) {
	printhttp.WriteStatus(w, http.StatusNoContent)
}

func WriteNotFound(w http.ResponseWriter) {
	printhttp.WriteErrorMessage(w, http.StatusNotFound, "not found")
}
