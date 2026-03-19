// Package baas — placeholder NoOp do gateway BaaS.
// Substitua por Celcoin, BMP, Inter ou outro provedor em produção.
package baas

import (
	"context"
	"fmt"
	"time"
)

// NoOpGateway implementa BaaSGateway sem fazer chamadas reais.
// Usado em desenvolvimento e testes.
type NoOpGateway struct{}

func NewNoOpGateway() *NoOpGateway { return &NoOpGateway{} }

func (n *NoOpGateway) CreatePixCharge(_ context.Context, charge *PixCharge) (*PixCharge, error) {
	charge.ID = fmt.Sprintf("mock_%d", time.Now().UnixNano())
	charge.QRCodeText = fmt.Sprintf("00020126580014br.gov.bcb.pix0136%s5204000053039865802BR5925Nexo Flash ERP6009Sao Paulo62070503***6304ABCD", charge.TxID)
	charge.QRCodeImage = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	charge.Status = PaymentPending
	fmt.Printf("[BaaS NoOp] PIX criado: txID=%s valor=R$%.2f\n", charge.TxID, charge.Amount)
	return charge, nil
}

func (n *NoOpGateway) CancelPixCharge(_ context.Context, _, txID string) error {
	fmt.Printf("[BaaS NoOp] PIX cancelado: %s\n", txID)
	return nil
}

func (n *NoOpGateway) GetPixCharge(_ context.Context, _, txID string) (*PixCharge, error) {
	return &PixCharge{TxID: txID, Status: PaymentPending}, nil
}

func (n *NoOpGateway) CreateBoleto(_ context.Context, boleto *BoletoCharge) (*BoletoCharge, error) {
	boleto.ID = fmt.Sprintf("mock_%d", time.Now().UnixNano())
	boleto.OurNumber = fmt.Sprintf("00190%010d", time.Now().UnixNano()%10000000000)
	boleto.BarCode = "34191.09009 01013.869335 61929.300000 1 00000000010000"
	boleto.DigitableLine = boleto.BarCode
	boleto.PDFURL = fmt.Sprintf("https://boleto.nexoflash.com.br/mock/%s.pdf", boleto.OurNumber)
	boleto.Status = PaymentPending
	fmt.Printf("[BaaS NoOp] Boleto criado: %s valor=R$%.2f\n", boleto.OurNumber, boleto.Amount)
	return boleto, nil
}

func (n *NoOpGateway) CancelBoleto(_ context.Context, _, ourNumber string) error {
	fmt.Printf("[BaaS NoOp] Boleto cancelado: %s\n", ourNumber)
	return nil
}

func (n *NoOpGateway) ProcessWebhook(_ context.Context, payload []byte, _ string) (*WebhookEvent, error) {
	// Em produção: validar assinatura HMAC antes de processar
	fmt.Printf("[BaaS NoOp] Webhook recebido: %d bytes\n", len(payload))
	return &WebhookEvent{
		EventType: "payment.paid",
		Amount:    0,
		PaidAt:    time.Now(),
	}, nil
}
