package store

import (
	"time"

	"github.com/mitja6889/HTTPUtils/internal/models"
)

func (s *Store) SeedIfEmpty() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.data.Plans) > 0 || len(s.data.Goals) > 0 || len(s.data.Habits) > 0 {
		return nil
	}

	s.data = demoData()
	return s.save()
}

func (s *Store) SeedDemo() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = demoData()
	return s.save()
}

func demoData() models.DataStore {
	now := time.Now().UTC()
	today := now.Format("2006-01-02")

	plans := []models.Plan{
		plan("seed01", "Jutranja meditacija", "10 minut dihanja in fokusa pred začetkom dneva", models.CategoryHealth, models.PriorityHigh, models.PlanStatusInProgress, dayOffset(today, 0), now),
		plan("seed02", "Pregled emailov in prioritet", "Odgovori na nujna sporočila, razporedi naloge", models.CategoryWork, models.PriorityMedium, models.PlanStatusTodo, dayOffset(today, 0), now),
		plan("seed03", "Telovadba — 30 minut", "Kombinacija kardio in raztezanja", models.CategoryHealth, models.PriorityMedium, models.PlanStatusTodo, dayOffset(today, 1), now),
		plan("seed04", "Dokončaj projektno poročilo", "Finaliziraj Q3 poročilo in pošlji nadrejenemu", models.CategoryWork, models.PriorityHigh, models.PlanStatusTodo, dayOffset(today, 1), now),
		plan("seed05", "Branje — 1 ura", "Nadaljuj z knjigo Atomic Habits", models.CategoryLearning, models.PriorityLow, models.PlanStatusTodo, dayOffset(today, 2), now),
		plan("seed06", "Srečanje z ekipo", "Tedenski sync — pripravi update o napredku", models.CategoryWork, models.PriorityHigh, models.PlanStatusTodo, dayOffset(today, 2), now),
		plan("seed07", "Nakup zdravih živil", "Priprava obrokov za preostali teden", models.CategoryPersonal, models.PriorityLow, models.PlanStatusTodo, dayOffset(today, 3), now),
		plan("seed08", "Priprava na predstavitev", "Slides + vaja predstavitve (15 min)", models.CategoryWork, models.PriorityHigh, models.PlanStatusInProgress, dayOffset(today, 3), now),
		plan("seed09", "Sprehod v naravi", "Vsaj 5 km hoje ali teka v parku", models.CategoryHealth, models.PriorityMedium, models.PlanStatusTodo, dayOffset(today, 4), now),
		plan("seed10", "Online tečaj Go", "Poglavje o REST API-jih in testiranju", models.CategoryLearning, models.PriorityMedium, models.PlanStatusTodo, dayOffset(today, 4), now),
		plan("seed11", "Generalno čiščenje stanovanja", "Kuhinja, kopalnica, organizacija omare", models.CategoryPersonal, models.PriorityLow, models.PlanStatusTodo, dayOffset(today, 5), now),
		plan("seed12", "Planiranje naslednjega tedna", "Preglej koledar, nastavi prioritete", models.CategoryWork, models.PriorityMedium, models.PlanStatusTodo, dayOffset(today, 5), now),
		plan("seed13", "Družinski izlet", "Popoldanski izlet — brez telefona", models.CategoryPersonal, models.PriorityMedium, models.PlanStatusTodo, dayOffset(today, 6), now),
		plan("seed14", "Tedenski pregled navad", "Oceni streak, prilagodi rutine", models.CategoryOther, models.PriorityLow, models.PlanStatusTodo, dayOffset(today, 6), now),
	}

	goals := []models.Goal{
		{
			ID:          "goal01",
			Title:       "Teči pol maraton",
			Description: "Priprava na ljubljanski polmaraton — 3x tedensko",
			TargetDate:  dayOffset(today, 60),
			Progress:    35,
			Status:      models.GoalStatusActive,
			CreatedAt:   models.NowISO(),
			UpdatedAt:   models.NowISO(),
		},
		{
			ID:          "goal02",
			Title:       "Nauči se Go programiranja",
			Description: "Dokončaj online tečaj in napiši prvo API aplikacijo",
			TargetDate:  dayOffset(today, 45),
			Progress:    60,
			Status:      models.GoalStatusActive,
			CreatedAt:   models.NowISO(),
			UpdatedAt:   models.NowISO(),
		},
		{
			ID:          "goal03",
			Title:       "Varčevanje za potovanje",
			Description: "Cilj: 1500 € do poletja",
			TargetDate:  dayOffset(today, 90),
			Progress:    20,
			Status:      models.GoalStatusActive,
			CreatedAt:   models.NowISO(),
			UpdatedAt:   models.NowISO(),
		},
	}

	habits := []models.Habit{
		{ID: "hab01", Name: "Jutranja voda", Icon: "💧", Streak: 5, LastDone: dayOffset(today, 0), CreatedAt: models.NowISO()},
		{ID: "hab02", Name: "Meditacija", Icon: "🧘", Streak: 3, LastDone: dayOffset(today, -1), CreatedAt: models.NowISO()},
		{ID: "hab03", Name: "Branje", Icon: "📚", Streak: 7, LastDone: dayOffset(today, 0), CreatedAt: models.NowISO()},
		{ID: "hab04", Name: "Telovadba", Icon: "💪", Streak: 2, LastDone: dayOffset(today, -1), CreatedAt: models.NowISO()},
	}

	return models.DataStore{
		Version: models.DataStoreVersion,
		Plans:   plans,
		Goals:  goals,
		Habits: habits,
	}
}

func plan(id, title, desc string, cat models.Category, pri models.Priority, status models.PlanStatus, due string, now time.Time) models.Plan {
	ts := now.Format(time.RFC3339)
	return models.Plan{
		ID:          id,
		Title:       title,
		Description: desc,
		Category:    cat,
		Priority:    pri,
		Status:      status,
		DueDate:     due,
		Tasks:       []models.Task{},
		CreatedAt:   ts,
		UpdatedAt:   ts,
	}
}

func dayOffset(base string, days int) string {
	t, err := time.Parse("2006-01-02", base)
	if err != nil {
		return base
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}
