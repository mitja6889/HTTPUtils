package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mitja6889/HTTPUtils/internal/models"
)

func TestStoreCreateAndOverview(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lifeflow.json")
	s, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := s.CreatePlan(models.Plan{
		ID:        "p1",
		Title:     "Test",
		Category:  models.CategoryWork,
		Priority:  models.PriorityHigh,
		Status:    models.PlanStatusTodo,
		DueDate:   "2020-01-01",
		Tasks:     []models.Task{},
		CreatedAt: models.NowISO(),
		UpdatedAt: models.NowISO(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.ID != "p1" {
		t.Fatalf("unexpected plan: %+v", plan)
	}

	overview := s.Overview("2026-08-17")
	if overview.OverduePlans != 1 {
		t.Fatalf("expected 1 overdue plan, got %d", overview.OverduePlans)
	}
	if len(overview.OverduePlansList) != 1 {
		t.Fatalf("expected overdue list length 1")
	}
}

func TestDeleteRollbackOnSaveFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lifeflow.json")

	s, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.CreatePlan(models.Plan{
		ID: "p1", Title: "Keep", Category: models.CategoryWork, Priority: models.PriorityLow,
		Status: models.PlanStatusTodo, Tasks: []models.Task{}, CreatedAt: models.NowISO(), UpdatedAt: models.NowISO(),
	}); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}

	err = s.DeletePlan("p1")
	if err == nil {
		t.Fatal("expected save failure")
	}

	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}

	plans := s.ListPlans()
	if len(plans) != 1 {
		t.Fatalf("expected plan to remain in memory, got %d plans", len(plans))
	}
}

func TestLoadCorruptFileRecovers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lifeflow.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o600); err != nil {
		t.Fatal(err)
	}

	s, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(s.ListPlans()) != 0 {
		t.Fatal("expected empty store after corrupt recovery")
	}

	matches, _ := filepath.Glob(path + ".corrupt.*")
	if len(matches) != 1 {
		t.Fatalf("expected corrupt backup, got %v", matches)
	}
}
