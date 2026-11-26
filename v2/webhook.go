package interage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Tipos de evento do webhook.
const (
	EventMessageCreated         = "message.created"
	EventConversationClosed     = "conversation.closed"
	EventCallEnded              = "call.ended"
	EventCallRecordingAvailable = "call.recording_available"
	EventExtensionStatusChanged = "extension.status_changed"
	EventPing                   = "ping" // botão Testar do webhook
)

// Headers de toda entrega do webhook.
const (
	HeaderWebhookSignature = "X-Interage-Signature"
	HeaderWebhookDelivery  = "X-Interage-Delivery" // id único do POST (não do evento)
)

// WebhookBatch é o corpo de cada POST do webhook: até 50 eventos. Nos eventos de chat e de
// ramal, processe events na ordem do array.
type WebhookBatch struct {
	Events []WebhookEvent `json:"events"`
}

// WebhookEvent é um evento. ID é único e se repete só em reentrega da mesma entrega: use-o
// para não processar duas vezes. Cada tipo preenche os seus campos.
type WebhookEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	// Conversation: message.created e conversation.closed (o detalhe de GET /conversations/{protocol}).
	Conversation *ConversationDetail `json:"conversation,omitempty"`
	// Message: message.created.
	Message *WebhookMessage `json:"message,omitempty"`
	// Call: call.ended e call.recording_available (o mesmo de GET /calls).
	Call *Call `json:"call,omitempty"`
	// Recording: call.recording_available (link válido por 7 dias).
	Recording *RecordingLink `json:"recording,omitempty"`
	// Extension: extension.status_changed.
	Extension *WebhookExtension `json:"extension,omitempty"`
}

// WebhookMessage é a Message da v2 sem status e sem reações (cada reação é um
// message.created de type reaction), com o link da mídia já gerado.
type WebhookMessage struct {
	ID        int64  `json:"id"`
	Direction string `json:"direction"`
	// SentBy: client, agent, bot, ai_agent ou agent_coex.
	SentBy    string                 `json:"sent_by"`
	Agent     *ConversationReference `json:"agent"`
	Type      string                 `json:"type"`
	Content   *string                `json:"content"`
	Caption   *string                `json:"caption"`
	ReplyToID *int64                 `json:"reply_to_id"`
	Media     *WebhookMessageMedia   `json:"media"`
	Edited    bool                   `json:"edited"`
	RemovedAt *time.Time             `json:"removed_at"`
	CreatedAt time.Time              `json:"created_at"`
}

// WebhookMessageMedia é a mídia da mensagem mais o link para baixar (7 dias). Com
// Downloadable e Link nil, o link não pôde ser gerado: peça por Conversations.FileLink.
type WebhookMessageMedia struct {
	Mime          *string    `json:"mime"`
	FileName      *string    `json:"file_name"`
	Downloadable  bool       `json:"downloadable"`
	URL           *string    `json:"url"` // link original, quando a mídia foi enviada por link externo
	Link          *string    `json:"link"`
	LinkExpiresAt *time.Time `json:"link_expires_at"`
}

// WebhookExtension é o ramal da v2 (GET /extensions/{number}) mais o status anterior.
type WebhookExtension struct {
	Extension
	PreviousStatus string `json:"previous_status"`
}

// Erros da verificação do webhook.
var (
	ErrWebhookSignature = errors.New("interage: assinatura do webhook inválida")
	ErrWebhookExpired   = errors.New("interage: assinatura do webhook fora da tolerância de tempo")
)

// DefaultWebhookTolerance é a diferença máxima aceita entre o horário da assinatura e agora.
const DefaultWebhookTolerance = 5 * time.Minute

// VerifyWebhookSignature confere o header X-Interage-Signature ("t=<unix>,v1=<hex>[,v1=...]",
// HMAC-SHA256 de "<t>.<corpo>" com o segredo do webhook) contra o corpo cru, como chegou.
// Durante a troca do segredo há dois v1: vale qualquer um. tolerance zero usa
// DefaultWebhookTolerance; negativo não confere o horário.
func VerifyWebhookSignature(secret, header string, body []byte, tolerance time.Duration) error {
	var ts string
	var assinaturas [][]byte
	for _, parte := range strings.Split(header, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(parte), "=")
		if !ok {
			continue
		}
		switch k {
		case "t":
			ts = v
		case "v1":
			if b, err := hex.DecodeString(v); err == nil {
				assinaturas = append(assinaturas, b)
			}
		}
	}
	unix, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || len(assinaturas) == 0 {
		return ErrWebhookSignature
	}
	if tolerance == 0 {
		tolerance = DefaultWebhookTolerance
	}
	if tolerance > 0 {
		if d := time.Since(time.Unix(unix, 0)); d > tolerance || d < -tolerance {
			return ErrWebhookExpired
		}
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(body)
	esperada := mac.Sum(nil)
	for _, a := range assinaturas {
		if hmac.Equal(a, esperada) {
			return nil
		}
	}
	return ErrWebhookSignature
}

// ParseWebhook lê o corpo do POST do webhook, confere a assinatura e decodifica o lote.
//
//	http.HandleFunc("/interage", func(w http.ResponseWriter, r *http.Request) {
//		lote, err := interage.ParseWebhook(r, segredo, 0)
//		if err != nil { http.Error(w, "assinatura inválida", http.StatusUnauthorized); return }
//		for _, ev := range lote.Events { ... }
//		w.WriteHeader(http.StatusNoContent)
//	})
func ParseWebhook(r *http.Request, secret string, tolerance time.Duration) (*WebhookBatch, error) {
	corpo, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("interage: erro ao ler o corpo do webhook: %w", err)
	}
	if err := VerifyWebhookSignature(secret, r.Header.Get(HeaderWebhookSignature), corpo, tolerance); err != nil {
		return nil, err
	}
	var lote WebhookBatch
	if err := json.Unmarshal(corpo, &lote); err != nil {
		return nil, fmt.Errorf("interage: corpo do webhook fora do formato: %w", err)
	}
	return &lote, nil
}
