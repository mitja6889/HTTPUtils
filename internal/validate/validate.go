package validate

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mitja6889/HTTPUtils/internal/models"
)

var (
	ErrInvalidDate   = errors.New("invalid date, expected YYYY-MM-DD")
	ErrInvalidEnum   = errors.New("invalid value")
	ErrInvalidIcon   = errors.New("icon must be a single emoji")
	ErrEmptyTitle    = errors.New("title is required")
	ErrEmptyName     = errors.New("name is required")
)

var allowedIcons = map[string]struct{}{
	"✨": {}, "🏃": {}, "📚": {}, "💧": {}, "🧘": {}, "💪": {}, "🥗": {}, "😴": {}, "📝": {}, "🎨": {},
	"🌿": {}, "🔥": {}, "⭐": {}, "🎯": {}, "☀️": {}, "🌙": {}, "🍎": {}, "🚶": {}, "🧠": {}, "❤️": {},
}

func Date(value string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return ErrInvalidDate
	}
	return nil
}

func Category(value models.Category) error {
	switch value {
	case models.CategoryWork, models.CategoryHealth, models.CategoryPersonal, models.CategoryLearning, models.CategoryOther:
		return nil
	default:
		return fmt.Errorf("%w: category", ErrInvalidEnum)
	}
}

func Priority(value models.Priority) error {
	switch value {
	case models.PriorityLow, models.PriorityMedium, models.PriorityHigh:
		return nil
	default:
		return fmt.Errorf("%w: priority", ErrInvalidEnum)
	}
}

func PlanStatus(value models.PlanStatus) error {
	switch value {
	case models.PlanStatusTodo, models.PlanStatusInProgress, models.PlanStatusDone:
		return nil
	default:
		return fmt.Errorf("%w: status", ErrInvalidEnum)
	}
}

func GoalStatus(value models.GoalStatus) error {
	switch value {
	case models.GoalStatusActive, models.GoalStatusCompleted, models.GoalStatusPaused:
		return nil
	default:
		return fmt.Errorf("%w: goal status", ErrInvalidEnum)
	}
}

func Title(value string) (string, error) {
	title := strings.TrimSpace(value)
	if title == "" {
		return "", ErrEmptyTitle
	}
	return title, nil
}

func Name(value string) (string, error) {
	name := strings.TrimSpace(value)
	if name == "" {
		return "", ErrEmptyName
	}
	return name, nil
}

func Icon(value string) (string, error) {
	icon := strings.TrimSpace(value)
	if icon == "" {
		return "✨", nil
	}
	if _, ok := allowedIcons[icon]; ok {
		return icon, nil
	}
	if utf8.RuneCountInString(icon) == 1 {
		return icon, nil
	}
	return "", ErrInvalidIcon
}

func ClampProgress(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func ClientToday(value string) (string, error) {
	if value == "" {
		return models.TodayDate(), nil
	}
	if err := Date(value); err != nil {
		return "", err
	}
	return value, nil
}

func AddDays(dateStr string, days int) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}

func EffectiveHabitStreak(h models.Habit, today string) int {
	if h.LastDone == "" {
		return 0
	}
	yesterday := AddDays(today, -1)
	if h.LastDone == today || h.LastDone == yesterday {
		return h.Streak
	}
	return 0
}
