package mock

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type CertificateRepo struct {
	mu     sync.RWMutex
	certs  map[uuid.UUID]*domain.GiftCertificate
	usages map[uuid.UUID]*domain.CertificateUsage
}

func NewCertificateRepo() repository.GiftCertificateRepository {
	return &CertificateRepo{
		certs:  make(map[uuid.UUID]*domain.GiftCertificate),
		usages: make(map[uuid.UUID]*domain.CertificateUsage),
	}
}

func (r *CertificateRepo) Create(_ context.Context, cert *domain.GiftCertificate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cert.ID == uuid.Nil {
		cert.ID = uuid.New()
	}

	// Check unique code
	for _, c := range r.certs {
		if c.Code == cert.Code {
			return domain.ErrAlreadyExists
		}
	}

	if cert.Status == "" {
		cert.Status = domain.CertificateStatusActive
	}
	if cert.CreatedAt.IsZero() {
		cert.CreatedAt = time.Now()
	}

	cp := *cert
	r.certs[cert.ID] = &cp
	return nil
}

func (r *CertificateRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.GiftCertificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cert, ok := r.certs[id]
	if !ok {
		return nil, domain.ErrCertificateNotFound
	}
	cp := *cert
	return &cp, nil
}

func (r *CertificateRepo) GetByCode(_ context.Context, code string) (*domain.GiftCertificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, cert := range r.certs {
		if cert.Code == code {
			cp := *cert
			return &cp, nil
		}
	}
	return nil, domain.ErrCertificateNotFound
}

func (r *CertificateRepo) UpdateBalance(_ context.Context, id uuid.UUID, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cert, ok := r.certs[id]
	if !ok {
		return domain.ErrCertificateInsufficientBalance
	}

	if cert.Status != domain.CertificateStatusActive || cert.Balance < amount {
		return domain.ErrCertificateInsufficientBalance
	}

	cert.Balance -= amount
	if cert.Balance == 0 {
		cert.Status = domain.CertificateStatusUsed
	}
	return nil
}

func (r *CertificateRepo) ListByUser(_ context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GiftCertificate], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var filtered []domain.GiftCertificate
	for _, cert := range r.certs {
		if (cert.RedeemedByID != nil && *cert.RedeemedByID == userID) ||
			(cert.PurchaserID != nil && *cert.PurchaserID == userID) {
			cp := *cert
			filtered = append(filtered, cp)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	totalCount := int64(len(filtered))
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset > int(totalCount) {
		offset = int(totalCount)
	}
	if end > int(totalCount) {
		end = int(totalCount)
	}

	return &domain.PaginatedResult[domain.GiftCertificate]{
		Items:      filtered[offset:end],
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (r *CertificateRepo) Redeem(_ context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cert, ok := r.certs[id]
	if !ok {
		return domain.ErrCertificateNotFound
	}

	if cert.RedeemedByID != nil || cert.Status != domain.CertificateStatusActive {
		return domain.ErrCertificateNotFound
	}

	cert.RedeemedByID = &userID
	return nil
}

func (r *CertificateRepo) CreateUsage(_ context.Context, usage *domain.CertificateUsage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}
	if usage.UsedAt.IsZero() {
		usage.UsedAt = time.Now()
	}

	cp := *usage
	r.usages[usage.ID] = &cp
	return nil
}
