package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/georgegromov/transactions-system/services/transaction-processor/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	CreateTransactionTimeout = 10 * time.Second
)

type validationService interface {
	Validate(ctx context.Context, s any) error
}

type transactionService interface {
	CreateTransaction(ctx context.Context, externalID, accountID uuid.UUID, amount decimal.Decimal, transactionType domain.TransactionType) error
}

type transactionsHandler struct {
	logger    *slog.Logger
	validator validationService
	service   transactionService
}

func NewTransactionsHandler(
	logger *slog.Logger,
	validator validationService,
	transactionService transactionService,
) *transactionsHandler {
	return &transactionsHandler{
		logger:    logger,
		validator: validator,
		service:   transactionService,
	}
}

func (h *transactionsHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	const op = "transactionsHandler.CreateTransaction"
	log := h.logger.With(slog.String("op", op))

	ctx, cancel := context.WithTimeout(r.Context(), CreateTransactionTimeout)
	defer cancel()

	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteResponse(log, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Validate(ctx, req); err != nil {
		WriteResponse(log, w, http.StatusBadRequest, err.Error())
		return
	}

	transactionType := domain.TransactionType(req.TransactionType)

	err := h.service.CreateTransaction(
		ctx,
		req.ExternalID,
		req.AccountID,
		req.Amount,
		transactionType,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTransactionAlreadyExists):
			WriteResponse(log, w, http.StatusConflict, err.Error())
		case errors.Is(err, domain.ErrInsufficientBalance):
			WriteResponse(log, w, http.StatusBadRequest, err.Error())
		default:
			WriteResponse(log, w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	WriteResponse(log, w, http.StatusCreated, nil)
}

type createTransactionRequest struct {
	ExternalID      uuid.UUID       `json:"external_id" validate:"required,uuid"`
	AccountID       uuid.UUID       `json:"account_id" validate:"required,uuid"`
	Amount          decimal.Decimal `json:"amount" validate:"required,gt=0"`
	TransactionType string          `json:"transaction_type" validate:"required,oneof=income expense"`
}
