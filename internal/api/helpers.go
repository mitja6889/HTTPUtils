package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	printhttp "github.com/mitja6889/HTTPUtils/PrintHTTP"
	"github.com/mitja6889/HTTPUtils/internal/store"
	"github.com/mitja6889/HTTPUtils/internal/validate"
)

const maxBodyBytes = 1 << 20

type API struct {
	store *store.Store
}

func New(s *store.Store) *API {
	return &API{store: s}
}

func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", a.handleHealth)
	mux.HandleFunc("/api/overview", a.handleOverview)
	mux.HandleFunc("/api/plans", a.handlePlans)
	mux.HandleFunc("/api/plans/", a.handlePlanByID)
	mux.HandleFunc("/api/goals", a.handleGoals)
	mux.HandleFunc("/api/goals/", a.handleGoalByID)
	mux.HandleFunc("/api/habits", a.handleHabits)
	mux.HandleFunc("/api/habits/", a.handleHabitByID)
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	printhttp.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func (a *API) writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		printhttp.WriteErrorMessage(w, http.StatusNotFound, "not found")
	case errors.Is(err, store.ErrAlreadyExists):
		printhttp.WriteErrorMessage(w, http.StatusConflict, "already exists")
	default:
		printhttp.WriteErrorMessage(w, http.StatusInternalServerError, "internal server error")
	}
}

func clientToday(r *http.Request) (string, error) {
	return validate.ClientToday(r.URL.Query().Get("today"))
}

func clientTodayFrom(value string) (string, error) {
	return validate.ClientToday(value)
}
