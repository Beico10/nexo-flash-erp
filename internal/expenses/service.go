// Package expenses implementa o módulo de despesas com leitura de QR Code.
//
// Fluxo:
//  1. Usuário aponta câmera para QR Code da NFC-e/NF-e
//  2. Sistema extrai URL e chave de acesso
//  3. Consulta dados na SEFAZ (scraping da página de consulta)
//  4. Registra como despesa com categorização automática
//  5. Calcula crédito de imposto (IBS/CBS)
//
// Suporta:
//  - NFC-e (Nota Fiscal do Consumidor Eletrônica)
//  - NF-e (Nota Fiscal Eletrônica)
//  - SAT (Sistema Autenticador e Transmissor - SP)
//  - CT-e (Conhecimento de Transporte - para logística)
package expenses

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ════════════════════════════════════════════════════════════
// TIPOS
// ════════════════════════════════════════════════════════════

// Expense representa uma despesa do negócio.
type Expense struct {
	ID             string
	TenantID       string
	Source         string // qrcode, xml_upload, manual, recurring
	
	// NF-e
	NFeKey         string
	NFeNumber      string
	NFeSeries      string
	NFeType        string // nfe, nfce, cte, sat
	NFeURL         string
	
	// Fornecedor
	SupplierCNPJ   string
	SupplierName   string
	SupplierIE     string
	
	// Valores
	TotalProducts  float64
	TotalDiscount  float64
	TotalShipping  float64
	TotalAmount    float64
	
	// Impostos (crédito)
	ICMSAmount     float64
	IPIAmount      float64
	PISAmount      float64
	COFINSAmount   float64
	IBSCredit      float64
	CBSCredit      float64
	
	// Categorização
	Category       string
	Subcategory    string
	Tags           []string
	
	// Referência
	ModuleRef      string
	ReferenceType  string
	ReferenceID    string
	
	// Pagamento
	PaymentMethod  string
	Paid           bool
	DueDate        *time.Time
	
	// Datas
	IssueDate      time.Time
	RegisteredAt   time.Time
	RegisteredBy   string
	
	// Status
	Status         string
	Notes          string
	
	// Itens
	Items          []ExpenseItem
}

// ExpenseItem representa um item da despesa.
type ExpenseItem struct {
	ID           string
	ItemOrder    int
	ProductCode  string
	EAN          string
	Description  string
	NCM          string
	CFOP         string
	Quantity     float64
	Unit         string
	UnitPrice    float64
	Discount     float64
	TotalPrice   float64
	ICMSBase     float64
	ICMSRate     float64
	ICMSAmount   float64
	IPIAmount    float64
	AutoCategory string
}

// QRCodeResult resultado do parse do QR Code.
type QRCodeResult struct {
	Type       string // nfce, nfe, sat, cte, pix, unknown
	URL        string
	AccessKey  string
	UF         string
	IsValid    bool
	RawContent string
}

// ExpenseSummary resumo de despesas.
type ExpenseSummary struct {
	Month         time.Time
	Category      string
	Count         int
	Total         float64
	IBSCredit     float64
	CBSCredit     float64
	ICMSPaid      float64
}

// TaxReport relatório para IR/imposto.
type TaxReport struct {
	Year          int
	Month         int
	CategoryCode  string
	CategoryName  string
	TaxDeductible bool
	Total         float64
	TaxCredit     float64
	DocCount      int
}

// ════════════════════════════════════════════════════════════
// ERROS
// ════════════════════════════════════════════════════════════

var (
	ErrInvalidQRCode      = errors.New("QR Code inválido ou não reconhecido")
	ErrDuplicateExpense   = errors.New("esta nota já foi registrada")
	ErrSEFAZUnavailable   = errors.New("SEFAZ indisponível, tente novamente")
	ErrExpenseNotFound    = errors.New("despesa não encontrada")
)

// ════════════════════════════════════════════════════════════
// REPOSITÓRIO
// ════════════════════════════════════════════════════════════

type ExpenseRepository interface {
	Create(ctx context.Context, e *Expense) error
	GetByID(ctx context.Context, tenantID, id string) (*Expense, error)
	GetByNFeKey(ctx context.Context, tenantID, nfeKey string) (*Expense, error)
	List(ctx context.Context, tenantID string, filter ExpenseFilter) ([]*Expense, error)
	Update(ctx context.Context, e *Expense) error
	Delete(ctx context.Context, tenantID, id string) error
	
	// Itens
	CreateItems(ctx context.Context, tenantID, expenseID string, items []ExpenseItem) error
	GetItems(ctx context.Context, tenantID, expenseID string) ([]ExpenseItem, error)
	
	// Categorias
	GetCategories(ctx context.Context, tenantID string) ([]ExpenseCategory, error)
	AutoCategorize(ctx context.Context, tenantID, ncm string) (string, error)
	
	// Relatórios
	GetSummary(ctx context.Context, tenantID string, from, to time.Time) ([]ExpenseSummary, error)
	GetTaxReport(ctx context.Context, tenantID string, year int) ([]TaxReport, error)
	
	// Log de scans
	LogScan(ctx context.Context, tenantID, userID, content, qrType string, success bool, expenseID, errorMsg string) error
}

// ExpenseFilter filtros para listagem.
type ExpenseFilter struct {
	Category    string
	SupplierCNPJ string
	DateFrom    *time.Time
	DateTo      *time.Time
	Status      string
	MinAmount   *float64
	MaxAmount   *float64
	Limit       int
	Offset      int
}

// ExpenseCategory categoria de despesa.
type ExpenseCategory struct {
	ID           string
	Code         string
	Name         string
	Icon         string
	Color        string
	TaxDeductible bool
	NCMPatterns  []string
}

// ════════════════════════════════════════════════════════════
// CONSULTA SEFAZ (Interface)
// ════════════════════════════════════════════════════════════

// SEFAZConsulta interface para consultar NF-e na SEFAZ.
type SEFAZConsulta interface {
	ConsultarNFCe(ctx context.Context, url string) (*Expense, error)
	ConsultarNFe(ctx context.Context, chave string) (*Expense, error)
	ConsultarSAT(ctx context.Context, chave string) (*Expense, error)
}

// ════════════════════════════════════════════════════════════
// SERVIÇO
// ════════════════════════════════════════════════════════════

type Service struct {
	repo   ExpenseRepository
	sefaz  SEFAZConsulta
}

func NewService(repo ExpenseRepository, sefaz SEFAZConsulta) *Service {
	return &Service{repo: repo, sefaz: sefaz}
}

// ────────────────────────────────────────────────────────────
// LEITURA DE QR CODE
// ────────────────────────────────────────────────────────────

// ParseQRCode analisa o conteúdo do QR Code e identifica o tipo.
func (s *Service) ParseQRCode(content string) (*QRCodeResult, error) {
	content = strings.TrimSpace(content)
	result := &QRCodeResult{RawContent: content}
	
	// Tentar identificar o tipo
	
	// 1. NFC-e (URL de consulta)
	if strings.Contains(content, "nfce.fazenda") || strings.Contains(content, "nfce.sefaz") {
		result.Type = "nfce"
		result.URL = content
		result.AccessKey = extractAccessKeyFromURL(content)
		result.UF = extractUFFromURL(content)
		result.IsValid = len(result.AccessKey) == 44
		return result, nil
	}
	
	// 2. NF-e (URL de consulta)
	if strings.Contains(content, "nfe.fazenda") || strings.Contains(content, "portalnfe") {
		result.Type = "nfe"
		result.URL = content
		result.AccessKey = extractAccessKeyFromURL(content)
		result.UF = extractUFFromURL(content)
		result.IsValid = len(result.AccessKey) == 44
		return result, nil
	}
	
	// 3. SAT-CF-e (São Paulo)
	if strings.Contains(content, "sat.sef.sp.gov.br") || strings.Contains(content, "satsp") {
		result.Type = "sat"
		result.URL = content
		result.AccessKey = extractAccessKeyFromURL(content)
		result.UF = "SP"
		result.IsValid = len(result.AccessKey) == 44
		return result, nil
	}
	
	// 4. CT-e
	if strings.Contains(content, "cte.fazenda") {
		result.Type = "cte"
		result.URL = content
		result.AccessKey = extractAccessKeyFromURL(content)
		result.UF = extractUFFromURL(content)
		result.IsValid = len(result.AccessKey) == 44
		return result, nil
	}
	
	// 5. Chave de acesso direta (44 dígitos numéricos)
	if regexp.MustCompile(`^\d{44}$`).MatchString(content) {
		result.Type = "nfe" // Assume NF-e
		result.AccessKey = content
		result.UF = content[0:2] // Primeiros 2 dígitos = UF
		result.IsValid = true
		return result, nil
	}
	
	// 6. PIX (não é despesa, mas podemos informar)
	if strings.HasPrefix(content, "00020126") || strings.Contains(content, "br.gov.bcb.pix") {
		result.Type = "pix"
		result.IsValid = false // Não é uma despesa
		return result, nil
	}
	
	// Não reconhecido
	result.Type = "unknown"
	result.IsValid = false
	return result, ErrInvalidQRCode
}

// ProcessQRCode processa o QR Code e registra a despesa.
func (s *Service) ProcessQRCode(ctx context.Context, tenantID, userID, qrContent string) (*Expense, error) {
	// 1. Parse do QR Code
	qr, err := s.ParseQRCode(qrContent)
	if err != nil {
		s.repo.LogScan(ctx, tenantID, userID, qrContent, "unknown", false, "", err.Error())
		return nil, err
	}
	
	if !qr.IsValid {
		s.repo.LogScan(ctx, tenantID, userID, qrContent, qr.Type, false, "", "QR Code inválido")
		return nil, ErrInvalidQRCode
	}
	
	// 2. Verificar duplicidade
	if qr.AccessKey != "" {
		existing, _ := s.repo.GetByNFeKey(ctx, tenantID, qr.AccessKey)
		if existing != nil {
			s.repo.LogScan(ctx, tenantID, userID, qrContent, qr.Type, false, existing.ID, "Duplicado")
			return nil, ErrDuplicateExpense
		}
	}
	
	// 3. Consultar SEFAZ
	var expense *Expense
	switch qr.Type {
	case "nfce", "sat":
		expense, err = s.sefaz.ConsultarNFCe(ctx, qr.URL)
	case "nfe":
		expense, err = s.sefaz.ConsultarNFe(ctx, qr.AccessKey)
	default:
		err = ErrInvalidQRCode
	}
	
	if err != nil {
		s.repo.LogScan(ctx, tenantID, userID, qrContent, qr.Type, false, "", err.Error())
		return nil, err
	}
	
	// 4. Preencher dados do tenant
	expense.TenantID = tenantID
	expense.Source = "qrcode"
	expense.NFeURL = qr.URL
	expense.NFeKey = qr.AccessKey
	expense.NFeType = qr.Type
	expense.RegisteredBy = userID
	expense.RegisteredAt = time.Now()
	expense.Status = "active"
	
	// 5. Categorizar automaticamente
	if len(expense.Items) > 0 {
		// Usar NCM do primeiro item para categorizar
		category, _ := s.repo.AutoCategorize(ctx, tenantID, expense.Items[0].NCM)
		expense.Category = category
	} else {
		expense.Category = "outros"
	}
	
	// 6. Calcular crédito de imposto (IBS/CBS 2026)
	expense.IBSCredit = s.calculateIBSCredit(expense)
	expense.CBSCredit = s.calculateCBSCredit(expense)
	
	// 7. Salvar
	if err := s.repo.Create(ctx, expense); err != nil {
		s.repo.LogScan(ctx, tenantID, userID, qrContent, qr.Type, false, "", err.Error())
		return nil, err
	}
	
	// 8. Salvar itens
	if len(expense.Items) > 0 {
		s.repo.CreateItems(ctx, tenantID, expense.ID, expense.Items)
	}
	
	// 9. Log de sucesso
	s.repo.LogScan(ctx, tenantID, userID, qrContent, qr.Type, true, expense.ID, "")
	
	return expense, nil
}

// ────────────────────────────────────────────────────────────
// CRUD
// ────────────────────────────────────────────────────────────

// CreateManual cria uma despesa manual (sem QR Code).
func (s *Service) CreateManual(ctx context.Context, e *Expense) error {
	e.Source = "manual"
	e.RegisteredAt = time.Now()
	e.Status = "active"
	
	if e.Category == "" {
		e.Category = "outros"
	}
	
	return s.repo.Create(ctx, e)
}

// GetByID busca despesa por ID.
func (s *Service) GetByID(ctx context.Context, tenantID, id string) (*Expense, error) {
	expense, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	
	// Carregar itens
	items, _ := s.repo.GetItems(ctx, tenantID, id)
	expense.Items = items
	
	return expense, nil
}

// List lista despesas com filtros.
func (s *Service) List(ctx context.Context, tenantID string, filter ExpenseFilter) ([]*Expense, error) {
	return s.repo.List(ctx, tenantID, filter)
}

// Update atualiza despesa.
func (s *Service) Update(ctx context.Context, e *Expense) error {
	return s.repo.Update(ctx, e)
}

// Delete marca despesa como cancelada.
func (s *Service) Delete(ctx context.Context, tenantID, id string) error {
	e, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	e.Status = "cancelled"
	return s.repo.Update(ctx, e)
}

// ────────────────────────────────────────────────────────────
// RELATÓRIOS
// ────────────────────────────────────────────────────────────

// GetSummary retorna resumo de despesas por período.
func (s *Service) GetSummary(ctx context.Context, tenantID string, from, to time.Time) ([]ExpenseSummary, error) {
	return s.repo.GetSummary(ctx, tenantID, from, to)
}

// GetTaxReport retorna relatório para IR/imposto.
func (s *Service) GetTaxReport(ctx context.Context, tenantID string, year int) ([]TaxReport, error) {
	return s.repo.GetTaxReport(ctx, tenantID, year)
}

// GetCategories retorna categorias disponíveis.
func (s *Service) GetCategories(ctx context.Context, tenantID string) ([]ExpenseCategory, error) {
	return s.repo.GetCategories(ctx, tenantID)
}

// ────────────────────────────────────────────────────────────
// CÁLCULO DE IMPOSTOS
// ────────────────────────────────────────────────────────────

// calculateIBSCredit calcula crédito de IBS (imposto estadual/municipal 2026).
func (s *Service) calculateIBSCredit(e *Expense) float64 {
	// Simplificação: 9.25% sobre o total (alíquota padrão IBS)
	// Em produção, usar tabela de NCM para alíquota correta
	return e.TotalAmount * 0.0925
}

// calculateCBSCredit calcula crédito de CBS (imposto federal 2026).
func (s *Service) calculateCBSCredit(e *Expense) float64 {
	// Simplificação: 3.75% sobre o total (alíquota padrão CBS)
	return e.TotalAmount * 0.0375
}

// ════════════════════════════════════════════════════════════
// HELPERS
// ════════════════════════════════════════════════════════════

// extractAccessKeyFromURL extrai a chave de acesso de 44 dígitos da URL.
func extractAccessKeyFromURL(rawURL string) string {
	// Padrão: chave de 44 dígitos numéricos
	re := regexp.MustCompile(`\d{44}`)
	matches := re.FindAllString(rawURL, -1)
	if len(matches) > 0 {
		return matches[0]
	}
	
	// Tentar extrair do parâmetro p= ou chNFe=
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	
	// Parâmetro p (NFC-e)
	if p := parsed.Query().Get("p"); p != "" {
		// Formato: chave|versão|ambiente|...
		parts := strings.Split(p, "|")
		if len(parts) > 0 && len(parts[0]) == 44 {
			return parts[0]
		}
	}
	
	// Parâmetro chNFe
	if ch := parsed.Query().Get("chNFe"); len(ch) == 44 {
		return ch
	}
	
	return ""
}

// extractUFFromURL extrai o código da UF da URL.
func extractUFFromURL(rawURL string) string {
	// Mapear domínios para UF
	ufMap := map[string]string{
		"sp.gov.br": "SP", "rj.gov.br": "RJ", "mg.gov.br": "MG",
		"rs.gov.br": "RS", "pr.gov.br": "PR", "sc.gov.br": "SC",
		"ba.gov.br": "BA", "pe.gov.br": "PE", "ce.gov.br": "CE",
		"go.gov.br": "GO", "df.gov.br": "DF", "es.gov.br": "ES",
		"mt.gov.br": "MT", "ms.gov.br": "MS", "pa.gov.br": "PA",
		"am.gov.br": "AM", "ma.gov.br": "MA", "pb.gov.br": "PB",
		"rn.gov.br": "RN", "pi.gov.br": "PI", "al.gov.br": "AL",
		"se.gov.br": "SE", "ro.gov.br": "RO", "to.gov.br": "TO",
		"ac.gov.br": "AC", "ap.gov.br": "AP", "rr.gov.br": "RR",
	}
	
	for domain, uf := range ufMap {
		if strings.Contains(rawURL, domain) {
			return uf
		}
	}
	
	// Tentar extrair da chave de acesso (primeiros 2 dígitos)
	key := extractAccessKeyFromURL(rawURL)
	if len(key) >= 2 {
		return key[0:2]
	}
	
	return ""
}
