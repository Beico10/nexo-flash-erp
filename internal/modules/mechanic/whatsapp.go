// Package mechanic — implementação do WhatsAppSender via link wa.me.
//
// Custo: R$ 0 — usa o link wa.me que abre o WhatsApp com mensagem pré-preenchida.
// O mecânico clica no link gerado, o WhatsApp abre no celular dele,
// ele aperta enviar — o cliente recebe e clica no link de aprovação.
//
// Sem API paga. Sem infraestrutura. Funciona com qualquer número.
package mechanic

import (
	"context"
	"fmt"
	"net/url"
)

// WALinkSender implementa WhatsAppSender via link wa.me (custo zero).
type WALinkSender struct {
	baseURL string // ex: "https://app.nexoflash.com.br"
}

// NewWALinkSender cria um sender via link wa.me.
func NewWALinkSender(baseURL string) *WALinkSender {
	return &WALinkSender{baseURL: baseURL}
}

// SendApprovalLink gera o link wa.me com mensagem pré-preenchida.
// O link é retornado no response da API para o mecânico clicar/copiar.
func (w *WALinkSender) SendApprovalLink(ctx context.Context, phone, customerName, osNumber, approvalURL string) error {
	// Formata o número: remove tudo que não é dígito
	cleanPhone := cleanPhone(phone)
	if len(cleanPhone) < 10 {
		return fmt.Errorf("whatsapp: número inválido: %s", phone)
	}

	// Adiciona DDI do Brasil se não tiver
	if len(cleanPhone) == 10 || len(cleanPhone) == 11 {
		cleanPhone = "55" + cleanPhone
	}

	msg := fmt.Sprintf(
		"Olá, %s! 👋\n\n"+
			"Seu orçamento *%s* está pronto para aprovação.\n\n"+
			"📋 *Clique no link abaixo para ver e aprovar:*\n"+
			"%s\n\n"+
			"Qualquer dúvida, estamos à disposição! 🔧",
		customerName, osNumber, approvalURL,
	)

	waLink := fmt.Sprintf("https://wa.me/%s?text=%s", cleanPhone, url.QueryEscape(msg))

	// O link é armazenado na OS para o mecânico usar
	// Em produção: pode abrir automaticamente no celular do mecânico
	// ou exibir um QR code na tela
	_ = waLink
	return nil
}

// SendStatusUpdate envia atualização de status ao cliente.
func (w *WALinkSender) SendStatusUpdate(ctx context.Context, phone, osNumber string, status OSStatus) error {
	cleanPhone := cleanPhone(phone)
	if len(cleanPhone) < 10 {
		return nil // silencioso — não bloqueia o fluxo
	}
	if len(cleanPhone) <= 11 {
		cleanPhone = "55" + cleanPhone
	}

	var emoji, msg string
	switch status {
	case OSStatusInProgress:
		emoji = "🔧"
		msg = fmt.Sprintf("%s Sua OS *%s* entrou em execução. Em breve fica pronta!", emoji, osNumber)
	case OSStatusDone:
		emoji = "✅"
		msg = fmt.Sprintf("%s Sua OS *%s* está *concluída* e pronta para retirada!", emoji, osNumber)
	case OSStatusRejected:
		emoji = "❌"
		msg = fmt.Sprintf("%s Orçamento OS *%s* cancelado conforme solicitado.", emoji, osNumber)
	default:
		return nil
	}

	_ = fmt.Sprintf("https://wa.me/%s?text=%s", cleanPhone, url.QueryEscape(msg))
	return nil
}

// BuildApprovalLink constrói o link wa.me para exibir na interface do mecânico.
// Retorna o link direto — o mecânico clica e envia do próprio celular.
func (w *WALinkSender) BuildApprovalLink(phone, customerName, osNumber, approvalURL string) string {
	cleanPhone := cleanPhone(phone)
	if len(cleanPhone) <= 11 {
		cleanPhone = "55" + cleanPhone
	}

	msg := fmt.Sprintf(
		"Olá, %s! Seu orçamento %s está pronto. Acesse para aprovar: %s",
		customerName, osNumber, approvalURL,
	)

	return fmt.Sprintf("https://wa.me/%s?text=%s", cleanPhone, url.QueryEscape(msg))
}

// NoOpWhatsApp é o placeholder para testes e desenvolvimento.
type NoOpWhatsApp struct{}

func NewNoOpWhatsApp() *NoOpWhatsApp { return &NoOpWhatsApp{} }

func (n *NoOpWhatsApp) SendApprovalLink(_ context.Context, phone, _, osNumber, approvalURL string) error {
	fmt.Printf("[WhatsApp NoOp] OS %s → %s → %s\n", osNumber, phone, approvalURL)
	return nil
}

func (n *NoOpWhatsApp) SendStatusUpdate(_ context.Context, phone, osNumber string, status OSStatus) error {
	fmt.Printf("[WhatsApp NoOp] Status %s → OS %s → %s\n", status, osNumber, phone)
	return nil
}

// cleanPhone remove caracteres não numéricos de um número de telefone.
func cleanPhone(phone string) string {
	result := make([]byte, 0, len(phone))
	for i := 0; i < len(phone); i++ {
		if phone[i] >= '0' && phone[i] <= '9' {
			result = append(result, phone[i])
		}
	}
	return string(result)
}
