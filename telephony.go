package interage

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/WonitTecnologia/interage-sdk-go/models/telephony"
)

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
	ListExtensions(ctx context.Context, params ListExtensionsParams) (*telephony.ListExtensionsResponse, error)
	// ListCallHistory lista o histórico paginado de ligações do período.
	ListCallHistory(ctx context.Context, params ListCallHistoryParams) (*telephony.ListCallHistoryResponse, error)
	// OriginateCall origina uma chamada (click-to-call): o ramal toca primeiro
	// e, ao atender, conecta ao número de destino.
	OriginateCall(ctx context.Context, req telephony.OriginateCallRequest) (*telephony.OriginateCallResponse, error)
	// CreateRecordingTempLink gera link temporário de download da gravação de uma ligação.
	// expiresIn em segundos (0 = padrão do servidor, 3600).
	CreateRecordingTempLink(ctx context.Context, callID string, expiresIn int) (*telephony.TempLinkResponse, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Implementação
// ─────────────────────────────────────────────────────────────────────────────

type telephonyClient struct{ http *httpClient }

func newTelephonyClient(hc *httpClient) TelephonyCase { return &telephonyClient{http: hc} }

func (t *telephonyClient) ListExtensions(ctx context.Context, params ListExtensionsParams) (*telephony.ListExtensionsResponse, error) {
	q := pageQuery(params.Page, params.PageSize)
	if params.Search != "" {
		q.Set("search", params.Search)
	}
	var out telephony.ListExtensionsResponse
	if err := t.http.get(ctx, pathExtensions, q, &out); err != nil {
		return nil, fmt.Errorf("interage/telephony.ListExtensions: %w", err)
	}
	return &out, nil
}

func (t *telephonyClient) ListCallHistory(ctx context.Context, params ListCallHistoryParams) (*telephony.ListCallHistoryResponse, error) {
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
	var out telephony.ListCallHistoryResponse
	if err := t.http.get(ctx, pathCallHistory, q, &out); err != nil {
		return nil, fmt.Errorf("interage/telephony.ListCallHistory: %w", err)
	}
	return &out, nil
}

func (t *telephonyClient) OriginateCall(ctx context.Context, req telephony.OriginateCallRequest) (*telephony.OriginateCallResponse, error) {
	var out telephony.OriginateCallResponse
	if err := t.http.post(ctx, pathOriginate, nil, req, &out); err != nil {
		return nil, fmt.Errorf("interage/telephony.OriginateCall: %w", err)
	}
	return &out, nil
}

func (t *telephonyClient) CreateRecordingTempLink(ctx context.Context, callID string, expiresIn int) (*telephony.TempLinkResponse, error) {
	q := url.Values{}
	if expiresIn > 0 {
		q.Set("expires_in", strconv.Itoa(expiresIn))
	}
	var out telephony.TempLinkResponse
	path := fmt.Sprintf(pathRecordingTempLink, url.PathEscape(callID))
	if err := t.http.post(ctx, path, q, nil, &out); err != nil {
		return nil, fmt.Errorf("interage/telephony.CreateRecordingTempLink: %w", err)
	}
	return &out, nil
}
