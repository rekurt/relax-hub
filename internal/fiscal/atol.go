package fiscal

import (
	"context"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/logger"
)

// ATOLProvider is a placeholder implementation of FiscalProvider for ATOL Online.
// It logs receipt requests but does not call the real ATOL API.
type ATOLProvider struct {
	login     string
	password  string
	groupCode string
	logger    *logger.Logger
}

// NewATOLProvider creates a new ATOL placeholder provider.
func NewATOLProvider(login, password, groupCode string, log *logger.Logger) *ATOLProvider {
	return &ATOLProvider{
		login:     login,
		password:  password,
		groupCode: groupCode,
		logger:    log,
	}
}

// CreateReceipt logs the receipt request and returns a mock receipt.
func (p *ATOLProvider) CreateReceipt(_ context.Context, req ReceiptRequest) (*Receipt, error) {
	taxSystem := req.TaxSystem
	if taxSystem == "" {
		taxSystem = TaxSystemDefault
	}
	p.logger.Info("ATOL receipt created (placeholder)",
		"type", string(req.Type),
		"amount", req.Amount,
		"email", req.Email,
		"phone", req.Phone,
		"items_count", len(req.Items),
		"tax_system", string(taxSystem),
	)
	return &Receipt{
		ID:     uuid.New().String(),
		Status: "pending",
	}, nil
}
