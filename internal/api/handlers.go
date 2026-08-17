package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/mitja6889/HTTPUtils/internal/models"
	"github.com/mitja6889/HTTPUtils/internal/validate"
)

func (a *API) handleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	today, err := clientToday(r)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	printhttpWriteJSON(w, http.StatusOK, a.store.Overview(today))
}

func (a *API) handlePlans(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		printhttpWriteJSON(w, http.StatusOK, a.store.ListPlans())
	case http.MethodPost:
		var input createPlanInput
		if err := decodeJSON(w, r, &input); err != nil {
			writeBadRequest(w, "invalid JSON body")
			return
		}

		plan, err := a.createPlan(input)
		if err != nil {
			writeBadRequest(w, err.Error())
			return
		}

		printhttpWriteJSON(w, http.StatusCreated, plan)
	default:
		writeMethodNotAllowed(w)
	}
}

func (a *API) handlePlanByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/plans/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		a.handlePlans(w, r)
		return
	}
	if strings.Contains(id, "/") {
		writeBadRequest(w, "invalid plan id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		plan, err := a.store.GetPlan(id)
		if err != nil {
			a.writeStoreError(w, err)
			return
		}
		printhttpWriteJSON(w, http.StatusOK, plan)
	case http.MethodPut, http.MethodPatch:
		var input updatePlanInput
		if err := decodeJSON(w, r, &input); err != nil {
			writeBadRequest(w, "invalid JSON body")
			return
		}

		plan, err := a.store.UpdatePlan(id, func(p *models.Plan) error {
			return applyPlanUpdate(p, input)
		})
		if err != nil {
			a.writeStoreError(w, err)
			return
		}

		printhttpWriteJSON(w, http.StatusOK, plan)
	case http.MethodDelete:
		if err := a.store.DeletePlan(id); err != nil {
			a.writeStoreError(w, err)
			return
		}
		writeNoContent(w)
	default:
		writeMethodNotAllowed(w)
	}
}

func (a *API) handleGoals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		printhttpWriteJSON(w, http.StatusOK, a.store.ListGoals())
	case http.MethodPost:
		var input createGoalInput
		if err := decodeJSON(w, r, &input); err != nil {
			writeBadRequest(w, "invalid JSON body")
			return
		}

		goal, err := a.createGoal(input)
		if err != nil {
			writeBadRequest(w, err.Error())
			return
		}

		printhttpWriteJSON(w, http.StatusCreated, goal)
	default:
		writeMethodNotAllowed(w)
	}
}

func (a *API) handleGoalByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/goals/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		writeBadRequest(w, "invalid goal id")
		return
	}
	if strings.Contains(id, "/") {
		writeBadRequest(w, "invalid goal id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		goal, err := a.store.GetGoal(id)
		if err != nil {
			a.writeStoreError(w, err)
			return
		}
		printhttpWriteJSON(w, http.StatusOK, goal)
	case http.MethodPut, http.MethodPatch:
		var input updateGoalInput
		if err := decodeJSON(w, r, &input); err != nil {
			writeBadRequest(w, "invalid JSON body")
			return
		}

		goal, err := a.store.UpdateGoal(id, func(g *models.Goal) error {
			return applyGoalUpdate(g, input)
		})
		if err != nil {
			a.writeStoreError(w, err)
			return
		}

		printhttpWriteJSON(w, http.StatusOK, goal)
	case http.MethodDelete:
		if err := a.store.DeleteGoal(id); err != nil {
			a.writeStoreError(w, err)
			return
		}
		writeNoContent(w)
	default:
		writeMethodNotAllowed(w)
	}
}

func (a *API) handleHabits(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		today, err := clientToday(r)
		if err != nil {
			writeBadRequest(w, err.Error())
			return
		}
		printhttpWriteJSON(w, http.StatusOK, a.store.ListHabits(today))
	case http.MethodPost:
		var input createHabitInput
		if err := decodeJSON(w, r, &input); err != nil {
			writeBadRequest(w, "invalid JSON body")
			return
		}

		habit, err := a.createHabit(input)
		if err != nil {
			writeBadRequest(w, err.Error())
			return
		}

		printhttpWriteJSON(w, http.StatusCreated, habit)
	default:
		writeMethodNotAllowed(w)
	}
}

func (a *API) handleHabitByID(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/habits/")
	rest = strings.TrimSuffix(rest, "/")
	if rest == "" {
		writeBadRequest(w, "invalid habit id")
		return
	}

	parts := strings.Split(rest, "/")
	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}
	if len(parts) > 2 {
		writeBadRequest(w, "invalid habit path")
		return
	}

	switch r.Method {
	case http.MethodGet:
		habit, err := a.store.GetHabit(id)
		if err != nil {
			a.writeStoreError(w, err)
			return
		}
		today, err := clientToday(r)
		if err != nil {
			writeBadRequest(w, err.Error())
			return
		}
		habit.Streak = validate.EffectiveHabitStreak(habit, today)
		printhttpWriteJSON(w, http.StatusOK, habit)
	case http.MethodPut, http.MethodPatch:
		if action != "" {
			writeBadRequest(w, "unsupported action")
			return
		}
		var input updateHabitInput
		if err := decodeJSON(w, r, &input); err != nil {
			writeBadRequest(w, "invalid JSON body")
			return
		}

		habit, err := a.store.UpdateHabit(id, func(h *models.Habit) error {
			return applyHabitUpdate(h, input)
		})
		if err != nil {
			a.writeStoreError(w, err)
			return
		}
		printhttpWriteJSON(w, http.StatusOK, habit)
	case http.MethodPost:
		switch action {
		case "done":
			var input habitActionInput
			if r.ContentLength > 0 {
				if err := decodeJSON(w, r, &input); err != nil {
					writeBadRequest(w, "invalid JSON body")
					return
				}
			}

			today, err := clientTodayFrom(input.Date)
			if err != nil {
				writeBadRequest(w, err.Error())
				return
			}

			habit, err := a.store.UpdateHabit(id, func(h *models.Habit) error {
				if h.LastDone == today {
					return nil
				}

				yesterday := validate.AddDays(today, -1)
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

			printhttpWriteJSON(w, http.StatusOK, habit)
		case "undo":
			var input habitActionInput
			if r.ContentLength > 0 {
				if err := decodeJSON(w, r, &input); err != nil {
					writeBadRequest(w, "invalid JSON body")
					return
				}
			}

			today, err := clientTodayFrom(input.Date)
			if err != nil {
				writeBadRequest(w, err.Error())
				return
			}

			habit, err := a.store.UpdateHabit(id, func(h *models.Habit) error {
				if h.LastDone != today {
					return nil
				}

				h.LastDone = ""
				if h.Streak > 1 {
					h.Streak--
				} else {
					h.Streak = 0
				}
				return nil
			})
			if err != nil {
				a.writeStoreError(w, err)
				return
			}

			printhttpWriteJSON(w, http.StatusOK, habit)
		default:
			writeBadRequest(w, "unsupported action")
		}
	case http.MethodDelete:
		if action != "" {
			writeBadRequest(w, "unsupported action")
			return
		}
		if err := a.store.DeleteHabit(id); err != nil {
			a.writeStoreError(w, err)
			return
		}
		writeNoContent(w)
	default:
		writeMethodNotAllowed(w)
	}
}

func newID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return models.NowISO()
	}
	return hex.EncodeToString(buf)
}

type createPlanInput struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Category    models.Category   `json:"category"`
	Priority    models.Priority   `json:"priority"`
	Status      models.PlanStatus `json:"status"`
	DueDate     string            `json:"dueDate"`
	Tasks       []models.Task     `json:"tasks"`
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
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Kind         models.GoalKind   `json:"kind"`
	TargetDate   string            `json:"targetDate"`
	TargetAmount float64           `json:"targetAmount"`
	Progress     int               `json:"progress"`
	Status       models.GoalStatus `json:"status"`
}

type updateGoalInput struct {
	Title        *string            `json:"title"`
	Description  *string            `json:"description"`
	Kind         *models.GoalKind   `json:"kind"`
	TargetDate   *string            `json:"targetDate"`
	TargetAmount *float64           `json:"targetAmount"`
	Progress     *int               `json:"progress"`
	Status       *models.GoalStatus `json:"status"`
}

type createHabitInput struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type updateHabitInput struct {
	Name *string `json:"name"`
	Icon *string `json:"icon"`
}

type habitActionInput struct {
	Date string `json:"date"`
}

func (a *API) createPlan(input createPlanInput) (models.Plan, error) {
	title, err := validate.Title(input.Title)
	if err != nil {
		return models.Plan{}, err
	}

	category := input.Category
	if category == "" {
		category = models.CategoryPersonal
	}
	if err := validate.Category(category); err != nil {
		return models.Plan{}, err
	}

	priority := input.Priority
	if priority == "" {
		priority = models.PriorityMedium
	}
	if err := validate.Priority(priority); err != nil {
		return models.Plan{}, err
	}

	status := input.Status
	if status == "" {
		status = models.PlanStatusTodo
	}
	if err := validate.PlanStatus(status); err != nil {
		return models.Plan{}, err
	}

	if err := validate.Date(input.DueDate); err != nil {
		return models.Plan{}, err
	}

	now := models.NowISO()
	plan := models.Plan{
		ID:          newID(),
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Category:    category,
		Priority:    priority,
		Status:      status,
		DueDate:     input.DueDate,
		Tasks:       normalizeTasks(input.Tasks, now),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return a.store.CreatePlan(plan)
}

func applyPlanUpdate(p *models.Plan, input updatePlanInput) error {
	if input.Title != nil {
		title, err := validate.Title(*input.Title)
		if err != nil {
			return err
		}
		p.Title = title
	}
	if input.Description != nil {
		p.Description = strings.TrimSpace(*input.Description)
	}
	if input.Category != nil {
		if err := validate.Category(*input.Category); err != nil {
			return err
		}
		p.Category = *input.Category
	}
	if input.Priority != nil {
		if err := validate.Priority(*input.Priority); err != nil {
			return err
		}
		p.Priority = *input.Priority
	}
	if input.Status != nil {
		if err := validate.PlanStatus(*input.Status); err != nil {
			return err
		}
		p.Status = *input.Status
	}
	if input.DueDate != nil {
		if err := validate.Date(*input.DueDate); err != nil {
			return err
		}
		p.DueDate = *input.DueDate
	}
	if input.Tasks != nil {
		p.Tasks = normalizeTasks(*input.Tasks, models.NowISO())
	}

	return nil
}

func normalizeTasks(tasks []models.Task, now string) []models.Task {
	if tasks == nil {
		return []models.Task{}
	}

	out := make([]models.Task, 0, len(tasks))
	for _, task := range tasks {
		title := strings.TrimSpace(task.Title)
		if title == "" {
			continue
		}
		if task.ID == "" {
			task.ID = newID()
		}
		if task.CreatedAt == "" {
			task.CreatedAt = now
		}
		task.Title = title
		out = append(out, task)
	}
	return out
}

func (a *API) createGoal(input createGoalInput) (models.Goal, error) {
	title, err := validate.Title(input.Title)
	if err != nil {
		return models.Goal{}, err
	}

	status := input.Status
	if status == "" {
		status = models.GoalStatusActive
	}
	if err := validate.GoalStatus(status); err != nil {
		return models.Goal{}, err
	}

	if err := validate.Date(input.TargetDate); err != nil {
		return models.Goal{}, err
	}

	kind := input.Kind
	if kind == "" {
		if input.TargetAmount > 0 {
			kind = models.GoalKindFinancial
		} else {
			kind = models.GoalKindManual
		}
	}
	if err := validate.GoalKind(kind); err != nil {
		return models.Goal{}, err
	}

	now := models.NowISO()
	progress := validate.ClampProgress(input.Progress)
	if kind == models.GoalKindFinancial {
		progress = 0
	}

	goal := models.Goal{
		ID:           newID(),
		Title:        title,
		Description:  strings.TrimSpace(input.Description),
		Kind:         kind,
		TargetDate:   input.TargetDate,
		TargetAmount: input.TargetAmount,
		Progress:     progress,
		Status:       status,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	created, err := a.store.CreateGoal(goal)
	if err != nil {
		return models.Goal{}, err
	}
	return a.store.EnrichGoal(created), nil
}

func applyGoalUpdate(g *models.Goal, input updateGoalInput) error {
	if input.Title != nil {
		title, err := validate.Title(*input.Title)
		if err != nil {
			return err
		}
		g.Title = title
	}
	if input.Description != nil {
		g.Description = strings.TrimSpace(*input.Description)
	}
	if input.TargetDate != nil {
		if err := validate.Date(*input.TargetDate); err != nil {
			return err
		}
		g.TargetDate = *input.TargetDate
	}
	if input.Kind != nil {
		if err := validate.GoalKind(*input.Kind); err != nil {
			return err
		}
		g.Kind = *input.Kind
	}
	if input.TargetAmount != nil {
		g.TargetAmount = *input.TargetAmount
		if g.TargetAmount > 0 {
			g.Kind = models.GoalKindFinancial
		}
	}
	if input.Progress != nil && g.Kind != models.GoalKindFinancial {
		g.Progress = validate.ClampProgress(*input.Progress)
	}
	if input.Status != nil {
		if err := validate.GoalStatus(*input.Status); err != nil {
			return err
		}
		g.Status = *input.Status
	}

	return nil
}

func (a *API) createHabit(input createHabitInput) (models.Habit, error) {
	name, err := validate.Name(input.Name)
	if err != nil {
		return models.Habit{}, err
	}

	icon, err := validate.Icon(input.Icon)
	if err != nil {
		return models.Habit{}, err
	}

	habit := models.Habit{
		ID:        newID(),
		Name:      name,
		Icon:      icon,
		Streak:    0,
		CreatedAt: models.NowISO(),
	}

	return a.store.CreateHabit(habit)
}

func applyHabitUpdate(h *models.Habit, input updateHabitInput) error {
	if input.Name != nil {
		name, err := validate.Name(*input.Name)
		if err != nil {
			return err
		}
		h.Name = name
	}
	if input.Icon != nil {
		icon, err := validate.Icon(*input.Icon)
		if err != nil {
			return err
		}
		h.Icon = icon
	}
	return nil
}
