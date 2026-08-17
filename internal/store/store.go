package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mitja6889/HTTPUtils/internal/models"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Store struct {
	mu   sync.RWMutex
	path string
	data models.DataStore
}

func New(path string) (*Store, error) {
	s := &Store{
		path: path,
		data: models.DataStore{
			Plans:  []models.Plan{},
			Goals:  []models.Goal{},
			Habits: []models.Habit{},
		},
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	if _, err := os.Stat(path); err == nil {
		if err := s.load(); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("stat data file: %w", err)
	} else if err := s.save(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) load() error {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("read data file: %w", err)
	}

	if len(raw) == 0 {
		return nil
	}

	if err := json.Unmarshal(raw, &s.data); err != nil {
		return fmt.Errorf("parse data file: %w", err)
	}

	return nil
}

func (s *Store) save() error {
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

func (s *Store) ListPlans() []models.Plan {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]models.Plan, len(s.data.Plans))
	copy(out, s.data.Plans)
	return out
}

func (s *Store) GetPlan(id string) (models.Plan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.data.Plans {
		if p.ID == id {
			return p, nil
		}
	}

	return models.Plan{}, ErrNotFound
}

func (s *Store) CreatePlan(plan models.Plan) (models.Plan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.data.Plans {
		if p.ID == plan.ID {
			return models.Plan{}, ErrAlreadyExists
		}
	}

	if plan.Tasks == nil {
		plan.Tasks = []models.Task{}
	}

	s.data.Plans = append([]models.Plan{plan}, s.data.Plans...)

	if err := s.save(); err != nil {
		s.data.Plans = s.data.Plans[1:]
		return models.Plan{}, err
	}

	return plan, nil
}

func (s *Store) UpdatePlan(id string, update func(*models.Plan) error) (models.Plan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.data.Plans {
		if s.data.Plans[i].ID != id {
			continue
		}

		before := s.data.Plans[i]
		if err := update(&s.data.Plans[i]); err != nil {
			return models.Plan{}, err
		}

		s.data.Plans[i].UpdatedAt = models.NowISO()

		if err := s.save(); err != nil {
			s.data.Plans[i] = before
			return models.Plan{}, err
		}

		return s.data.Plans[i], nil
	}

	return models.Plan{}, ErrNotFound
}

func (s *Store) DeletePlan(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.data.Plans {
		if p.ID != id {
			continue
		}

		s.data.Plans = append(s.data.Plans[:i], s.data.Plans[i+1:]...)

		if err := s.save(); err != nil {
			return err
		}

		return nil
	}

	return ErrNotFound
}

func (s *Store) ListGoals() []models.Goal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]models.Goal, len(s.data.Goals))
	copy(out, s.data.Goals)
	return out
}

func (s *Store) CreateGoal(goal models.Goal) (models.Goal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, g := range s.data.Goals {
		if g.ID == goal.ID {
			return models.Goal{}, ErrAlreadyExists
		}
	}

	s.data.Goals = append([]models.Goal{goal}, s.data.Goals...)

	if err := s.save(); err != nil {
		s.data.Goals = s.data.Goals[1:]
		return models.Goal{}, err
	}

	return goal, nil
}

func (s *Store) UpdateGoal(id string, update func(*models.Goal) error) (models.Goal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.data.Goals {
		if s.data.Goals[i].ID != id {
			continue
		}

		before := s.data.Goals[i]
		if err := update(&s.data.Goals[i]); err != nil {
			return models.Goal{}, err
		}

		s.data.Goals[i].UpdatedAt = models.NowISO()

		if err := s.save(); err != nil {
			s.data.Goals[i] = before
			return models.Goal{}, err
		}

		return s.data.Goals[i], nil
	}

	return models.Goal{}, ErrNotFound
}

func (s *Store) DeleteGoal(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, g := range s.data.Goals {
		if g.ID != id {
			continue
		}

		s.data.Goals = append(s.data.Goals[:i], s.data.Goals[i+1:]...)

		if err := s.save(); err != nil {
			return err
		}

		return nil
	}

	return ErrNotFound
}

func (s *Store) ListHabits() []models.Habit {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]models.Habit, len(s.data.Habits))
	copy(out, s.data.Habits)
	return out
}

func (s *Store) CreateHabit(habit models.Habit) (models.Habit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, h := range s.data.Habits {
		if h.ID == habit.ID {
			return models.Habit{}, ErrAlreadyExists
		}
	}

	s.data.Habits = append([]models.Habit{habit}, s.data.Habits...)

	if err := s.save(); err != nil {
		s.data.Habits = s.data.Habits[1:]
		return models.Habit{}, err
	}

	return habit, nil
}

func (s *Store) UpdateHabit(id string, update func(*models.Habit) error) (models.Habit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.data.Habits {
		if s.data.Habits[i].ID != id {
			continue
		}

		before := s.data.Habits[i]
		if err := update(&s.data.Habits[i]); err != nil {
			return models.Habit{}, err
		}

		if err := s.save(); err != nil {
			s.data.Habits[i] = before
			return models.Habit{}, err
		}

		return s.data.Habits[i], nil
	}

	return models.Habit{}, ErrNotFound
}

func (s *Store) DeleteHabit(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, h := range s.data.Habits {
		if h.ID != id {
			continue
		}

		s.data.Habits = append(s.data.Habits[:i], s.data.Habits[i+1:]...)

		if err := s.save(); err != nil {
			return err
		}

		return nil
	}

	return ErrNotFound
}

func (s *Store) Overview() models.Overview {
	s.mu.RLock()
	defer s.mu.RUnlock()

	today := models.TodayDate()
	overview := models.Overview{
		PlansByCategory: map[string]int{},
		PlansByPriority: map[string]int{},
		RecentPlans:     []models.Plan{},
		UpcomingPlans:   []models.Plan{},
	}

	for _, p := range s.data.Plans {
		overview.TotalPlans++
		overview.PlansByCategory[string(p.Category)]++
		overview.PlansByPriority[string(p.Priority)]++

		switch p.Status {
		case models.PlanStatusDone:
			overview.CompletedPlans++
		case models.PlanStatusInProgress:
			overview.InProgressPlans++
		default:
			overview.TodoPlans++
		}

		if p.DueDate != "" && p.DueDate < today && p.Status != models.PlanStatusDone {
			overview.OverduePlans++
		}

		if p.DueDate != "" && p.DueDate >= today && p.Status != models.PlanStatusDone {
			if len(overview.UpcomingPlans) < 5 {
				overview.UpcomingPlans = append(overview.UpcomingPlans, p)
			}
		}
	}

	if len(s.data.Plans) > 5 {
		overview.RecentPlans = append(overview.RecentPlans, s.data.Plans[:5]...)
	} else {
		overview.RecentPlans = append(overview.RecentPlans, s.data.Plans...)
	}

	for _, g := range s.data.Goals {
		overview.TotalGoals++
		overview.AvgGoalProgress += float64(g.Progress)

		switch g.Status {
		case models.GoalStatusActive:
			overview.ActiveGoals++
		case models.GoalStatusCompleted:
			overview.CompletedGoals++
		}
	}

	if overview.TotalGoals > 0 {
		overview.AvgGoalProgress /= float64(overview.TotalGoals)
	}

	overview.TotalHabits = len(s.data.Habits)
	for _, h := range s.data.Habits {
		overview.TotalStreak += h.Streak
	}

	return overview
}
