// Package ai — IA Concierge do Nexo One.
//
// Função: lê XMLs de NF-e de compra e configura 90% do tenant e catálogo
// em até 5 minutos, sem que o usuário precise digitar produto por produto.
//
// DIRETRIZ CRÍTICA: O Concierge NUNCA persiste dados diretamente.
// Cada campo detectado no XML gera uma sugestão via Gateway.Suggest()
// que exige aprovação humana antes de ser salvo.
//
// Fluxo:
//  1. Usuário faz upload do XML da NF-e de compra
//  2. Concierge.ParseNFe() extrai produtos, NCMs, fornecedor, valores
//  3. Para cada produto: Gateway.Suggest(SuggestionOnboardField)
//  4. Interface exibe cards de aprovação em lote ("Aprovar tudo" / item por item)
//  5. Apenas itens aprovados são persistidos no catálogo
package ai

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NFeParsed representa os dados extraídos de um XML de NF-e.
type NFeParsed struct {
	ChaveNFe    string
	NumeroNFe   string
	DataEmissao time.Time
	Emitente    EmitenteNFe
	Destinatario DestinatarioNFe
	Items       []NFeItem
	TotalProdutos float64
	TotalNFe    float64
	CFOP        string
}

// EmitenteNFe dados do fornecedor extraídos da NF-e.
type EmitenteNFe struct {
	CNPJ        string
	RazaoSocial string
	NomeFantasia string
	IE          string
	Endereco    string
	Municipio   string
	UF          string
}

// DestinatarioNFe dados do destinatário (o tenant comprando).
type DestinatarioNFe struct {
	CNPJ string
	Nome string
}

// NFeItem representa um produto no XML da NF-e.
type NFeItem struct {
	Ordem       int
	CodigoProd  string  // código do fornecedor
	Descricao   string
	NCM         string  // 8 dígitos
	CFOP        string
	UnidadeCom  string  // UN, KG, CX, etc.
	QuantidadeCom float64
	ValorUnitCom  float64
	ValorTotalProd float64
	EAN         string  // código de barras EAN-13 (se presente)
	// Impostos da NF-e (para base do cashback tributário)
	ValorIPI    float64
	ValorICMS   float64
	// Alíquotas IBS/CBS serão calculadas pelo motor fiscal (NCM → rate)
}

// OnboardSuggestion é uma sugestão gerada a partir de um item da NF-e.
type OnboardSuggestion struct {
	NFeItem    NFeItem
	Action     string // "create_product" | "create_supplier" | "update_ncm"
	Confidence float64
	Reason     string
}

// Concierge é o agente de IA de onboarding do Nexo One.
type Concierge struct {
	gateway *Gateway
}

// NewConcierge cria um novo agente Concierge.
func NewConcierge(g *Gateway) *Concierge {
	return &Concierge{gateway: g}
}

// ProcessNFeXML recebe o conteúdo de um XML de NF-e e gera sugestões de onboarding.
// Não persiste nada — tudo via Gateway.Suggest() → status=pending.
func (c *Concierge) ProcessNFeXML(ctx context.Context, tenantID string, xmlData []byte) (*NFeParsed, int, error) {
	nfe, err := parseNFeXML(xmlData)
	if err != nil {
		return nil, 0, fmt.Errorf("concierge.ProcessNFeXML: XML inválido: %w", err)
	}

	suggestions := c.buildSuggestions(nfe)
	var created int

	for _, s := range suggestions {
		data := map[string]any{
			"codigo_produto":  s.NFeItem.CodigoProd,
			"descricao":       s.NFeItem.Descricao,
			"ncm":             s.NFeItem.NCM,
			"ean":             s.NFeItem.EAN,
			"unidade":         s.NFeItem.UnidadeCom,
			"preco_custo":     s.NFeItem.ValorUnitCom,
			"cfop":            s.NFeItem.CFOP,
			"acao":            s.Action,
			"fornecedor_cnpj": nfe.Emitente.CNPJ,
			"fornecedor_nome": nfe.Emitente.RazaoSocial,
			"nfe_chave":       nfe.ChaveNFe,
		}

		err := c.gateway.Suggest(ctx, &Suggestion{
			TenantID:      tenantID,
			Type:          SuggestionOnboardField,
			TargetTable:   "products",
			SuggestedData: data,
			Reason: fmt.Sprintf("Produto '%s' (NCM: %s) detectado na NF-e %s de %s.",
				s.NFeItem.Descricao, s.NFeItem.NCM, nfe.NumeroNFe, nfe.Emitente.RazaoSocial),
			Confidence:  s.Confidence,
			CreatedByAI: "concierge-v1",
		})
		if err != nil {
			continue // log e continua — não falha o lote todo
		}
		created++
	}

	return nfe, created, nil
}

// buildSuggestions analisa os itens da NF-e e gera sugestões.
func (c *Concierge) buildSuggestions(nfe *NFeParsed) []OnboardSuggestion {
	var suggestions []OnboardSuggestion
	for _, item := range nfe.Items {
		confidence := 0.9 // alta confiança se NCM e EAN presentes
		if item.NCM == "" {
			confidence -= 0.3
		}
		if item.EAN == "" {
			confidence -= 0.1
		}
		suggestions = append(suggestions, OnboardSuggestion{
			NFeItem:    item,
			Action:     "create_product",
			Confidence: confidence,
			Reason:     fmt.Sprintf("Item %d da NF-e", item.Ordem),
		})
	}
	return suggestions
}

// =============================================================================
// Parser XML da NF-e (estrutura simplificada — NF-e 4.0)
// Referência: Manual de Orientação do Contribuinte NF-e v7.0
// =============================================================================

// nfeXML é a estrutura para deserializar o XML da NF-e 4.0.
type nfeXML struct {
	XMLName xml.Name `xml:"nfeProc"`
	NFe     struct {
		InfNFe struct {
			ID   string `xml:"Id,attr"`
			Ide  struct {
				NNF  string `xml:"nNF"`
				DhEmi string `xml:"dhEmi"`
			} `xml:"ide"`
			Emit struct {
				CNPJ   string `xml:"CNPJ"`
				XNome  string `xml:"xNome"`
				XFant  string `xml:"xFant"`
				IE     string `xml:"IE"`
				EnderEmit struct {
					XLgr   string `xml:"xLgr"`
					XMun   string `xml:"xMun"`
					UF     string `xml:"UF"`
				} `xml:"enderEmit"`
			} `xml:"emit"`
			Dest struct {
				CNPJ  string `xml:"CNPJ"`
				XNome string `xml:"xNome"`
			} `xml:"dest"`
			Det []struct {
				NItem string `xml:"nItem,attr"`
				Prod struct {
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
				Imposto struct {
					IPI struct {
						VIpi string `xml:"IPITrib>vIPI"`
					} `xml:"IPI"`
					ICMS struct {
						ICMS00 struct {
							VICMS string `xml:"vICMS"`
						} `xml:"ICMS00"`
					} `xml:"ICMS"`
				} `xml:"imposto"`
			} `xml:"det"`
			Total struct {
				ICMSTot struct {
					VProd string `xml:"vProd"`
					VNF   string `xml:"vNF"`
				} `xml:"ICMSTot"`
			} `xml:"total"`
		} `xml:"infNFe"`
	} `xml:"NFe"`
}

func parseNFeXML(data []byte) (*NFeParsed, error) {
	// Tenta como nfeProc (com protocolo) e como nfeProc sem wrapper
	var raw nfeXML
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parseNFeXML: %w", err)
	}

	inf := raw.NFe.InfNFe
	nfe := &NFeParsed{
		ChaveNFe:  strings.TrimPrefix(inf.ID, "NFe"),
		NumeroNFe: inf.Ide.NNF,
		Emitente: EmitenteNFe{
			CNPJ:        inf.Emit.CNPJ,
			RazaoSocial: inf.Emit.XNome,
			NomeFantasia: inf.Emit.XFant,
			IE:          inf.Emit.IE,
			Municipio:   inf.Emit.EnderEmit.XMun,
			UF:          inf.Emit.EnderEmit.UF,
		},
		Destinatario: DestinatarioNFe{
			CNPJ: inf.Dest.CNPJ,
			Nome: inf.Dest.XNome,
		},
	}

	// Data de emissão
	if t, err := time.Parse(time.RFC3339, inf.Ide.DhEmi); err == nil {
		nfe.DataEmissao = t
	}

	// Totais
	nfe.TotalProdutos, _ = strconv.ParseFloat(inf.Total.ICMSTot.VProd, 64)
	nfe.TotalNFe, _ = strconv.ParseFloat(inf.Total.ICMSTot.VNF, 64)

	// Itens
	for _, det := range inf.Det {
		order, _ := strconv.Atoi(det.NItem)
		qtd, _ := strconv.ParseFloat(det.Prod.QCom, 64)
		vUnit, _ := strconv.ParseFloat(det.Prod.VUnCom, 64)
		vProd, _ := strconv.ParseFloat(det.Prod.VProd, 64)
		vIPI, _ := strconv.ParseFloat(det.Imposto.IPI.VIpi, 64)
		vICMS, _ := strconv.ParseFloat(det.Imposto.ICMS.ICMS00.VICMS, 64)

		item := NFeItem{
			Ordem:          order,
			CodigoProd:     det.Prod.CProd,
			Descricao:      det.Prod.XProd,
			NCM:            det.Prod.NCM,
			CFOP:           det.Prod.CFOP,
			UnidadeCom:     det.Prod.UCom,
			QuantidadeCom:  qtd,
			ValorUnitCom:   vUnit,
			ValorTotalProd: vProd,
			EAN:            det.Prod.CEAN,
			ValorIPI:       vIPI,
			ValorICMS:      vICMS,
		}
		nfe.Items = append(nfe.Items, item)
	}

	return nfe, nil
}
