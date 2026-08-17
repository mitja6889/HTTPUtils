package store

import (
	"sort"

	"github.com/mitja6889/HTTPUtils/internal/models"
	"github.com/mitja6889/HTTPUtils/internal/validate"
)

func normalizeDueDates(dueDate string, dueDates []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(dueDates)+1)

	add := func(d string) {
		if d == "" {
			return
		}
		if _, ok := seen[d]; ok {
			return
		}
		seen[d] = struct{}{}
		out = append(out, d)
	}

	for _, d := range dueDates {
		add(d)
	}
	if len(out) == 0 {
		add(dueDate)
	}

	sort.Strings(out)
	return out
}

func clonePlanForDate(source models.Plan, dueDate, seriesID, now string) models.Plan {
	plan := source
	plan.ID = models.NewID()
	plan.DueDate = dueDate
	plan.SeriesID = seriesID
	plan.CreatedAt = now
	plan.UpdatedAt = now
	plan.Tasks = cloneTasks(source.Tasks, now)
	return plan
}

func cloneTasks(tasks []models.Task, now string) []models.Task {
	if len(tasks) == 0 {
		return []models.Task{}
	}

	out := make([]models.Task, 0, len(tasks))
	for _, task := range tasks {
		task.ID = models.NewID()
		if task.CreatedAt == "" {
			task.CreatedAt = now
		}
		out = append(out, task)
	}
	return out
}

func (s *Store) CreatePlansFromTemplate(template models.Plan, dueDates []string) ([]models.Plan, error) {
	dates := normalizeDueDates(template.DueDate, dueDates)
	if err := validate.Dates(dates); err != nil {
		return nil, err
	}

	seriesID := ""
	if len(dates) > 1 {
		seriesID = models.NewID()
	}

	now := models.NowISO()
	created := make([]models.Plan, 0, len(dates))

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, dueDate := range dates {
		plan := clonePlanForDate(template, dueDate, seriesID, now)
		if len(dates) == 1 {
			plan.SeriesID = ""
		}
		s.data.Plans = append([]models.Plan{plan}, s.data.Plans...)
		created = append(created, plan)
	}

	if err := s.save(); err != nil {
		s.data.Plans = s.data.Plans[len(created):]
		return nil, err
	}

	return created, nil
}

func (s *Store) UpdatePlanAndSeries(id string, dueDates []string, sharedUpdate func(*models.Plan) error, selfUpdate func(*models.Plan) error) (models.Plan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	targetIdx := -1
	for i := range s.data.Plans {
		if s.data.Plans[i].ID == id {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		return models.Plan{}, ErrNotFound
	}

	before := append([]models.Plan(nil), s.data.Plans...)
	target := s.data.Plans[targetIdx]
	dates := normalizeDueDates(target.DueDate, dueDates)
	if err := validate.Dates(dates); err != nil {
		return models.Plan{}, err
	}

	seriesID := target.SeriesID
	if len(dates) > 1 && seriesID == "" {
		seriesID = models.NewID()
	}
	if len(dates) == 1 {
		seriesID = ""
	}

	seriesIndices := make([]int, 0)
	if target.SeriesID != "" {
		for i, plan := range s.data.Plans {
			if plan.SeriesID == target.SeriesID {
				seriesIndices = append(seriesIndices, i)
			}
		}
	} else {
		seriesIndices = []int{targetIdx}
	}

	dateToIdx := map[string]int{}
	for _, idx := range seriesIndices {
		dateToIdx[s.data.Plans[idx].DueDate] = idx
	}

	now := models.NowISO()
	keep := map[string]struct{}{}
	for _, dueDate := range dates {
		keep[dueDate] = struct{}{}
		if idx, ok := dateToIdx[dueDate]; ok {
			plan := &s.data.Plans[idx]
			plan.SeriesID = seriesID
			if sharedUpdate != nil {
				if err := sharedUpdate(plan); err != nil {
					return models.Plan{}, err
				}
			}
			if plan.ID == id && selfUpdate != nil {
				if err := selfUpdate(plan); err != nil {
					return models.Plan{}, err
				}
			}
			plan.UpdatedAt = now
			continue
		}

		newPlan := clonePlanForDate(target, dueDate, seriesID, now)
		if sharedUpdate != nil {
			if err := sharedUpdate(&newPlan); err != nil {
				return models.Plan{}, err
			}
		}
		s.data.Plans = append([]models.Plan{newPlan}, s.data.Plans...)
	}

	remove := make([]int, 0)
	for _, idx := range seriesIndices {
		plan := s.data.Plans[idx]
		if _, ok := keep[plan.DueDate]; ok {
			continue
		}
		remove = append(remove, idx)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(remove)))
	for _, idx := range remove {
		if s.data.Plans[idx].ID == id {
			return models.Plan{}, ErrNotFound
		}
		s.data.Plans = append(s.data.Plans[:idx], s.data.Plans[idx+1:]...)
	}

	if selfUpdate != nil {
		for i := range s.data.Plans {
			if s.data.Plans[i].ID != id {
				continue
			}
			if err := selfUpdate(&s.data.Plans[i]); err != nil {
				return models.Plan{}, err
			}
			s.data.Plans[i].UpdatedAt = now
			if len(dates) == 1 {
				s.data.Plans[i].SeriesID = ""
			} else {
				s.data.Plans[i].SeriesID = seriesID
			}
			if err := s.save(); err != nil {
				s.data.Plans = before
				return models.Plan{}, err
			}
			return s.data.Plans[i], nil
		}
		return models.Plan{}, ErrNotFound
	}

	for i := range s.data.Plans {
		if s.data.Plans[i].ID != id {
			continue
		}
		if len(dates) == 1 {
			s.data.Plans[i].SeriesID = ""
		}
		if err := s.save(); err != nil {
			s.data.Plans = before
			return models.Plan{}, err
		}
		return s.data.Plans[i], nil
	}

	return models.Plan{}, ErrNotFound
}
