package models

import "time"

type PlanStatus string

const (
	PlanStatusTodo       PlanStatus = "todo"
	PlanStatusInProgress PlanStatus = "in_progress"
	PlanStatusDone       PlanStatus = "done"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Category string

const (
	CategoryWork     Category = "work"
	CategoryHealth   Category = "health"
	CategoryPersonal Category = "personal"
	CategoryLearning Category = "learning"
	CategoryOther    Category = "other"
)

type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"createdAt"`
}

type Plan struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    Category   `json:"category"`
	Priority    Priority   `json:"priority"`
	Status      PlanStatus `json:"status"`
	DueDate     string     `json:"dueDate,omitempty"`
	Tasks       []Task     `json:"tasks"`
	CreatedAt   string     `json:"createdAt"`
	UpdatedAt   string     `json:"updatedAt"`
}

type GoalStatus string

const (
	GoalStatusActive    GoalStatus = "active"
	GoalStatusCompleted GoalStatus = "completed"
	GoalStatusPaused    GoalStatus = "paused"
)

type Goal struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	TargetDate  string     `json:"targetDate,omitempty"`
	Progress    int        `json:"progress"`
	Status      GoalStatus `json:"status"`
	CreatedAt   string     `json:"createdAt"`
	UpdatedAt   string     `json:"updatedAt"`
}

type Habit struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Streak    int    `json:"streak"`
	LastDone  string `json:"lastDone,omitempty"`
	CreatedAt string `json:"createdAt"`
}

type DataStore struct {
	Plans  []Plan  `json:"plans"`
	Goals  []Goal  `json:"goals"`
	Habits []Habit `json:"habits"`
}

type Overview struct {
	TotalPlans      int            `json:"totalPlans"`
	CompletedPlans  int            `json:"completedPlans"`
	InProgressPlans int            `json:"inProgressPlans"`
	TodoPlans       int            `json:"todoPlans"`
	OverduePlans    int            `json:"overduePlans"`
	TotalGoals      int            `json:"totalGoals"`
	ActiveGoals     int            `json:"activeGoals"`
	CompletedGoals  int            `json:"completedGoals"`
	AvgGoalProgress float64        `json:"avgGoalProgress"`
	TotalHabits     int            `json:"totalHabits"`
	TotalStreak     int            `json:"totalStreak"`
	PlansByCategory map[string]int `json:"plansByCategory"`
	PlansByPriority map[string]int `json:"plansByPriority"`
	RecentPlans     []Plan         `json:"recentPlans"`
	UpcomingPlans   []Plan         `json:"upcomingPlans"`
}

func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func TodayDate() string {
	return time.Now().UTC().Format("2006-01-02")
}
