package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/baas"
)

type PaymentRepo struct {
	mu      sync.RWMutex
	pix     map[string]*baas.PixCharge
	boletos map[string]*baas.BoletoCharge
}

func NewPaymentRepo() *PaymentRepo {
	return &PaymentRepo{
		pix:     make(map[string]*baas.PixCharge),
		boletos: make(map[string]*baas.BoletoCharge),
	}
}

func (r *PaymentRepo) SavePixCharge(_ context.Context, charge *baas.PixCharge) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if charge.ID == "" {
		charge.ID = uuid.New().String()
	}
	r.pix[charge.TxID] = charge
	return nil
}

func (r *PaymentRepo) UpdatePixStatus(_ context.Context, _, txID string, status baas.PaymentStatus, paidAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.pix[txID]
	if !ok {
		return fmt.Errorf("PIX %s nao encontrado", txID)
	}
	c.Status = status
	c.PaidAt = paidAt
	return nil
}

func (r *PaymentRepo) SaveBoleto(_ context.Context, boleto *baas.BoletoCharge) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if boleto.ID == "" {
		boleto.ID = uuid.New().String()
	}
	r.boletos[boleto.OurNumber] = boleto
	return nil
}

func (r *PaymentRepo) UpdateBoletoStatus(_ context.Context, _, ourNumber string, status baas.PaymentStatus, paidAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.boletos[ourNumber]
	if !ok {
		return fmt.Errorf("boleto %s nao encontrado", ourNumber)
	}
	b.Status = status
	b.PaidAt = paidAt
	return nil
}
