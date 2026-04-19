package handler

import (
	"net/http"
	"time"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type PaymentDetailsHandler struct {
	paymentDetailsSvc service.PaymentDetailsService
}

func NewPaymentDetailsHandler(paymentDetailsSvc service.PaymentDetailsService) *PaymentDetailsHandler {
	return &PaymentDetailsHandler{paymentDetailsSvc: paymentDetailsSvc}
}

type setPaymentDetailsRequest struct {
	EntityType           string `json:"entity_type"`
	BankCardNumber       string `json:"bank_card_number,omitempty"`
	CardHolderName       string `json:"card_holder_name,omitempty"`
	BankAccount          string `json:"bank_account,omitempty"`
	BIK                  string `json:"bik,omitempty"`
	INN                  string `json:"inn,omitempty"`
	CorrespondentAccount string `json:"correspondent_account,omitempty"`
	BankName             string `json:"bank_name,omitempty"`
}

type paymentDetailsResponse struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id"`
	EntityType           string    `json:"entity_type"`
	BankCardNumber       string    `json:"bank_card_number,omitempty"`
	CardHolderName       string    `json:"card_holder_name,omitempty"`
	BankAccount          string    `json:"bank_account,omitempty"`
	BIK                  string    `json:"bik,omitempty"`
	INN                  string    `json:"inn,omitempty"`
	CorrespondentAccount string    `json:"correspondent_account,omitempty"`
	BankName             string    `json:"bank_name,omitempty"`
	IsVerified           bool      `json:"is_verified"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func toPaymentDetailsResponse(pd *domain.PaymentDetails) paymentDetailsResponse {
	resp := paymentDetailsResponse{
		ID:         pd.ID.String(),
		UserID:     pd.UserID.String(),
		EntityType: string(pd.EntityType),
		IsVerified: pd.IsVerified,
		CreatedAt:  pd.CreatedAt,
		UpdatedAt:  pd.UpdatedAt,
		BIK:        pd.BIK,
		BankName:   pd.BankName,
	}

	if pd.BankCardNumber != "" {
		resp.BankCardNumber = domain.MaskCardNumber(pd.BankCardNumber)
	}
	if pd.CardHolderName != "" {
		resp.CardHolderName = pd.CardHolderName
	}
	if pd.BankAccount != "" {
		resp.BankAccount = domain.MaskBankAccount(pd.BankAccount)
	}
	if pd.INN != "" {
		resp.INN = pd.INN
	}
	if pd.CorrespondentAccount != "" {
		resp.CorrespondentAccount = pd.CorrespondentAccount
	}

	return resp
}

// SetPaymentDetails godoc
//
//	@Summary		Set owner payment details
//	@Description	Create or update payment details for the current owner (required for payouts)
//	@Tags			payment-details
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		setPaymentDetailsRequest	true	"Payment details"
//	@Success		200		{object}	APIResponse{data=paymentDetailsResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Router			/my/payment-details [put]
func (h *PaymentDetailsHandler) SetPaymentDetails(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req setPaymentDetailsRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	entityType := domain.KYCEntityType(req.EntityType)
	if !entityType.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid entity type")
		return
	}

	details, err := h.paymentDetailsSvc.Set(r.Context(), userID, service.SetPaymentDetailsInput{
		EntityType:           entityType,
		BankCardNumber:       req.BankCardNumber,
		CardHolderName:       req.CardHolderName,
		BankAccount:          req.BankAccount,
		BIK:                  req.BIK,
		INN:                  req.INN,
		CorrespondentAccount: req.CorrespondentAccount,
		BankName:             req.BankName,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPaymentDetailsResponse(details))
}

// GetPaymentDetails godoc
//
//	@Summary		Get owner payment details
//	@Description	Get current payment details for the authenticated owner (sensitive data is masked)
//	@Tags			payment-details
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=paymentDetailsResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/payment-details [get]
func (h *PaymentDetailsHandler) GetPaymentDetails(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	details, err := h.paymentDetailsSvc.Get(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPaymentDetailsResponse(details))
}
