// Package eventbus implementa o barramento de eventos interno do Nexo One.
// Utiliza NATS JetStream para garantia de entrega (at-least-once) e
// desacoplamento total entre módulos.
//
// Exemplo de uso:
//
//	eventbus.Publish(ctx, EventBillingCreated, BillingPayload{...})
//	eventbus.Subscribe(EventBillingCreated, handler)
package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// EventType define os eventos do sistema — contrato imutável entre módulos.
// Novos eventos: adicione aqui e nunca remova (backward compat).
type EventType string

const (
	// Financeiro
	EventBillingCreated  EventType = "nexo.billing.created"
	EventPaymentReceived EventType = "nexo.payment.received"
	EventSplitProcessed  EventType = "nexo.payment.split_processed"

	// Estoque
	EventStockUpdated  EventType = "nexo.stock.updated"
	EventStockLow      EventType = "nexo.stock.low_alert"

	// Fiscal
	EventInvoiceIssued     EventType = "nexo.fiscal.invoice_issued"
	EventInvoiceContingency EventType = "nexo.fiscal.contingency_mode"

	// Logística
	EventRouteClosed   EventType = "nexo.logistics.route_closed"
	EventCTeIssued     EventType = "nexo.logistics.cte_issued"

	// IA — todas as ações de IA geram eventos PENDING, nunca escrevem diretamente
	EventAISuggestion EventType = "nexo.ai.suggestion_pending"
	EventAIApproved   EventType = "nexo.ai.suggestion_approved"
	EventAIRejected   EventType = "nexo.ai.suggestion_rejected"
)

// Envelope é o wrapper padrão de todos os eventos.
// Inclui tenant_id para rastreabilidade e conformidade com LGPD.
type Envelope struct {
	EventID   string          `json:"event_id"`
	TenantID  string          `json:"tenant_id"`
	EventType EventType       `json:"event_type"`
	OccurredAt time.Time      `json:"occurred_at"`
	Payload   json.RawMessage `json:"payload"`
}

// Bus é a interface do event bus — facilita mock em testes.
type Bus interface {
	Publish(ctx context.Context, tenantID string, eventType EventType, payload any) error
	Subscribe(eventType EventType, handler func(Envelope)) error
	Close()
}

type natsBus struct {
	nc *nats.Conn
	js jetstream.JetStream
}

// New cria um novo Bus conectado ao NATS JetStream.
// natsURL ex: "nats://localhost:4222"
func New(natsURL string) (Bus, error) {
	nc, err := nats.Connect(natsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("eventbus: falha ao conectar NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("eventbus: falha ao iniciar JetStream: %w", err)
	}

	// Stream "NEXO" cobre todos os eventos do sistema
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "NEXO",
		Subjects: []string{"nexo.>"},
		MaxAge:   72 * time.Hour,
		Storage:  jetstream.FileStorage,
		Replicas: 1, // aumente para 3 em produção HA
	})
	if err != nil {
		return nil, fmt.Errorf("eventbus: falha ao criar stream NEXO: %w", err)
	}

	return &natsBus{nc: nc, js: js}, nil
}

func (b *natsBus) Publish(ctx context.Context, tenantID string, eventType EventType, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("eventbus: marshal payload: %w", err)
	}
	env := Envelope{
		EventID:    newULID(),
		TenantID:   tenantID,
		EventType:  eventType,
		OccurredAt: time.Now().UTC(),
		Payload:    raw,
	}
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	_, err = b.js.Publish(ctx, string(eventType), data)
	if err != nil {
		return fmt.Errorf("eventbus: publish %s: %w", eventType, err)
	}
	slog.Debug("evento publicado", "type", eventType, "tenant", tenantID)
	return nil
}

func (b *natsBus) Subscribe(eventType EventType, handler func(Envelope)) error {
	ctx := context.Background()
	cons, err := b.js.CreateOrUpdateConsumer(ctx, "NEXO", jetstream.ConsumerConfig{
		FilterSubject: string(eventType),
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverAllPolicy,
	})
	if err != nil {
		return fmt.Errorf("eventbus: criar consumer %s: %w", eventType, err)
	}

	_, err = cons.Consume(func(msg jetstream.Msg) {
		var env Envelope
		if err := json.Unmarshal(msg.Data(), &env); err != nil {
			slog.Error("eventbus: unmarshal falhou", "err", err)
			msg.Nak()
			return
		}
		handler(env)
		msg.Ack()
	})
	return err
}

func (b *natsBus) Close() {
	b.nc.Drain()
	b.nc.Close()
}

// newULID gera um ID único — substitua por github.com/oklog/ulid em produção
func newULID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
