package interage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Webhooks — verificação da assinatura e modelos dos eventos
// ─────────────────────────────────────────────────────────────────────────────

// Headers de toda entrega de webhook.
const (
	// HeaderWebhookSignature traz "t=<unix>,v1=<hex HMAC-SHA256>"; durante a troca do
	// segredo (24h), vem um segundo v1, assinado com o segredo anterior.
	HeaderWebhookSignature = "X-Interage-Signature"
	// HeaderWebhookDelivery identifica cada POST (não o evento): muda a cada envio e
	// tem o mesmo valor do header Idempotency-Key. Para deduplicar, use WebhookEvent.ID.
	HeaderWebhookDelivery = "X-Interage-Delivery"
)

// DefaultWebhookTolerance é a diferença máxima aceita entre o t da assinatura e o
// relógio local — barra a reprodução de uma entrega capturada.
const DefaultWebhookTolerance = 300 * time.Second

// Tipos de evento de webhook (WebhookEvent.Type).
const (
	// WebhookMessageCreated — mensagem nova na conversa (Conversation + Message). Chega em ordem.
	WebhookMessageCreated = "message.created"
	// WebhookConversationClosed — conversa encerrada (Conversation). Chega em ordem.
	WebhookConversationClosed = "conversation.closed"
	// WebhookCallEnded — ligação encerrada (Call). Pode chegar fora de ordem: use CreatedAt.
	WebhookCallEnded = "call.ended"
	// WebhookCallRecordingAvailable — gravação disponível (Call + Recording). Pode chegar fora de ordem.
	WebhookCallRecordingAvailable = "call.recording_available"
	// WebhookExtensionStatusChanged — ramal mudou de status (Extension). Chega em ordem.
	WebhookExtensionStatusChanged = "extension.status_changed"
	// WebhookPing — evento do botão Testar do painel: só ID, Type e CreatedAt.
	WebhookPing = "ping"
)

// ErrInvalidWebhookSignature indica assinatura ausente, malformada, fora da
// tolerância de tempo ou que não confere com nenhum segredo.
var ErrInvalidWebhookSignature = errors.New("interage: assinatura do webhook inválida")

// WebhookBatch é o corpo de cada POST: um ou mais eventos (até 50). Processe
// Events na ordem do array.
type WebhookBatch struct {
	Events []WebhookEvent `json:"events"`
}

// WebhookEvent é um evento entregue ao seu endpoint. Só os campos do tipo vêm
// preenchidos (ver as constantes Webhook*).
type WebhookEvent struct {
	// ID é único por evento e o mesmo num reenvio raro: use para deduplicar.
	ID   string `json:"id"`
	Type string `json:"type"`
	// CreatedAt em RFC3339.
	CreatedAt string `json:"created_at"`
	// Conversation vem em message.created e conversation.closed — o mesmo formato
	// do cabeçalho de Omni.GetConversationHistory.
	Conversation *ConversationSummaryResponse `json:"conversation,omitempty"`
	Message      *WebhookMessage              `json:"message,omitempty"`
	Call         *WebhookCall                 `json:"call,omitempty"`
	Recording    *WebhookRecording            `json:"recording,omitempty"`
	Extension    *WebhookExtension            `json:"extension,omitempty"`
}

// WebhookMessage é a mensagem do message.created: a do histórico, sem status e sem
// media_url, mais o link pronto para baixar a mídia.
type WebhookMessage struct {
	ID        int64  `json:"id"`
	Direction string `json:"direction"`
	// SentBy: client | bot | ai_agent | agent | agent_coex.
	SentBy        string `json:"sent_by"`
	MessageType   string `json:"message_type"`
	Content       string `json:"content,omitempty"`
	Caption       string `json:"caption,omitempty"`
	MediaMime     string `json:"media_mime,omitempty"`
	MediaFileName string `json:"media_file_name,omitempty"`
	// MediaLink é o link temporário para baixar a mídia (vale até MediaLinkExpiresAt).
	MediaLink          string `json:"media_link,omitempty"`
	MediaLinkExpiresAt string `json:"media_link_expires_at,omitempty"`
	// MediaUnavailable: a mensagem tem mídia, mas sem link disponível.
	MediaUnavailable bool `json:"media_unavailable,omitempty"`
	// ReplyToID é a mensagem citada; numa reação, a mensagem que recebeu a reação.
	ReplyToID *int64 `json:"reply_to_id,omitempty"`
	CreatedAt string `json:"created_at"`
}

// WebhookCall é a ligação do call.ended e do call.recording_available: o item de
// Telephony.ListCallHistory sem unique_id e linked_id.
type WebhookCall struct {
	CallID           string `json:"call_id"`
	CallerIDNum      string `json:"caller_id_num,omitempty"`
	CallerIDName     string `json:"caller_id_name,omitempty"`
	CalledNumber     string `json:"called_number,omitempty"`
	DestinationNum   string `json:"destination_num,omitempty"`
	DIDNumber        string `json:"did_number,omitempty"`
	CallType         string `json:"call_type,omitempty"`
	RouteType        string `json:"route_type,omitempty"`
	TrunkName        string `json:"trunk_name,omitempty"`
	Status           string `json:"status"`
	CallResult       string `json:"call_result,omitempty"`
	HangupCause      string `json:"hangup_cause,omitempty"`
	IsAbandoned      bool   `json:"is_abandoned"`
	AbandonReason    string `json:"abandon_reason,omitempty"`
	StartTime        string `json:"start_time"`
	AnswerTime       string `json:"answer_time,omitempty"`
	EndTime          string `json:"end_time,omitempty"`
	Duration         *int   `json:"duration,omitempty"`
	BillableSec      *int   `json:"billable_sec,omitempty"`
	QueueWaitTime    *int   `json:"queue_wait_time,omitempty"`
	QueueName        string `json:"queue_name,omitempty"`
	AgentExtension   string `json:"agent_extension,omitempty"`
	AgentName        string `json:"agent_name,omitempty"`
	AgentAnsweredAt  string `json:"agent_answered_at,omitempty"`
	RecordingEnabled *bool  `json:"recording_enabled,omitempty"`
	ProtocolNumber   string `json:"protocol_number,omitempty"`
}

// WebhookRecording é o arquivo da gravação no call.recording_available — o mesmo
// link de Telephony.CreateRecordingTempLink, já gerado.
type WebhookRecording struct {
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// WebhookExtension é a foto do ramal no extension.status_changed.
// Status (e PreviousStatus) segue a prioridade offline, in_call, dnd, online.
type WebhookExtension struct {
	Number          string `json:"number"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	PreviousStatus  string `json:"previous_status"`
	Registered      bool   `json:"registered"`
	InCall          bool   `json:"in_call"`
	Dnd             bool   `json:"dnd"`
	ForwardAlways   string `json:"forward_always,omitempty"`
	ForwardNoAnswer string `json:"forward_no_answer,omitempty"`
	ForwardBusy     string `json:"forward_busy,omitempty"`
}

// VerifyWebhookSignature confere o header X-Interage-Signature contra o corpo
// BRUTO da requisição (não reserialize o JSON), com tolerância de 300s. Informe o
// segredo atual e, durante a troca, também o anterior: basta um v1 conferir com
// algum deles. A comparação é em tempo constante.
func VerifyWebhookSignature(header string, body []byte, secrets ...string) error {
	return VerifyWebhookSignatureAt(header, body, time.Now(), DefaultWebhookTolerance, secrets...)
}

// VerifyWebhookSignatureAt é VerifyWebhookSignature com relógio e tolerância
// explícitos (tolerance <= 0 usa DefaultWebhookTolerance).
func VerifyWebhookSignatureAt(header string, body []byte, now time.Time, tolerance time.Duration, secrets ...string) error {
	if tolerance <= 0 {
		tolerance = DefaultWebhookTolerance
	}
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
	if err != nil {
		return fmt.Errorf("%w: timestamp ausente ou inválido", ErrInvalidWebhookSignature)
	}
	if len(assinaturas) == 0 {
		return fmt.Errorf("%w: nenhum v1 no header", ErrInvalidWebhookSignature)
	}
	diff := now.Sub(time.Unix(unix, 0))
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		return fmt.Errorf("%w: timestamp fora da tolerância de %s", ErrInvalidWebhookSignature, tolerance)
	}
	conferiu := false
	for _, segredo := range secrets {
		if segredo == "" {
			continue
		}
		mac := hmac.New(sha256.New, []byte(segredo))
		mac.Write([]byte(ts + "."))
		mac.Write(body)
		esperado := mac.Sum(nil)
		for _, recebido := range assinaturas {
			// Sem sair no primeiro acerto: o tempo não depende de qual par conferiu.
			if hmac.Equal(recebido, esperado) {
				conferiu = true
			}
		}
	}
	if !conferiu {
		return fmt.Errorf("%w: nenhum v1 confere com os segredos informados", ErrInvalidWebhookSignature)
	}
	return nil
}

// DecodeWebhook decodifica o corpo de uma entrega, sem verificar a assinatura.
func DecodeWebhook(body []byte) (*WebhookBatch, error) {
	var b WebhookBatch
	if err := json.Unmarshal(body, &b); err != nil {
		return nil, fmt.Errorf("interage: corpo do webhook inválido: %w", err)
	}
	return &b, nil
}

// ParseWebhook verifica a assinatura (VerifyWebhookSignature) e decodifica o corpo.
//
//	body, _ := io.ReadAll(r.Body)
//	lote, err := interage.ParseWebhook(r.Header.Get(interage.HeaderWebhookSignature), body, segredo)
//	if err != nil { w.WriteHeader(http.StatusUnauthorized); return }
//	for _, ev := range lote.Events { ... } // em ordem; deduplique por ev.ID
func ParseWebhook(signatureHeader string, body []byte, secrets ...string) (*WebhookBatch, error) {
	if err := VerifyWebhookSignature(signatureHeader, body, secrets...); err != nil {
		return nil, err
	}
	return DecodeWebhook(body)
}
