package interage

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/WonitTecnologia/interage-sdk-go/models/omni"
)

// ─────────────────────────────────────────────────────────────────────────────
// Interface
// ─────────────────────────────────────────────────────────────────────────────

// ListConversationsParams filtra a listagem de conversas.
type ListConversationsParams struct {
	// Status filtra por status da conversa (ex.: bot, waiting, in_progress).
	Status string
	// ChannelType filtra por tipo de canal (ex.: whatsapp, webchat).
	ChannelType string
	// Page é a página (padrão: 1).
	Page int
	// PageSize é a quantidade por página (padrão: 10, máximo: 100).
	PageSize int
}

// OmniCase expõe filas, agentes e conversas do atendimento omnichannel.
type OmniCase interface {
	// ListQueues lista as filas omni com seus agentes e presença.
	ListQueues(ctx context.Context, page, pageSize int) (*omni.ListQueuesResponse, error)
	// ListAgents lista os agentes do tenant com presença em tempo real.
	ListAgents(ctx context.Context, page, pageSize int) (*omni.ListAgentsResponse, error)
	// ListConversations lista as conversas abertas do tenant.
	ListConversations(ctx context.Context, params ListConversationsParams) (*omni.ListConversationsResponse, error)
	// GetConversationHistory retorna as mensagens de uma conversa pelo protocolo, paginado.
	GetConversationHistory(ctx context.Context, protocol string, page, pageSize int) (*omni.ConversationHistoryResponse, error)
	// CloseConversation encerra uma conversa ativa pelo protocolo.
	CloseConversation(ctx context.Context, protocol string) (*omni.CloseConversationResponse, error)
	// TransferConversation transfere a conversa para uma fila ou um agente.
	TransferConversation(ctx context.Context, protocol string, req omni.TransferConversationRequest) (*omni.TransferConversationResponse, error)
	// CreateMessageFileTempLink gera link temporário de download do arquivo de uma mensagem.
	// expiresIn em segundos (0 = padrão do servidor, 3600).
	CreateMessageFileTempLink(ctx context.Context, messageID string, expiresIn int) (*omni.TempLinkResponse, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Implementação
// ─────────────────────────────────────────────────────────────────────────────

type omniClient struct{ http *httpClient }

func newOmniClient(hc *httpClient) OmniCase { return &omniClient{http: hc} }

func (o *omniClient) ListQueues(ctx context.Context, page, pageSize int) (*omni.ListQueuesResponse, error) {
	var out omni.ListQueuesResponse
	if err := o.http.get(ctx, pathOmniQueues, pageQuery(page, pageSize), &out); err != nil {
		return nil, fmt.Errorf("interage/omni.ListQueues: %w", err)
	}
	return &out, nil
}

func (o *omniClient) ListAgents(ctx context.Context, page, pageSize int) (*omni.ListAgentsResponse, error) {
	var out omni.ListAgentsResponse
	if err := o.http.get(ctx, pathOmniAgents, pageQuery(page, pageSize), &out); err != nil {
		return nil, fmt.Errorf("interage/omni.ListAgents: %w", err)
	}
	return &out, nil
}

func (o *omniClient) ListConversations(ctx context.Context, params ListConversationsParams) (*omni.ListConversationsResponse, error) {
	q := pageQuery(params.Page, params.PageSize)
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.ChannelType != "" {
		q.Set("channel_type", params.ChannelType)
	}
	var out omni.ListConversationsResponse
	if err := o.http.get(ctx, pathOmniConversations, q, &out); err != nil {
		return nil, fmt.Errorf("interage/omni.ListConversations: %w", err)
	}
	return &out, nil
}

func (o *omniClient) GetConversationHistory(ctx context.Context, protocol string, page, pageSize int) (*omni.ConversationHistoryResponse, error) {
	var out omni.ConversationHistoryResponse
	path := fmt.Sprintf(pathOmniConversationHistory, url.PathEscape(protocol))
	if err := o.http.get(ctx, path, pageQuery(page, pageSize), &out); err != nil {
		return nil, fmt.Errorf("interage/omni.GetConversationHistory: %w", err)
	}
	return &out, nil
}

func (o *omniClient) CloseConversation(ctx context.Context, protocol string) (*omni.CloseConversationResponse, error) {
	var out omni.CloseConversationResponse
	path := fmt.Sprintf(pathOmniConversationClose, url.PathEscape(protocol))
	if err := o.http.post(ctx, path, nil, nil, &out); err != nil {
		return nil, fmt.Errorf("interage/omni.CloseConversation: %w", err)
	}
	return &out, nil
}

func (o *omniClient) TransferConversation(ctx context.Context, protocol string, req omni.TransferConversationRequest) (*omni.TransferConversationResponse, error) {
	var out omni.TransferConversationResponse
	path := fmt.Sprintf(pathOmniConversationTransfer, url.PathEscape(protocol))
	if err := o.http.post(ctx, path, nil, req, &out); err != nil {
		return nil, fmt.Errorf("interage/omni.TransferConversation: %w", err)
	}
	return &out, nil
}

func (o *omniClient) CreateMessageFileTempLink(ctx context.Context, messageID string, expiresIn int) (*omni.TempLinkResponse, error) {
	q := url.Values{}
	if expiresIn > 0 {
		q.Set("expires_in", strconv.Itoa(expiresIn))
	}
	var out omni.TempLinkResponse
	path := fmt.Sprintf(pathOmniMessageFileTempLink, url.PathEscape(messageID))
	if err := o.http.post(ctx, path, q, nil, &out); err != nil {
		return nil, fmt.Errorf("interage/omni.CreateMessageFileTempLink: %w", err)
	}
	return &out, nil
}
