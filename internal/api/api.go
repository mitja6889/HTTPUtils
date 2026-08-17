package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	printhttp "github.com/mitja6889/HTTPUtils/PrintHTTP"
	"github.com/mitja6889/HTTPUtils/internal/models"
	"github.com/mitja6889/HTTPUtils/internal/store"
)

type API struct {
	store *store.Store
}

func New(s *store.Store) *API {
	return &API{store: s}
}

func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/overview", a.handleOverview)
	mux.HandleFunc("/api/plans", a.handlePlans)
	mux.HandleFunc("/api/plans/", a.handlePlanByID)
	mux.HandleFunc("/api/goals", a.handleGoals)
	mux.HandleFunc("/api/goals/", a.handleGoalByID)
	mux.HandleFunc("/api/habits", a.handleHabits)
	mux.HandleFunc("/api/habits/", a.handleHabitByID)
}

func (a *API) handleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		printhttp.WriteErrorMessage(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	printhttp.WriteJSON(w, http.StatusOK, a.store.Overview())
}

func (a *API) handlePlans(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		printhttp.WriteJSON(w, http.StatusOK, a.store.ListPlans())
	case http.MethodPost:
		var input createPlanInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			printhttp.WriteErrorMessage(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		plan, err := a.createPlan(input)
		if err != nil {
			printhttp.WriteErrorMessage(w, http.StatusBadRequest, err.Error())
			return
		}

		printhttp.WriteJSON(w, http.StatusCreated, plan)
	default:
		printhttp.WriteErrorMessage(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *API) handlePlanByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/plans/")
	if id == "" || strings.Contains(id, "/") {
		printhttp.WriteErrorMessage(w, http.StatusBadRequest, "invalid plan id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		plan, err := a.store.GetPlan(id)
		if err != nil {
			a.writeStoreError(w, err)
			return
		}
		printhttp.WriteJSON(w, http.StatusOK, plan)
	case http.MethodPut, http.MethodPatch:
		var input updatePlanInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			printhttp.WriteErrorMessage(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		plan, err := a.store.UpdatePlan(id, func(p *models.Plan) error {
			return applyPlanUpdate(p, input)
		})
		if err != nil {
			a.writeStoreError(w, err)
			return
		}

		printhttp.WriteJSON(w, http.StatusOK, plan)
	case http.MethodDelete:
		if err := a.store.DeletePlan(id); err != nil {
			a.writeStoreError(w, err)
			return
		}
		printhttp.WriteStatus(w, http.StatusNoContent)
	default:
		printhttp.WriteErrorMessage(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *API) handleGoals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		printhttp.WriteJSON(w, http.StatusOK, a.store.ListGoals())
	case http.MethodPost:
		var input createGoalInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			printhttp.WriteErrorMessage(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		goal, err := a.createGoal(input)
		if err != nil {
			printhttp.WriteErrorMessage(w, http.StatusBadRequest, err.Error())
			return
		}

		printhttp.WriteJSON(w, http.StatusCreated, goal)
	default:
		printhttp.WriteErrorMessage(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *API) handleGoalByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/goals/")
	if id == "" || strings.Contains(id, "/") {
		printhttp.WriteErrorMessage(w, http.StatusBadRequest, "invalid goal id")
		return
	}

	switch r.Method {
	case http.MethodPut, http.MethodPatch:
		var input updateGoalInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			printhttp.WriteErrorMessage(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		goal, err := a.store.UpdateGoal(id, func(g *models.Goal) error {
			return applyGoalUpdate(g, input)
		})
		if err != nil {
			a.writeStoreError(w, err)
			return
		}

		printhttp.WriteJSON(w, http.StatusOK, goal)
	case http.MethodDelete:
		if err := a.store.DeleteGoal(id); err != nil {
			a.writeStoreError(w, err)
			return
		}
		printhttp.WriteStatus(w, http.StatusNoContent)
	default:
		printhttp.WriteErrorMessage(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *API) handleHabits(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		printhttp.WriteJSON(w, http.StatusOK, a.store.ListHabits())
	case http.MethodPost:
		var input createHabitInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			printhttp.WriteErrorMessage(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		habit, err := a.createHabit(input)
		if err != nil {
			printhttp.WriteErrorMessage(w, http.StatusBadRequest, err.Error())
			return
		}

		printhttp.WriteJSON(w, http.StatusCreated, habit)
	default:
		printhttp.WriteErrorMessage(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *API) handleHabitByID(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/habits/")
	if rest == "" {
		printhttp.WriteErrorMessage(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	parts := strings.Split(rest, "/")
	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch r.Method {
	case http.MethodPost:
		if action == "done" {
			habit, err := a.store.UpdateHabit(id, func(h *models.Habit) error {
				today := models.TodayDate()
				if h.LastDone == today {
					return nil
				}

				yesterday := yesterdayDate()
				if h.LastDone == yesterday {
					h.Streak++
				} else {
					h.Streak = 1
				}

				h.LastDone = today
				return nil
			})
			if err != nil {
				a.writeStoreError(w, err)
				return
			}

			printhttp.WriteJSON(w, http.StatusOK, habit)
			return
		}

		printhttp.WriteErrorMessage(w, http.StatusBadRequest, "unsupported action")
	case http.MethodDelete:
		if err := a.store.DeleteHabit(id); err != nil {
			a.writeStoreError(w, err)
			return
		}
		printhttp.WriteStatus(w, http.StatusNoContent)
	default:
		printhttp.WriteErrorMessage(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *API) writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		printhttp.WriteErrorMessage(w, http.StatusNotFound, "not found")
		return
	}

	printhttp.WriteInternalError(w, err)
}

func newID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return models.NowISO()
	}

	return hex.EncodeToString(buf)
}

type createPlanInput struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Category    models.Category  `json:"category"`
	Priority    models.Priority  `json:"priority"`
	Status      models.PlanStatus `json:"status"`
	DueDate     string           `json:"dueDate"`
	Tasks       []models.Task    `json:"tasks"`
}

type updatePlanInput struct {
	Title       *string            `json:"title"`
	Description *string            `json:"description"`
	Category    *models.Category   `json:"category"`
	Priority    *models.Priority   `json:"priority"`
	Status      *models.PlanStatus `json:"status"`
	DueDate     *string            `json:"dueDate"`
	Tasks       *[]models.Task     `json:"tasks"`
}

type createGoalInput struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	TargetDate  string            `json:"targetDate"`
	Progress    int               `json:"progress"`
	Status      models.GoalStatus `json:"status"`
}

type updateGoalInput struct {
	Title       *string            `json:"title"`
	Description *string            `json:"description"`
	TargetDate  *string            `json:"targetDate"`
	Progress    *int               `json:"progress"`
	Status      *models.GoalStatus `json:"status"`
}

type createHabitInput struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

func (a *API) createPlan(input createPlanInput) (models.Plan, error) {
	if strings.TrimSpace(input.Title) == "" {
		return models.Plan{}, errors.New("title is required")
	}

	now := models.NowISO()
	plan := models.Plan{
		ID:          newID(),
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Category:    defaultCategory(input.Category),
		Priority:    defaultPriority(input.Priority),
		Status:      defaultPlanStatus(input.Status),
		DueDate:     input.DueDate,
		Tasks:       input.Tasks,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if plan.Tasks == nil {
		plan.Tasks = []models.Task{}
	}

	return a.store.CreatePlan(plan)
}

func applyPlanUpdate(p *models.Plan, input updatePlanInput) error {
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return errors.New("title cannot be empty")
		}
		p.Title = title
	}
	if input.Description != nil {
		p.Description = strings.TrimSpace(*input.Description)
	}
	if input.Category != nil {
		p.Category = defaultCategory(*input.Category)
	}
	if input.Priority != nil {
		p.Priority = defaultPriority(*input.Priority)
	}
	if input.Status != nil {
		p.Status = defaultPlanStatus(*input.Status)
	}
	if input.DueDate != nil {
		p.DueDate = *input.DueDate
	}
	if input.Tasks != nil {
		p.Tasks = *input.Tasks
	}

	return nil
}

func (a *API) createGoal(input createGoalInput) (models.Goal, error) {
	if strings.TrimSpace(input.Title) == "" {
		return models.Goal{}, errors.New("title is required")
	}

	now := models.NowISO()
	goal := models.Goal{
		ID:          newID(),
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		TargetDate:  input.TargetDate,
		Progress:    clampProgress(input.Progress),
		Status:      defaultGoalStatus(input.Status),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return a.store.CreateGoal(goal)
}

func applyGoalUpdate(g *models.Goal, input updateGoalInput) error {
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return errors.New("title cannot be empty")
		}
		g.Title = title
	}
	if input.Description != nil {
		g.Description = strings.TrimSpace(*input.Description)
	}
	if input.TargetDate != nil {
		g.TargetDate = *input.TargetDate
	}
	if input.Progress != nil {
		g.Progress = clampProgress(*input.Progress)
	}
	if input.Status != nil {
		g.Status = defaultGoalStatus(*input.Status)
	}

	return nil
}

func (a *API) createHabit(input createHabitInput) (models.Habit, error) {
	if strings.TrimSpace(input.Name) == "" {
		return models.Habit{}, errors.New("name is required")
	}

	icon := strings.TrimSpace(input.Icon)
	if icon == "" {
		icon = "✨"
	}

	habit := models.Habit{
		ID:        newID(),
		Name:      strings.TrimSpace(input.Name),
		Icon:      icon,
		Streak:    0,
		CreatedAt: models.NowISO(),
	}

	return a.store.CreateHabit(habit)
}

func defaultCategory(c models.Category) models.Category {
	switch c {
	case models.CategoryWork, models.CategoryHealth, models.CategoryPersonal, models.CategoryLearning, models.CategoryOther:
		return c
	default:
		return models.CategoryPersonal
	}
}

func defaultPriority(p models.Priority) models.Priority {
	switch p {
	case models.PriorityLow, models.PriorityMedium, models.PriorityHigh:
		return p
	default:
		return models.PriorityMedium
	}
}

func defaultPlanStatus(s models.PlanStatus) models.PlanStatus {
	switch s {
	case models.PlanStatusTodo, models.PlanStatusInProgress, models.PlanStatusDone:
		return s
	default:
		return models.PlanStatusTodo
	}
}

func defaultGoalStatus(s models.GoalStatus) models.GoalStatus {
	switch s {
	case models.GoalStatusActive, models.GoalStatusCompleted, models.GoalStatusPaused:
		return s
	default:
		return models.GoalStatusActive
	}
}

func clampProgress(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func yesterdayDate() string {
	return time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
}
