package store

import (
	"path/filepath"
	"testing"

	"github.com/mitja6889/HTTPUtils/internal/models"
)

func TestCreatePlansFromTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")

	s, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	template := models.Plan{
		ID:       models.NewID(),
		Title:    "Telovadba",
		Category: models.CategoryHealth,
		Priority: models.PriorityMedium,
		Status:   models.PlanStatusTodo,
	}

	created, err := s.CreatePlansFromTemplate(template, []string{"2026-08-17", "2026-08-19", "2026-08-21"})
	if err != nil {
		t.Fatalf("CreatePlansFromTemplate: %v", err)
	}
	if len(created) != 3 {
		t.Fatalf("expected 3 plans, got %d", len(created))
	}

	seriesID := created[0].SeriesID
	if seriesID == "" {
		t.Fatal("expected series id")
	}
	for _, plan := range created {
		if plan.SeriesID != seriesID {
			t.Fatalf("series mismatch: %s vs %s", plan.SeriesID, seriesID)
		}
		if plan.Title != "Telovadba" {
			t.Fatalf("unexpected title: %s", plan.Title)
		}
	}

	updated, err := s.UpdatePlanAndSeries(
		created[0].ID,
		[]string{"2026-08-17", "2026-08-18", "2026-08-19", "2026-08-21"},
		func(p *models.Plan) error {
			p.Title = "Telovadba posodobljena"
			return nil
		},
		func(p *models.Plan) error {
			p.Status = models.PlanStatusInProgress
			return nil
		},
	)
	if err != nil {
		t.Fatalf("UpdatePlanAndSeries: %v", err)
	}
	if updated.Title != "Telovadba posodobljena" {
		t.Fatalf("unexpected updated title: %s", updated.Title)
	}
	if updated.Status != models.PlanStatusInProgress {
		t.Fatalf("unexpected status: %s", updated.Status)
	}

	all := s.ListPlans()
	if len(all) != 4 {
		t.Fatalf("expected 4 plans after update, got %d", len(all))
	}
}
