package validate

import (
	"testing"

	"github.com/mitja6889/HTTPUtils/internal/models"
)

func TestDate(t *testing.T) {
	if err := Date("2026-08-17"); err != nil {
		t.Fatal(err)
	}
	if err := Date("invalid"); err == nil {
		t.Fatal("expected error")
	}
}

func TestEffectiveHabitStreak(t *testing.T) {
	today := "2026-08-17"
	h := models.Habit{Streak: 5, LastDone: "2026-08-16"}
	if got := EffectiveHabitStreak(h, today); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
	h.LastDone = "2026-08-15"
	if got := EffectiveHabitStreak(h, today); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestIcon(t *testing.T) {
	icon, err := Icon("💧")
	if err != nil || icon != "💧" {
		t.Fatalf("unexpected: %q %v", icon, err)
	}
	if _, err := Icon("<script>"); err == nil {
		t.Fatal("expected error for invalid icon")
	}
}
