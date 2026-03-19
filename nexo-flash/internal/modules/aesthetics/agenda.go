// Package aesthetics implementa o módulo de Estética do Nexo Flash.
//
// Funcionalidades (Briefing Mestre §1 — Nicho Estética):
//   - Agenda com trava de conflito (sem double-booking)
//   - Split de Pagamento para Salão Parceiro
//     (ex: 60% profissional autônomo + 40% salão)
package aesthetics

import (
	"context"
	"fmt"
	"time"
)

// AppointmentStatus estados de um agendamento.
type AppointmentStatus string

const (
	AppointmentScheduled  AppointmentStatus = "scheduled"
	AppointmentConfirmed  AppointmentStatus = "confirmed"
	AppointmentInProgress AppointmentStatus = "in_progress"
	AppointmentDone       AppointmentStatus = "done"
	AppointmentCancelled  AppointmentStatus = "cancelled"
	AppointmentNoShow     AppointmentStatus = "no_show"
)

// Appointment representa um agendamento na agenda de estética.
type Appointment struct {
	ID            string
	TenantID      string
	ProfessionalID string
	CustomerID    string
	CustomerName  string
	CustomerPhone string
	ServiceID     string
	ServiceName   string
	ServicePrice  float64
	StartTime     time.Time
	EndTime       time.Time    // calculado: StartTime + duração do serviço
	DurationMin   int
	Status        AppointmentStatus
	Notes         string
	// Split de pagamento
	SplitEnabled  bool
	SplitRules    []SplitRule
	CreatedAt     time.Time
}

// SplitRule define como dividir o pagamento entre profissional e salão.
type SplitRule struct {
	RecipientID   string  // ID do profissional ou do salão
	RecipientType string  // "professional" | "salon"
	Percentage    float64 // ex: 60.0 = 60%
	FixedAmount   float64 // alternativa: valor fixo
	BaaSAccountID string  // conta BaaS para recebimento automático
}

// SplitResult é o resultado do cálculo de split após uma venda.
type SplitResult struct {
	TotalAmount   float64
	Distributions []SplitDistribution
}

// SplitDistribution é o valor destinado a cada parte.
type SplitDistribution struct {
	RecipientID   string
	RecipientType string
	Amount        float64
	BaaSAccountID string
}

// ConflictError é retornado quando há conflito de horário na agenda.
type ConflictError struct {
	ProfessionalID  string
	ConflictingID   string
	ConflictingTime time.Time
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflito de agenda: profissional %s já tem compromisso às %s (OS: %s)",
		e.ProfessionalID, e.ConflictingTime.Format("15:04"), e.ConflictingID)
}

// AgendaRepository persiste agendamentos.
type AgendaRepository interface {
	Create(ctx context.Context, apt *Appointment) error
	Update(ctx context.Context, apt *Appointment) error
	GetByID(ctx context.Context, tenantID, id string) (*Appointment, error)
	// FindConflicts é o coração da trava de conflito:
	// retorna agendamentos que se sobrepõem ao intervalo dado para o profissional.
	FindConflicts(ctx context.Context, tenantID, professionalID string, start, end time.Time, excludeID string) ([]*Appointment, error)
	ListByProfessional(ctx context.Context, tenantID, professionalID string, date time.Time) ([]*Appointment, error)
	ListByDate(ctx context.Context, tenantID string, date time.Time) ([]*Appointment, error)
}

// AgendaService gerencia a agenda com trava de conflito e split de pagamento.
type AgendaService struct {
	repo AgendaRepository
}

func NewAgendaService(r AgendaRepository) *AgendaService {
	return &AgendaService{repo: r}
}

// Book cria um agendamento com verificação obrigatória de conflito.
// Retorna *ConflictError se o profissional já tiver agendamento no horário.
// A trava é implementada no banco com FOR UPDATE — segura para concorrência.
func (s *AgendaService) Book(ctx context.Context, apt *Appointment) error {
	if apt.StartTime.IsZero() || apt.DurationMin <= 0 {
		return fmt.Errorf("aesthetics.Book: StartTime e DurationMin são obrigatórios")
	}

	// Calcular horário de término
	apt.EndTime = apt.StartTime.Add(time.Duration(apt.DurationMin) * time.Minute)

	// Verificar conflito ANTES de salvar — trava de conflito
	conflicts, err := s.repo.FindConflicts(ctx,
		apt.TenantID, apt.ProfessionalID,
		apt.StartTime, apt.EndTime, "")
	if err != nil {
		return fmt.Errorf("aesthetics.Book: verificação de conflito falhou: %w", err)
	}
	if len(conflicts) > 0 {
		return &ConflictError{
			ProfessionalID:  apt.ProfessionalID,
			ConflictingID:   conflicts[0].ID,
			ConflictingTime: conflicts[0].StartTime,
		}
	}

	apt.Status = AppointmentScheduled
	apt.CreatedAt = time.Now().UTC()
	return s.repo.Create(ctx, apt)
}

// Reschedule move um agendamento, re-verificando conflito no novo horário.
func (s *AgendaService) Reschedule(ctx context.Context, tenantID, aptID string, newStart time.Time) error {
	apt, err := s.repo.GetByID(ctx, tenantID, aptID)
	if err != nil {
		return err
	}

	newEnd := newStart.Add(time.Duration(apt.DurationMin) * time.Minute)
	conflicts, err := s.repo.FindConflicts(ctx, tenantID, apt.ProfessionalID, newStart, newEnd, aptID)
	if err != nil {
		return err
	}
	if len(conflicts) > 0 {
		return &ConflictError{
			ProfessionalID:  apt.ProfessionalID,
			ConflictingID:   conflicts[0].ID,
			ConflictingTime: conflicts[0].StartTime,
		}
	}

	apt.StartTime = newStart
	apt.EndTime = newEnd
	return s.repo.Update(ctx, apt)
}

// CalculateSplit calcula a distribuição de pagamento entre profissional e salão.
// O repasse via BaaS é disparado como evento — não executado aqui diretamente.
func CalculateSplit(apt *Appointment) (*SplitResult, error) {
	if !apt.SplitEnabled || len(apt.SplitRules) == 0 {
		return &SplitResult{
			TotalAmount: apt.ServicePrice,
			Distributions: []SplitDistribution{
				{RecipientType: "salon", Amount: apt.ServicePrice},
			},
		}, nil
	}

	// Valida: a soma dos percentuais deve ser exatamente 100%
	var totalPct float64
	for _, r := range apt.SplitRules {
		totalPct += r.Percentage
	}
	if totalPct != 100 {
		return nil, fmt.Errorf("aesthetics.CalculateSplit: soma dos percentuais deve ser 100%%, recebido %.1f%%", totalPct)
	}

	var distributions []SplitDistribution
	for _, rule := range apt.SplitRules {
		amount := apt.ServicePrice * (rule.Percentage / 100)
		distributions = append(distributions, SplitDistribution{
			RecipientID:   rule.RecipientID,
			RecipientType: rule.RecipientType,
			Amount:        roundPrice(amount),
			BaaSAccountID: rule.BaaSAccountID,
		})
	}

	return &SplitResult{
		TotalAmount:   apt.ServicePrice,
		Distributions: distributions,
	}, nil
}

func roundPrice(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
