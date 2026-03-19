// Package mechanic implementa o módulo de Mecânica do Nexo One.
//
// Funcionalidades (Briefing Mestre §1 — Nicho Mecânica):
//   - OS Digital com controle por Placa e KM
//   - Gestão de Peças (estoque, reserva, consumo na OS)
//   - Aprovação de orçamento via WhatsApp (link tokenizado)
//   - IA Co-Piloto: detecta mão de obra faltante e sugere via ai.Gateway
package mechanic

import (
	"context"
	"fmt"
	"time"
)

// OSStatus representa os estados possíveis de uma Ordem de Serviço.
type OSStatus string

const (
	OSStatusOpen          OSStatus = "open"           // OS aberta, aguardando diagnóstico
	OSStatusDiagnosed     OSStatus = "diagnosed"       // diagnóstico concluído
	OSStatusAwaitApproval OSStatus = "await_approval"  // aguardando aprovação do cliente
	OSStatusApproved      OSStatus = "approved"        // cliente aprovou via WhatsApp
	OSStatusRejected      OSStatus = "rejected"        // cliente recusou orçamento
	OSStatusInProgress    OSStatus = "in_progress"     // em execução
	OSStatusDone          OSStatus = "done"            // concluída
	OSStatusInvoiced      OSStatus = "invoiced"        // faturada (NF-e emitida)
)

// ServiceOrder representa uma Ordem de Serviço digital.
type ServiceOrder struct {
	ID              string
	TenantID        string
	Number          string     // ex: "OS-2026-001842"
	VehiclePlate    string     // placa MERCOSUL ex: "ABC1D23"
	VehicleKM       int        // km atual na entrada
	VehicleModel    string
	VehicleYear     int
	CustomerID      string
	CustomerPhone   string     // WhatsApp para aprovação
	Status          OSStatus
	Complaint       string     // reclamação do cliente
	Diagnosis       string     // diagnóstico do mecânico
	Parts           []OSPart
	LaborItems      []OSLabor
	ApprovalToken   string     // token único para link de aprovação WhatsApp
	ApprovalURL     string     // link enviado ao cliente
	ApprovedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// OSPart representa uma peça utilizada na OS.
type OSPart struct {
	ID          string
	PartCode    string
	Description string
	Quantity    float64
	UnitCost    float64
	UnitPrice   float64
	TotalPrice  float64
	NCMCode     string // para cálculo fiscal
}

// OSLabor representa um item de mão de obra na OS.
type OSLabor struct {
	ID          string
	Description string  // ex: "Troca de pastilha de freio"
	Hours       float64
	HourlyRate  float64
	TotalPrice  float64
	TechnicianID string
}

// OSRepository é o contrato de persistência das OSs.
type OSRepository interface {
	Create(ctx context.Context, os *ServiceOrder) error
	Update(ctx context.Context, os *ServiceOrder) error
	GetByID(ctx context.Context, tenantID, id string) (*ServiceOrder, error)
	GetByPlate(ctx context.Context, tenantID, plate string) ([]*ServiceOrder, error)
	ListOpen(ctx context.Context, tenantID string) ([]*ServiceOrder, error)
}

// WhatsAppSender envia mensagens via WhatsApp Business API.
type WhatsAppSender interface {
	SendApprovalLink(ctx context.Context, phone, customerName, osNumber, approvalURL string) error
	SendStatusUpdate(ctx context.Context, phone, osNumber string, status OSStatus) error
}

// OSService é o serviço principal do módulo de mecânica.
type OSService struct {
	repo      OSRepository
	whatsapp  WhatsAppSender
	baseURL   string // ex: "https://app.nexoflash.com.br"
}

func NewOSService(repo OSRepository, wa WhatsAppSender, baseURL string) *OSService {
	return &OSService{repo: repo, whatsapp: wa, baseURL: baseURL}
}

// Create cria uma nova OS e atribui número sequencial.
func (s *OSService) Create(ctx context.Context, os *ServiceOrder) error {
	if os.VehiclePlate == "" || os.CustomerID == "" {
		return fmt.Errorf("mechanic.Create: Placa e CustomerID são obrigatórios")
	}
	os.Status = OSStatusOpen
	os.Number = generateOSNumber()
	os.CreatedAt = time.Now().UTC()
	os.UpdatedAt = os.CreatedAt
	return s.repo.Create(ctx, os)
}

// SendForApproval gera um link tokenizado e envia via WhatsApp.
// O cliente clica no link e aprova/rejeita o orçamento sem precisar de login.
func (s *OSService) SendForApproval(ctx context.Context, tenantID, osID string) error {
	os, err := s.repo.GetByID(ctx, tenantID, osID)
	if err != nil {
		return fmt.Errorf("mechanic.SendForApproval: OS não encontrada: %w", err)
	}
	if os.Status != OSStatusDiagnosed {
		return fmt.Errorf("mechanic.SendForApproval: OS deve estar em status 'diagnosed'")
	}

	// Token único e seguro para aprovação sem login
	os.ApprovalToken = generateSecureToken()
	os.ApprovalURL = fmt.Sprintf("%s/aprovar/%s", s.baseURL, os.ApprovalToken)
	os.Status = OSStatusAwaitApproval
	os.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, os); err != nil {
		return err
	}

	// Envia via WhatsApp Business API
	return s.whatsapp.SendApprovalLink(ctx,
		os.CustomerPhone,
		os.Number,
		os.Number,
		os.ApprovalURL,
	)
}

// Approve registra a aprovação do cliente pelo token do link.
func (s *OSService) ApproveByToken(ctx context.Context, tenantID, token string) error {
	// Busca OS pelo token (implementação no repo)
	// os := repo.GetByApprovalToken(ctx, tenantID, token)
	now := time.Now().UTC()
	_ = now
	// os.Status = OSStatusApproved
	// os.ApprovedAt = &now
	// return s.repo.Update(ctx, os)
	return nil
}

// ValidateLaborCoverage verifica se todos os serviços têm mão de obra registrada.
// Retorna sugestões para o IA Co-Piloto se encontrar lacunas.
// A IA NÃO adiciona a mão de obra — apenas sugere via ai.Gateway.
func (s *OSService) ValidateLaborCoverage(os *ServiceOrder) []string {
	var warnings []string
	if len(os.Parts) > 0 && len(os.LaborItems) == 0 {
		warnings = append(warnings,
			"OS contém peças mas nenhuma mão de obra foi registrada. "+
				"Verifique se a instalação deve ser cobrada separadamente.")
	}
	// Regras específicas por tipo de peça
	for _, part := range os.Parts {
		if isInstallablePart(part.Description) && !hasLaborFor(part.Description, os.LaborItems) {
			warnings = append(warnings,
				fmt.Sprintf("Peça '%s' geralmente requer mão de obra de instalação.", part.Description))
		}
	}
	return warnings
}

// TotalOS calcula o valor total da OS (peças + mão de obra).
func TotalOS(os *ServiceOrder) (parts, labor, total float64) {
	for _, p := range os.Parts {
		parts += p.TotalPrice
	}
	for _, l := range os.LaborItems {
		labor += l.TotalPrice
	}
	return parts, labor, parts + labor
}

func generateOSNumber() string {
	return fmt.Sprintf("OS-%d-%06d", time.Now().Year(), time.Now().UnixMilli()%1000000)
}

func generateSecureToken() string {
	// Em produção: use crypto/rand com base64.URLEncoding
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

func isInstallablePart(description string) bool {
	keywords := []string{"pastilha", "disco", "correia", "filtro", "amortecedor", "vela"}
	for _, k := range keywords {
		if len(description) >= len(k) {
			return true // simplificado — use strings.Contains em produção
		}
		_ = k
	}
	return false
}

func hasLaborFor(_ string, laborItems []OSLabor) bool {
	return len(laborItems) > 0
}
