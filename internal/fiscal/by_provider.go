package fiscal

import (
	"context"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/logger"
)

// BYProvider is a FiscalProvider for Belarus region.
// It logs receipt data for manual processing until a specific BY fiscal
// operator (e.g. СККО) is integrated.
type BYProvider struct {
	logger *logger.Logger
}

func NewBYProvider(log *logger.Logger) *BYProvider {
	return &BYProvider{logger: log}
}

func (p *BYProvider) CreateReceipt(_ context.Context, req ReceiptRequest) (*Receipt, error) {
	receiptID := uuid.New().String()

	p.logger.Info("BY fiscal receipt recorded (pending integration)",
		"receipt_id", receiptID,
		"type", string(req.Type),
		"amount", req.Amount,
		"email", req.Email,
		"phone", req.Phone,
		"items_count", len(req.Items),
	)

	return &Receipt{
		ID:     receiptID,
		Status: "pending_manual",
	}, nil
}
