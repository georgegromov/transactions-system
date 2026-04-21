package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/georgegromov/transactions-system/services/balance-viewer/internal/domain"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	GetAccountBalanceTimeout = 10 * time.Second
)

type balanceService interface {
	GetAccountBalance(ctx context.Context, accountID uuid.UUID) (*domain.Balance, error)
}

type balanceHandler struct {
	logger  *slog.Logger
	service balanceService
}

func NewBalanceHandler(logger *slog.Logger, service balanceService) *balanceHandler {
	return &balanceHandler{logger: logger, service: service}
}

func (h *balanceHandler) GetAccountBalance(w http.ResponseWriter, r *http.Request) {
	const op = "balanceHandler.GetAccountBalance"
	log := h.logger.With(slog.String("op", op))

	ctx, cancel := context.WithTimeout(r.Context(), GetAccountBalanceTimeout)
	defer cancel()

	accountID, err := uuid.Parse(chi.URLParam(r, "account_id"))
	if err != nil {
		WriteResponse(log, w, http.StatusBadRequest, "invalid account_id format")
		return
	}

	balance, err := h.service.GetAccountBalance(ctx, accountID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAccountBalanceNotFound):
			WriteResponse(log, w, http.StatusNotFound, err.Error())
		default:
			WriteResponse(log, w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	WriteResponse(log, w, http.StatusOK, toGetAccountBalanceResponse(balance))
}

type GetAccountBalanceResponse struct {
	AccountID      uuid.UUID       `json:"account_id"`
	CurrentBalance decimal.Decimal `json:"current_balance"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func toGetAccountBalanceResponse(balance *domain.Balance) *GetAccountBalanceResponse {
	return &GetAccountBalanceResponse{
		AccountID:      balance.AccountID(),
		CurrentBalance: balance.CurrentBalance(),
		UpdatedAt:      balance.UpdatedAt(),
	}
}
