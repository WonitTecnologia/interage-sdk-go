package interage

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ─────────────────────────────────────────────────────────────────────────────
// Modelos
// ─────────────────────────────────────────────────────────────────────────────

// ExtensionResponse é um ramal da central telefônica.
type ExtensionResponse struct {
	Number     string `json:"number"`
	Name       string `json:"name"`
	CallerID   string `json:"caller_id,omitempty"`
	IsOnline   bool   `json:"is_online"`
	DNDEnabled bool   `json:"dnd_enabled"`
}

// ListExtensionsResponse é o envelope paginado da listagem de ramais.
type ListExtensionsResponse struct {
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Items    []ExtensionResponse `json:"items"`
}

// CallHistoryResponse é um registro do histórico de ligações.
type CallHistoryResponse struct {
	CallID           string `json:"call_id"`
	UniqueID         string `json:"unique_id"`
	LinkedID         string `json:"linked_id,omitempty"`
	ProtocolNumber   string `json:"protocol_number,omitempty"`
	CallType         string `json:"call_type"`
	CallResult       string `json:"call_result"`
	Status           string `json:"status"`
	CallerIDNum      string `json:"caller_id_num"`
	CallerIDName     string `json:"caller_id_name,omitempty"`
	CalledNumber     string `json:"called_number"`
	DestinationNum   string `json:"destination_num,omitempty"`
	DIDNumber        string `json:"did_number,omitempty"`
	AgentExtension   string `json:"agent_extension,omitempty"`
	AgentName        string `json:"agent_name,omitempty"`
	AgentAnsweredAt  string `json:"agent_answered_at,omitempty"`
	QueueName        string `json:"queue_name,omitempty"`
	QueueWaitTime    int    `json:"queue_wait_time,omitempty"`
	IsAbandoned      bool   `json:"is_abandoned"`
	AbandonReason    string `json:"abandon_reason,omitempty"`
	TrunkName        string `json:"trunk_name,omitempty"`
	RouteType        string `json:"route_type,omitempty"`
	HangupCause      string `json:"hangup_cause,omitempty"`
	StartTime        string `json:"start_time"`
	AnswerTime       string `json:"answer_time,omitempty"`
	EndTime          string `json:"end_time,omitempty"`
	Duration         int    `json:"duration"`
	BillableSec      int    `json:"billable_sec"`
	RecordingEnabled bool   `json:"recording_enabled"`
}

// ListCallHistoryResponse é o envelope paginado do histórico de ligações.
type ListCallHistoryResponse struct {
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Items    []CallHistoryResponse `json:"items"`
}

// OriginateCallRequest são os dados do click-to-call.
type OriginateCallRequest struct {
	// FromExtension é o ramal de origem (toca primeiro). Obrigatório.
	FromExtension string `json:"from_extension"`
	// ToNumber é o número de destino. Obrigatório.
	ToNumber string `json:"to_number"`
}

// OriginateCallResponse é o retorno da originação de chamada.
type OriginateCallResponse struct {
	ActionID  string `json:"action_id"`
	Channel   string `json:"channel,omitempty"`
	Extension string `json:"extension"`
	ToNumber  string `json:"to_number"`
}

// TelephonyTempLinkResponse é um link temporário de download de gravação.
// Prefixado com o domínio (Telephony) para não colidir com OmniTempLinkResponse.
type TelephonyTempLinkResponse struct {
	URL       string `json:"url"`
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	ExpiresAt string `json:"expires_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Interface
// ─────────────────────────────────────────────────────────────────────────────

// ListExtensionsParams filtra a listagem de ramais.
type ListExtensionsParams struct {
	// Search busca por número ou nome do ramal.
	Search string
	// Page é a página (padrão: 1).
	Page int
	// PageSize é a quantidade por página (padrão: 10, máximo: 100).
	PageSize int
}

// ListCallHistoryParams filtra o histórico de ligações.
// DateFrom e DateTo são obrigatórios (período máximo de 3 meses).
type ListCallHistoryParams struct {
	// DateFrom é o início do período (ex.: 2026-07-01). Obrigatório.
	DateFrom string
	// DateTo é o fim do período (ex.: 2026-07-09). Obrigatório.
	DateTo string
	// CallerIDNum filtra por número de origem.
	CallerIDNum string
	// CalledNumber filtra por número chamado.
	CalledNumber string
	// Extension filtra por ramal.
	Extension string
	// CallResult filtra por resultado: ANSWERED, NO ANSWER, BUSY, FAILED.
	CallResult string
	// CallType filtra por tipo de chamada (ex.: inbound, outbound, internal).
	CallType string
	// Page é a página (padrão: 1).
	Page int
	// PageSize é a quantidade por página (padrão: 10, máximo: 100).
	PageSize int
}

// TelephonyCase expõe ramais, histórico de ligações, click-to-call e gravações.
type TelephonyCase interface {
	// ListExtensions lista os ramais com status de registro SIP.
	ListExtensions(ctx context.Context, params ListExtensionsParams) (*ListExtensionsResponse, error)
	// ListCallHistory lista o histórico paginado de ligações do período.
	ListCallHistory(ctx context.Context, params ListCallHistoryParams) (*ListCallHistoryResponse, error)
	// OriginateCall origina uma chamada (click-to-call): o ramal toca primeiro
	// e, ao atender, conecta ao número de destino.
	OriginateCall(ctx context.Context, req OriginateCallRequest) (*OriginateCallResponse, error)
	// CreateRecordingTempLink gera link temporário de download da gravação de uma ligação.
	// expiresIn em segundos (0 = padrão do servidor, 3600).
	CreateRecordingTempLink(ctx context.Context, callID string, expiresIn int) (*TelephonyTempLinkResponse, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Implementação
// ─────────────────────────────────────────────────────────────────────────────

type telephonyClient struct{ http *httpClient }

func newTelephonyClient(hc *httpClient) TelephonyCase { return &telephonyClient{http: hc} }

func (t *telephonyClient) ListExtensions(ctx context.Context, params ListExtensionsParams) (*ListExtensionsResponse, error) {
	q := pageQuery(params.Page, params.PageSize)
	if params.Search != "" {
		q.Set("search", params.Search)
	}
	var out ListExtensionsResponse
	if err := t.http.get(ctx, pathExtensions, q, &out); err != nil {
		return nil, fmt.Errorf("interage/telephony.ListExtensions: %w", err)
	}
	return &out, nil
}

func (t *telephonyClient) ListCallHistory(ctx context.Context, params ListCallHistoryParams) (*ListCallHistoryResponse, error) {
	q := pageQuery(params.Page, params.PageSize)
	if params.DateFrom != "" {
		q.Set("date_from", params.DateFrom)
	}
	if params.DateTo != "" {
		q.Set("date_to", params.DateTo)
	}
	if params.CallerIDNum != "" {
		q.Set("caller_id_num", params.CallerIDNum)
	}
	if params.CalledNumber != "" {
		q.Set("called_number", params.CalledNumber)
	}
	if params.Extension != "" {
		q.Set("extension", params.Extension)
	}
	if params.CallResult != "" {
		q.Set("call_result", params.CallResult)
	}
	if params.CallType != "" {
		q.Set("call_type", params.CallType)
	}
	var out ListCallHistoryResponse
	if err := t.http.get(ctx, pathCallHistory, q, &out); err != nil {
		return nil, fmt.Errorf("interage/telephony.ListCallHistory: %w", err)
	}
	return &out, nil
}

func (t *telephonyClient) OriginateCall(ctx context.Context, req OriginateCallRequest) (*OriginateCallResponse, error) {
	var out OriginateCallResponse
	if err := t.http.post(ctx, pathOriginate, nil, req, &out); err != nil {
		return nil, fmt.Errorf("interage/telephony.OriginateCall: %w", err)
	}
	return &out, nil
}

func (t *telephonyClient) CreateRecordingTempLink(ctx context.Context, callID string, expiresIn int) (*TelephonyTempLinkResponse, error) {
	q := url.Values{}
	if expiresIn > 0 {
		q.Set("expires_in", strconv.Itoa(expiresIn))
	}
	var out TelephonyTempLinkResponse
	path := fmt.Sprintf(pathRecordingTempLink, url.PathEscape(callID))
	if err := t.http.post(ctx, path, q, nil, &out); err != nil {
		return nil, fmt.Errorf("interage/telephony.CreateRecordingTempLink: %w", err)
	}
	return &out, nil
}
