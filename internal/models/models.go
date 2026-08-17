package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

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
	SeriesID    string     `json:"seriesId,omitempty"`
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

type GoalKind string

const (
	GoalKindManual    GoalKind = "manual"
	GoalKindFinancial GoalKind = "financial"
)

type Goal struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Kind          GoalKind   `json:"kind"`
	TargetDate    string     `json:"targetDate,omitempty"`
	TargetAmount  float64    `json:"targetAmount,omitempty"`
	CurrentAmount float64    `json:"currentAmount,omitempty"`
	Progress      int        `json:"progress"`
	Status        GoalStatus `json:"status"`
	CreatedAt     string     `json:"createdAt"`
	UpdatedAt     string     `json:"updatedAt"`
}

type Habit struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Streak    int    `json:"streak"`
	LastDone  string `json:"lastDone,omitempty"`
	CreatedAt string `json:"createdAt"`
}

type TransactionType string

const (
	TransactionIncome  TransactionType = "income"
	TransactionExpense TransactionType = "expense"
)

type FinanceCategory string

const (
	FinanceSalary       FinanceCategory = "salary"
	FinanceFreelance    FinanceCategory = "freelance"
	FinanceInvestment   FinanceCategory = "investment"
	FinanceFood         FinanceCategory = "food"
	FinanceTransport    FinanceCategory = "transport"
	FinanceHousing      FinanceCategory = "housing"
	FinanceEntertainment FinanceCategory = "entertainment"
	FinanceHealth       FinanceCategory = "health"
	FinanceShopping     FinanceCategory = "shopping"
	FinanceOther        FinanceCategory = "other"
)

type Transaction struct {
	ID          string            `json:"id"`
	Type        TransactionType   `json:"type"`
	Amount      float64           `json:"amount"`
	Category    FinanceCategory   `json:"category"`
	Description string            `json:"description"`
	Date        string            `json:"date"`
	GoalID      string            `json:"goalId,omitempty"`
	CreatedAt   string            `json:"createdAt"`
}

type FinanceSettings struct {
	InitialBalance float64 `json:"initialBalance"`
}

type MonthlyPoint struct {
	Month   string  `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

type ChartPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type FinanceStats struct {
	Balance              float64            `json:"balance"`
	TotalIncome          float64            `json:"totalIncome"`
	TotalExpense         float64            `json:"totalExpense"`
	MonthIncome          float64            `json:"monthIncome"`
	MonthExpense         float64            `json:"monthExpense"`
	MonthBalance         float64            `json:"monthBalance"`
	ByCategory           map[string]float64 `json:"byCategory"`
	MonthlyTrend         []MonthlyPoint     `json:"monthlyTrend"`
	RecentTransactions   []Transaction      `json:"recentTransactions"`
}

type DashboardCharts struct {
	PlansByStatus      []ChartPoint   `json:"plansByStatus"`
	GoalsProgress      []ChartPoint   `json:"goalsProgress"`
	FinanceMonthly     []MonthlyPoint `json:"financeMonthly"`
	ExpenseCategories  []ChartPoint   `json:"expenseCategories"`
	HabitsActivity     []ChartPoint   `json:"habitsActivity"`
}

type DataStore struct {
	Version      int               `json:"version"`
	Plans        []Plan            `json:"plans"`
	Goals        []Goal            `json:"goals"`
	Habits       []Habit           `json:"habits"`
	Transactions []Transaction     `json:"transactions"`
	Finance      FinanceSettings   `json:"finance"`
}

const DataStoreVersion = 2

type Overview struct {
	TotalPlans       int            `json:"totalPlans"`
	CompletedPlans   int            `json:"completedPlans"`
	InProgressPlans  int            `json:"inProgressPlans"`
	TodoPlans        int            `json:"todoPlans"`
	OverduePlans     int            `json:"overduePlans"`
	TotalGoals       int            `json:"totalGoals"`
	ActiveGoals      int            `json:"activeGoals"`
	CompletedGoals   int            `json:"completedGoals"`
	PausedGoals      int            `json:"pausedGoals"`
	AvgGoalProgress  float64        `json:"avgGoalProgress"`
	TotalHabits      int            `json:"totalHabits"`
	TotalStreak      int            `json:"totalStreak"`
	PlansByCategory  map[string]int `json:"plansByCategory"`
	PlansByPriority  map[string]int `json:"plansByPriority"`
	RecentPlans      []Plan         `json:"recentPlans"`
	UpcomingPlans    []Plan         `json:"upcomingPlans"`
	OverduePlansList []Plan         `json:"overduePlansList"`
	Finance          FinanceStats   `json:"finance"`
	Charts           DashboardCharts `json:"charts"`
}

func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func NewID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return NowISO()
	}
	return hex.EncodeToString(buf)
}

func TodayDate() string {
	return time.Now().UTC().Format("2006-01-02")
}
