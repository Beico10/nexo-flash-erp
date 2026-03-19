// Package expenses — consulta SEFAZ via scraping.
//
// A SEFAZ não oferece API pública para consulta de NFC-e/NF-e.
// A solução é fazer scraping da página de consulta pública.
//
// URLs de consulta por estado:
//   - SP: https://www.nfce.fazenda.sp.gov.br/consulta
//   - Nacional: https://www.nfe.fazenda.gov.br/portal/consultaRecaptcha.aspx
package expenses

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SEFAZScraper implementa consulta à SEFAZ via scraping.
type SEFAZScraper struct {
	client *http.Client
}

func NewSEFAZScraper() *SEFAZScraper {
	return &SEFAZScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ConsultarNFCe consulta NFC-e pela URL do QR Code.
func (s *SEFAZScraper) ConsultarNFCe(ctx context.Context, qrURL string) (*Expense, error) {
	// Fazer request para a URL do QR Code
	req, err := http.NewRequestWithContext(ctx, "GET", qrURL, nil)
	if err != nil {
		return nil, fmt.Errorf("sefaz: erro ao criar request: %w", err)
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, ErrSEFAZUnavailable
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("sefaz: status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	html := string(body)
	
	// Parse do HTML para extrair dados
	expense := &Expense{
		IssueDate: time.Now(), // Será sobrescrito se encontrar
	}
	
	// Extrair chave de acesso
	if key := extractFromHTML(html, `Chave de Acesso.*?(\d{44})`); key != "" {
		expense.NFeKey = key
	}
	
	// Extrair número da nota
	if num := extractFromHTML(html, `(?:N[úu]mero|NFC-e).*?(\d+)`); num != "" {
		expense.NFeNumber = num
	}
	
	// Extrair dados do emitente
	if cnpj := extractFromHTML(html, `CNPJ.*?(\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2}|\d{14})`); cnpj != "" {
		expense.SupplierCNPJ = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(cnpj, ".", ""), "/", ""), "-", "")
	}
	if nome := extractFromHTML(html, `(?:Raz[ãa]o Social|Nome|Emitente).*?([A-Z][A-Za-z\s&\-\.]+(?:LTDA|ME|EPP|S\.?A\.?)?)`); nome != "" {
		expense.SupplierName = strings.TrimSpace(nome)
	}
	
	// Extrair valor total
	if total := extractFromHTML(html, `(?:Valor Total|TOTAL).*?R?\$?\s*([\d\.,]+)`); total != "" {
		expense.TotalAmount = parseMoneyBR(total)
		expense.TotalProducts = expense.TotalAmount
	}
	
	// Extrair data de emissão
	if data := extractFromHTML(html, `(?:Emiss[ãa]o|Data).*?(\d{2}/\d{2}/\d{4})`); data != "" {
		if t, err := time.Parse("02/01/2006", data); err == nil {
			expense.IssueDate = t
		}
	}
	
	// Extrair itens (se disponível na página)
	expense.Items = s.extractItems(html)
	
	// Se não conseguiu extrair dados básicos, retorna erro
	if expense.SupplierName == "" && expense.TotalAmount == 0 {
		return nil, fmt.Errorf("sefaz: não foi possível extrair dados da nota")
	}
	
	// Fallback para nome do fornecedor
	if expense.SupplierName == "" {
		expense.SupplierName = "Fornecedor não identificado"
	}
	
	return expense, nil
}

// ConsultarNFe consulta NF-e pela chave de acesso.
func (s *SEFAZScraper) ConsultarNFe(ctx context.Context, chave string) (*Expense, error) {
	// URL de consulta nacional
	consultaURL := fmt.Sprintf("https://www.nfe.fazenda.gov.br/portal/consultaRecaptcha.aspx?tipoConsulta=completa&tipoConteudo=XbSeqxE8pl8%%3d&nfe=%s", chave)
	
	req, err := http.NewRequestWithContext(ctx, "GET", consultaURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, ErrSEFAZUnavailable
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	
	expense := &Expense{
		NFeKey:    chave,
		IssueDate: time.Now(),
	}
	
	// Parse similar ao NFC-e
	if cnpj := extractFromHTML(html, `CNPJ.*?(\d{14})`); cnpj != "" {
		expense.SupplierCNPJ = cnpj
	}
	if nome := extractFromHTML(html, `Emitente.*?([A-Z][A-Za-z\s&\-\.]+)`); nome != "" {
		expense.SupplierName = strings.TrimSpace(nome)
	}
	if total := extractFromHTML(html, `Valor Total.*?R?\$?\s*([\d\.,]+)`); total != "" {
		expense.TotalAmount = parseMoneyBR(total)
		expense.TotalProducts = expense.TotalAmount
	}
	
	if expense.SupplierName == "" {
		expense.SupplierName = "Fornecedor não identificado"
	}
	
	return expense, nil
}

// ConsultarSAT consulta SAT-CF-e (São Paulo).
func (s *SEFAZScraper) ConsultarSAT(ctx context.Context, chave string) (*Expense, error) {
	// SAT usa formato similar ao NFC-e
	consultaURL := fmt.Sprintf("https://satsp.fazenda.sp.gov.br/COMSAT/Public/ConsultaPublica/ConsultaPublicaCfe.aspx?chaveConsulta=%s", chave)
	return s.ConsultarNFCe(ctx, consultaURL)
}

// extractItems tenta extrair itens da página HTML.
func (s *SEFAZScraper) extractItems(html string) []ExpenseItem {
	var items []ExpenseItem
	
	// Padrão para tabela de itens (varia por estado)
	// Tentar extrair linhas de tabela com: código, descrição, qtd, valor
	
	// Regex simplificado para capturar itens
	itemRegex := regexp.MustCompile(`(?s)<tr[^>]*>.*?(?:Código|Cód)[^<]*</td>.*?(\d+).*?</td>.*?<td[^>]*>([^<]+)</td>.*?<td[^>]*>(\d+(?:[,\.]\d+)?)</td>.*?<td[^>]*>R?\$?\s*([\d\.,]+)</td>.*?</tr>`)
	
	matches := itemRegex.FindAllStringSubmatch(html, -1)
	for i, match := range matches {
		if len(match) >= 5 {
			item := ExpenseItem{
				ItemOrder:   i + 1,
				ProductCode: match[1],
				Description: strings.TrimSpace(match[2]),
				Quantity:    parseNumberBR(match[3]),
				TotalPrice:  parseMoneyBR(match[4]),
			}
			if item.Quantity > 0 {
				item.UnitPrice = item.TotalPrice / item.Quantity
			}
			item.Unit = "UN"
			items = append(items, item)
		}
	}
	
	return items
}

// ════════════════════════════════════════════════════════════
// HELPERS
// ════════════════════════════════════════════════════════════

func extractFromHTML(html, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func parseMoneyBR(s string) float64 {
	// Formato brasileiro: 1.234,56
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")  // Remove separador de milhar
	s = strings.ReplaceAll(s, ",", ".") // Troca vírgula por ponto
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseNumberBR(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// ════════════════════════════════════════════════════════════
// FALLBACK: Parser de XML (quando disponível)
// ════════════════════════════════════════════════════════════

// ParseNFeXML faz parse de um XML de NF-e/NFC-e.
// Usado quando o usuário faz upload do arquivo XML.
func ParseNFeXML(xmlData []byte) (*Expense, error) {
	// Estrutura simplificada do XML da NF-e 4.0
	type nfeXML struct {
		XMLName xml.Name `xml:"nfeProc"`
		NFe     struct {
			InfNFe struct {
				ID   string `xml:"Id,attr"`
				Ide  struct {
					NNF   string `xml:"nNF"`
					Serie string `xml:"serie"`
					DhEmi string `xml:"dhEmi"`
				} `xml:"ide"`
				Emit struct {
					CNPJ  string `xml:"CNPJ"`
					XNome string `xml:"xNome"`
					IE    string `xml:"IE"`
				} `xml:"emit"`
				Det []struct {
					NItem string `xml:"nItem,attr"`
					Prod  struct {
						CProd  string `xml:"cProd"`
						CEAN   string `xml:"cEAN"`
						XProd  string `xml:"xProd"`
						NCM    string `xml:"NCM"`
						CFOP   string `xml:"CFOP"`
						UCom   string `xml:"uCom"`
						QCom   string `xml:"qCom"`
						VUnCom string `xml:"vUnCom"`
						VProd  string `xml:"vProd"`
					} `xml:"prod"`
				} `xml:"det"`
				Total struct {
					ICMSTot struct {
						VProd   string `xml:"vProd"`
						VDesc   string `xml:"vDesc"`
						VFrete  string `xml:"vFrete"`
						VNF     string `xml:"vNF"`
						VICMS   string `xml:"vICMS"`
						VIPI    string `xml:"vIPI"`
						VPIS    string `xml:"vPIS"`
						VCOFINS string `xml:"vCOFINS"`
					} `xml:"ICMSTot"`
				} `xml:"total"`
			} `xml:"infNFe"`
		} `xml:"NFe"`
	}
	
	var raw nfeXML
	if err := xml.Unmarshal(xmlData, &raw); err != nil {
		return nil, fmt.Errorf("XML inválido: %w", err)
	}
	
	inf := raw.NFe.InfNFe
	expense := &Expense{
		NFeKey:        strings.TrimPrefix(inf.ID, "NFe"),
		NFeNumber:     inf.Ide.NNF,
		NFeSeries:     inf.Ide.Serie,
		NFeType:       "nfe",
		SupplierCNPJ:  inf.Emit.CNPJ,
		SupplierName:  inf.Emit.XNome,
		SupplierIE:    inf.Emit.IE,
		TotalProducts: parseNumberBR(inf.Total.ICMSTot.VProd),
		TotalDiscount: parseNumberBR(inf.Total.ICMSTot.VDesc),
		TotalShipping: parseNumberBR(inf.Total.ICMSTot.VFrete),
		TotalAmount:   parseNumberBR(inf.Total.ICMSTot.VNF),
		ICMSAmount:    parseNumberBR(inf.Total.ICMSTot.VICMS),
		IPIAmount:     parseNumberBR(inf.Total.ICMSTot.VIPI),
		PISAmount:     parseNumberBR(inf.Total.ICMSTot.VPIS),
		COFINSAmount:  parseNumberBR(inf.Total.ICMSTot.VCOFINS),
	}
	
	// Data de emissão
	if t, err := time.Parse(time.RFC3339, inf.Ide.DhEmi); err == nil {
		expense.IssueDate = t
	}
	
	// Itens
	for _, det := range inf.Det {
		order, _ := strconv.Atoi(det.NItem)
		item := ExpenseItem{
			ItemOrder:   order,
			ProductCode: det.Prod.CProd,
			EAN:         det.Prod.CEAN,
			Description: det.Prod.XProd,
			NCM:         det.Prod.NCM,
			CFOP:        det.Prod.CFOP,
			Unit:        det.Prod.UCom,
			Quantity:    parseNumberBR(det.Prod.QCom),
			UnitPrice:   parseNumberBR(det.Prod.VUnCom),
			TotalPrice:  parseNumberBR(det.Prod.VProd),
		}
		expense.Items = append(expense.Items, item)
	}
	
	return expense, nil
}
