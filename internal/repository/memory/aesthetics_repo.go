package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/modules/aesthetics"
)

type AestheticsRepo struct {
	mu   sync.RWMutex
	apts map[string]*aesthetics.Appointment
}

func NewAestheticsRepo() *AestheticsRepo {
	return &AestheticsRepo{apts: make(map[string]*aesthetics.Appointment)}
}

func (r *AestheticsRepo) Create(_ context.Context, apt *aesthetics.Appointment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	apt.ID = uuid.New().String()
	apt.CreatedAt = time.Now().UTC()
	r.apts[apt.ID] = apt
	return nil
}

func (r *AestheticsRepo) Update(_ context.Context, apt *aesthetics.Appointment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.apts[apt.ID]; !ok {
		return fmt.Errorf("agendamento %s nao encontrado", apt.ID)
	}
	r.apts[apt.ID] = apt
	return nil
}

func (r *AestheticsRepo) GetByID(_ context.Context, tenantID, id string) (*aesthetics.Appointment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.apts[id]
	if !ok || a.TenantID != tenantID {
		return nil, fmt.Errorf("agendamento %s nao encontrado", id)
	}
	return a, nil
}

func (r *AestheticsRepo) FindConflicts(_ context.Context, tenantID, professionalID string, start, end time.Time, excludeID string) ([]*aesthetics.Appointment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var conflicts []*aesthetics.Appointment
	for _, a := range r.apts {
		if a.TenantID != tenantID || a.ProfessionalID != professionalID {
			continue
		}
		if a.ID == excludeID {
			continue
		}
		if a.Status == aesthetics.AppointmentCancelled || a.Status == aesthetics.AppointmentNoShow {
			continue
		}
		if a.StartTime.Before(end) && a.EndTime.After(start) {
			conflicts = append(conflicts, a)
		}
	}
	return conflicts, nil
}

func (r *AestheticsRepo) ListByProfessional(_ context.Context, tenantID, professionalID string, date time.Time) ([]*aesthetics.Appointment, error) {
	return r.listByDate(tenantID, date, professionalID)
}

func (r *AestheticsRepo) ListByDate(_ context.Context, tenantID string, date time.Time) ([]*aesthetics.Appointment, error) {
	return r.listByDate(tenantID, date, "")
}

func (r *AestheticsRepo) listByDate(tenantID string, date time.Time, professionalID string) ([]*aesthetics.Appointment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	var result []*aesthetics.Appointment
	for _, a := range r.apts {
		if a.TenantID != tenantID {
			continue
		}
		if professionalID != "" && a.ProfessionalID != professionalID {
			continue
		}
		if a.StartTime.After(startOfDay) && a.StartTime.Before(endOfDay) {
			result = append(result, a)
		}
	}
	return result, nil
}
