// Package enterprise — serviço de gestão enterprise.
package enterprise

import (
	"context"
	"time"
)

type Plan struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // starter, professional, enterprise
	MaxUsers    int       `json:"max_users"`
	MaxTenants  int       `json:"max_tenants"`
	Features    []string  `json:"features"`
	PriceMonth  float64   `json:"price_month"`
	PriceYear   float64   `json:"price_year"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type License struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	PlanID      string    `json:"plan_id"`
	Status      string    `json:"status"` // active, expired, suspended
	ValidUntil  time.Time `json:"valid_until"`
	MaxUsers    int       `json:"max_users"`
	UsedUsers   int       `json:"used_users"`
	CreatedAt   time.Time `json:"created_at"`
}

type Usage struct {
	TenantID     string    `json:"tenant_id"`
	Month        string    `json:"month"`
	APIRequests  int64     `json:"api_requests"`
	Storage      int64     `json:"storage_bytes"`
	NFesEmitted  int       `json:"nfes_emitted"`
	UsersActive  int       `json:"users_active"`
	RecordedAt   time.Time `json:"recorded_at"`
}

type Repository interface {
	ListPlans(ctx context.Context) ([]*Plan, error)
	GetPlan(ctx context.Context, id string) (*Plan, error)
	CreatePlan(ctx context.Context, plan *Plan) error
	UpdatePlan(ctx context.Context, plan *Plan) error
	
	GetLicense(ctx context.Context, tenantID string) (*License, error)
	CreateLicense(ctx context.Context, license *License) error
	UpdateLicense(ctx context.Context, license *License) error
	
	GetUsage(ctx context.Context, tenantID, month string) (*Usage, error)
	RecordUsage(ctx context.Context, usage *Usage) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListPlans(ctx context.Context) ([]*Plan, error) {
	return s.repo.ListPlans(ctx)
}

func (s *Service) GetPlan(ctx context.Context, id string) (*Plan, error) {
	return s.repo.GetPlan(ctx, id)
}

func (s *Service) CreatePlan(ctx context.Context, plan *Plan) error {
	plan.CreatedAt = time.Now()
	plan.IsActive = true
	return s.repo.CreatePlan(ctx, plan)
}

func (s *Service) UpdatePlan(ctx context.Context, plan *Plan) error {
	return s.repo.UpdatePlan(ctx, plan)
}

func (s *Service) GetLicense(ctx context.Context, tenantID string) (*License, error) {
	return s.repo.GetLicense(ctx, tenantID)
}

func (s *Service) ActivateLicense(ctx context.Context, tenantID, planID string, months int) (*License, error) {
	plan, err := s.repo.GetPlan(ctx, planID)
	if err != nil {
		return nil, err
	}
	
	license := &License{
		ID:         tenantID + "-license",
		TenantID:   tenantID,
		PlanID:     planID,
		Status:     "active",
		ValidUntil: time.Now().AddDate(0, months, 0),
		MaxUsers:   plan.MaxUsers,
		UsedUsers:  0,
		CreatedAt:  time.Now(),
	}
	
	if err := s.repo.CreateLicense(ctx, license); err != nil {
		return nil, err
	}
	return license, nil
}

func (s *Service) GetUsage(ctx context.Context, tenantID, month string) (*Usage, error) {
	return s.repo.GetUsage(ctx, tenantID, month)
}

func (s *Service) RecordUsage(ctx context.Context, usage *Usage) error {
	usage.RecordedAt = time.Now()
	return s.repo.RecordUsage(ctx, usage)
}
