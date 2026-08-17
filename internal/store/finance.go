package store

import (
	"math"
	"sort"

	"github.com/mitja6889/HTTPUtils/internal/models"
	"github.com/mitja6889/HTTPUtils/internal/validate"
)

func (s *Store) ListTransactions() []models.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]models.Transaction, len(s.data.Transactions))
	copy(out, s.data.Transactions)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Date == out[j].Date {
			return out[i].CreatedAt > out[j].CreatedAt
		}
		return out[i].Date > out[j].Date
	})
	return out
}

func (s *Store) GetTransaction(id string) (models.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.data.Transactions {
		if t.ID == id {
			return t, nil
		}
	}
	return models.Transaction{}, ErrNotFound
}

func (s *Store) CreateTransaction(tx models.Transaction) (models.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.data.Transactions {
		if t.ID == tx.ID {
			return models.Transaction{}, ErrAlreadyExists
		}
	}

	s.data.Transactions = append([]models.Transaction{tx}, s.data.Transactions...)
	if err := s.save(); err != nil {
		s.data.Transactions = s.data.Transactions[1:]
		return models.Transaction{}, err
	}

	s.syncFinancialGoalsLocked()
	return tx, nil
}

func (s *Store) UpdateTransaction(id string, update func(*models.Transaction) error) (models.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.data.Transactions {
		if s.data.Transactions[i].ID != id {
			continue
		}

		before := s.data.Transactions[i]
		if err := update(&s.data.Transactions[i]); err != nil {
			return models.Transaction{}, err
		}

		if err := s.save(); err != nil {
			s.data.Transactions[i] = before
			return models.Transaction{}, err
		}

		s.syncFinancialGoalsLocked()
		return s.data.Transactions[i], nil
	}

	return models.Transaction{}, ErrNotFound
}

func (s *Store) DeleteTransaction(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.data.Transactions {
		if t.ID != id {
			continue
		}

		removed := t
		s.data.Transactions = append(s.data.Transactions[:i], s.data.Transactions[i+1:]...)

		if err := s.save(); err != nil {
			s.data.Transactions = append(s.data.Transactions[:i], append([]models.Transaction{removed}, s.data.Transactions[i:]...)...)
			return err
		}

		s.syncFinancialGoalsLocked()
		return nil
	}

	return ErrNotFound
}

func (s *Store) FinanceStats(today string) models.FinanceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.financeStatsLocked(today)
}

func (s *Store) financeStatsLocked(today string) models.FinanceStats {
	stats := models.FinanceStats{
		ByCategory:         map[string]float64{},
		MonthlyTrend:       []models.MonthlyPoint{},
		RecentTransactions: []models.Transaction{},
	}

	monthPrefix := today[:7]
	monthly := map[string]*models.MonthlyPoint{}

	for _, t := range s.data.Transactions {
		switch t.Type {
		case models.TransactionIncome:
			stats.TotalIncome += t.Amount
			if len(t.Date) >= 7 && t.Date[:7] == monthPrefix {
				stats.MonthIncome += t.Amount
			}
		case models.TransactionExpense:
			stats.TotalExpense += t.Amount
			stats.ByCategory[string(t.Category)] += t.Amount
			if len(t.Date) >= 7 && t.Date[:7] == monthPrefix {
				stats.MonthExpense += t.Amount
			}
		}

		if len(t.Date) >= 7 {
			key := t.Date[:7]
			if monthly[key] == nil {
				monthly[key] = &models.MonthlyPoint{Month: key}
			}
			if t.Type == models.TransactionIncome {
				monthly[key].Income += t.Amount
			} else {
				monthly[key].Expense += t.Amount
			}
		}
	}

	stats.Balance = s.data.Finance.InitialBalance + stats.TotalIncome - stats.TotalExpense
	stats.MonthBalance = stats.MonthIncome - stats.MonthExpense

	for _, p := range monthly {
		stats.MonthlyTrend = append(stats.MonthlyTrend, *p)
	}
	sort.Slice(stats.MonthlyTrend, func(i, j int) bool {
		return stats.MonthlyTrend[i].Month < stats.MonthlyTrend[j].Month
	})
	if len(stats.MonthlyTrend) > 6 {
		stats.MonthlyTrend = stats.MonthlyTrend[len(stats.MonthlyTrend)-6:]
	}

	limit := 8
	if len(s.data.Transactions) < limit {
		limit = len(s.data.Transactions)
	}
	for i := 0; i < limit; i++ {
		stats.RecentTransactions = append(stats.RecentTransactions, s.data.Transactions[i])
	}

	return stats
}

func (s *Store) syncFinancialGoalsLocked() {
	for i := range s.data.Goals {
		g := s.enrichGoalLocked(s.data.Goals[i])
		s.data.Goals[i].CurrentAmount = g.CurrentAmount
		s.data.Goals[i].Progress = g.Progress
		if g.Status != s.data.Goals[i].Status && g.Progress >= 100 {
			s.data.Goals[i].Status = models.GoalStatusCompleted
		}
		s.data.Goals[i].UpdatedAt = models.NowISO()
	}
}

func (s *Store) EnrichGoal(g models.Goal) models.Goal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enrichGoalLocked(g)
}

func (s *Store) enrichGoalLocked(g models.Goal) models.Goal {
	if g.Kind != models.GoalKindFinancial || g.TargetAmount <= 0 {
		return g
	}

	var saved float64
	for _, t := range s.data.Transactions {
		if t.GoalID != g.ID {
			continue
		}
		if t.Type == models.TransactionIncome {
			saved += t.Amount
		} else {
			saved -= t.Amount
		}
	}

	g.CurrentAmount = math.Max(0, saved)
	g.Progress = int(math.Min(100, g.CurrentAmount/g.TargetAmount*100))
	if g.Progress >= 100 && g.Status == models.GoalStatusActive {
		g.Status = models.GoalStatusCompleted
	}
	return g
}

func (s *Store) buildChartsLocked(today string) models.DashboardCharts {
	charts := models.DashboardCharts{
		PlansByStatus:     []models.ChartPoint{},
		GoalsProgress:     []models.ChartPoint{},
		ExpenseCategories: []models.ChartPoint{},
		HabitsActivity:    []models.ChartPoint{},
	}

	statusCounts := map[string]float64{}
	for _, p := range s.data.Plans {
		statusCounts[string(p.Status)]++
	}
	for k, v := range statusCounts {
		charts.PlansByStatus = append(charts.PlansByStatus, models.ChartPoint{Label: k, Value: v})
	}

	for _, g := range s.data.Goals {
		enriched := s.enrichGoalLocked(g)
		charts.GoalsProgress = append(charts.GoalsProgress, models.ChartPoint{
			Label: g.Title,
			Value: float64(enriched.Progress),
		})
	}

	finance := s.financeStatsLocked(today)
	charts.FinanceMonthly = finance.MonthlyTrend
	for cat, amount := range finance.ByCategory {
		if amount > 0 {
			charts.ExpenseCategories = append(charts.ExpenseCategories, models.ChartPoint{
				Label: cat,
				Value: amount,
			})
		}
	}
	sort.Slice(charts.ExpenseCategories, func(i, j int) bool {
		return charts.ExpenseCategories[i].Value > charts.ExpenseCategories[j].Value
	})

	for _, h := range s.data.Habits {
		charts.HabitsActivity = append(charts.HabitsActivity, models.ChartPoint{
			Label: h.Name,
			Value: float64(validate.EffectiveHabitStreak(h, today)),
		})
	}

	return charts
}
