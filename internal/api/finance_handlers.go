package api

import (
	"net/http"
	"strings"

	"github.com/mitja6889/HTTPUtils/internal/models"
	"github.com/mitja6889/HTTPUtils/internal/validate"
)

func (a *API) handleFinanceStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	today, err := clientToday(r)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	printhttpWriteJSON(w, http.StatusOK, a.store.FinanceStats(today))
}

func (a *API) handleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		printhttpWriteJSON(w, http.StatusOK, a.store.ListTransactions())
	case http.MethodPost:
		var input createTransactionInput
		if err := decodeJSON(w, r, &input); err != nil {
			writeBadRequest(w, "invalid JSON body")
			return
		}

		tx, err := a.createTransaction(input)
		if err != nil {
			writeBadRequest(w, err.Error())
			return
		}

		printhttpWriteJSON(w, http.StatusCreated, tx)
	default:
		writeMethodNotAllowed(w)
	}
}

func (a *API) handleTransactionByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/transactions/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		writeBadRequest(w, "invalid transaction id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		tx, err := a.store.GetTransaction(id)
		if err != nil {
			a.writeStoreError(w, err)
			return
		}
		printhttpWriteJSON(w, http.StatusOK, tx)
	case http.MethodPut, http.MethodPatch:
		var input updateTransactionInput
		if err := decodeJSON(w, r, &input); err != nil {
			writeBadRequest(w, "invalid JSON body")
			return
		}

		tx, err := a.store.UpdateTransaction(id, func(t *models.Transaction) error {
			return applyTransactionUpdate(t, input)
		})
		if err != nil {
			a.writeStoreError(w, err)
			return
		}

		printhttpWriteJSON(w, http.StatusOK, tx)
	case http.MethodDelete:
		if err := a.store.DeleteTransaction(id); err != nil {
			a.writeStoreError(w, err)
			return
		}
		writeNoContent(w)
	default:
		writeMethodNotAllowed(w)
	}
}

type createTransactionInput struct {
	Type        models.TransactionType  `json:"type"`
	Amount      float64                 `json:"amount"`
	Category    models.FinanceCategory  `json:"category"`
	Description string                  `json:"description"`
	Date        string                  `json:"date"`
	GoalID      string                  `json:"goalId"`
}

type updateTransactionInput struct {
	Type        *models.TransactionType  `json:"type"`
	Amount      *float64                 `json:"amount"`
	Category    *models.FinanceCategory  `json:"category"`
	Description *string                  `json:"description"`
	Date        *string                  `json:"date"`
	GoalID      *string                  `json:"goalId"`
}

func (a *API) createTransaction(input createTransactionInput) (models.Transaction, error) {
	if err := validate.TransactionType(input.Type); err != nil {
		return models.Transaction{}, err
	}
	if err := validate.Amount(input.Amount); err != nil {
		return models.Transaction{}, err
	}

	category := input.Category
	if category == "" {
		category = models.FinanceOther
	}
	if err := validate.FinanceCategory(category); err != nil {
		return models.Transaction{}, err
	}

	date := input.Date
	if date == "" {
		date = models.TodayDate()
	}
	if err := validate.Date(date); err != nil {
		return models.Transaction{}, err
	}

	tx := models.Transaction{
		ID:          newID(),
		Type:        input.Type,
		Amount:      input.Amount,
		Category:    category,
		Description: strings.TrimSpace(input.Description),
		Date:        date,
		GoalID:      strings.TrimSpace(input.GoalID),
		CreatedAt:   models.NowISO(),
	}

	return a.store.CreateTransaction(tx)
}

func applyTransactionUpdate(t *models.Transaction, input updateTransactionInput) error {
	if input.Type != nil {
		if err := validate.TransactionType(*input.Type); err != nil {
			return err
		}
		t.Type = *input.Type
	}
	if input.Amount != nil {
		if err := validate.Amount(*input.Amount); err != nil {
			return err
		}
		t.Amount = *input.Amount
	}
	if input.Category != nil {
		if err := validate.FinanceCategory(*input.Category); err != nil {
			return err
		}
		t.Category = *input.Category
	}
	if input.Description != nil {
		t.Description = strings.TrimSpace(*input.Description)
	}
	if input.Date != nil {
		if err := validate.Date(*input.Date); err != nil {
			return err
		}
		t.Date = *input.Date
	}
	if input.GoalID != nil {
		t.GoalID = strings.TrimSpace(*input.GoalID)
	}
	return nil
}
