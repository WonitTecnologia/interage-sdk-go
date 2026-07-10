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

// QueueUserResponse é um agente vinculado a uma fila, com presença.
type QueueUserResponse struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	IsOnline    bool   `json:"is_online"`
	IsPaused    bool   `json:"is_paused"`
	PauseReason string `json:"pause_reason,omitempty"`
	PausedAt    string `json:"paused_at,omitempty"`
	LoggedInAt  string `json:"logged_in_at,omitempty"`
}

// QueueWithUsersResponse é uma fila omni com seus agentes.
type QueueWithUsersResponse struct {
	ID          int                 `json:"id"`
	Name        string              `json:"name"`
	ChannelType string              `json:"channel_type"`
	Strategy    string              `json:"strategy"`
	IsActive    bool                `json:"is_active"`
	Users       []QueueUserResponse `json:"users"`
}

// ListQueuesResponse é o envelope paginado da listagem de filas.
type ListQueuesResponse struct {
	Total    int                      `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
	Items    []QueueWithUsersResponse `json:"items"`
}

// AgentStatusResponse é um agente do tenant com status de presença.
type AgentStatusResponse struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	IsOnline    bool   `json:"is_online"`
	IsPaused    bool   `json:"is_paused"`
	PauseReason string `json:"pause_reason,omitempty"`
	PausedAt    string `json:"paused_at,omitempty"`
	LoggedInAt  string `json:"logged_in_at,omitempty"`
}

// ListAgentsResponse é o envelope paginado da listagem de agentes.
type ListAgentsResponse struct {
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Items    []AgentStatusResponse `json:"items"`
}

// ConversationResponse é uma conversa omni ativa.
type ConversationResponse struct {
	ID                int64  `json:"id"`
	Protocol          string `json:"protocol"`
	Status            string `json:"status"`
	ChannelType       string `json:"channel_type"`
	ChannelSource     string `json:"channel_source"`
	ContactName       string `json:"contact_name"`
	ContactPhone      string `json:"contact_phone"`
	QueueID           *int   `json:"queue_id,omitempty"`
	QueuedAt          string `json:"queued_at,omitempty"`
	AssignedAgentID   string `json:"assigned_agent_id,omitempty"`
	AssignedAgentName string `json:"assigned_agent_name,omitempty"`
	AssignedAt        string `json:"assigned_at,omitempty"`
	LastMessage       string `json:"last_message,omitempty"`
	LastMessageAt     string `json:"last_message_at,omitempty"`
	CreatedAt         string `json:"created_at"`
}

// ListConversationsResponse é o envelope paginado da listagem de conversas.
type ListConversationsResponse struct {
	Total    int                    `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Items    []ConversationResponse `json:"items"`
}

// ConversationSummaryResponse são os metadados da conversa no histórico.
type ConversationSummaryResponse struct {
	ID                int64  `json:"id"`
	Protocol          string `json:"protocol"`
	Status            string `json:"status"`
	Orientation       string `json:"orientation,omitempty"`
	ChannelType       string `json:"channel_type"`
	ChannelSource     string `json:"channel_source"`
	ContactName       string `json:"contact_name"`
	ContactPhone      string `json:"contact_phone"`
	QueueID           *int   `json:"queue_id,omitempty"`
	AssignedAgentName string `json:"assigned_agent_name,omitempty"`
	CreatedAt         string `json:"created_at"`
	FinishedAt        string `json:"finished_at,omitempty"`
}

// ConversationMessageResponse é uma mensagem do histórico da conversa.
type ConversationMessageResponse struct {
	ID          int64  `json:"id"`
	Direction   string `json:"direction"`
	SentBy      string `json:"sent_by,omitempty"`
	MessageType string `json:"message_type"`
	Content     string `json:"content,omitempty"`
	Caption     string `json:"caption,omitempty"`
	MediaURL    string `json:"media_url,omitempty"`
	MediaMime   string `json:"media_mime,omitempty"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// ConversationHistoryResponse é o histórico paginado de uma conversa.
type ConversationHistoryResponse struct {
	Conversation ConversationSummaryResponse   `json:"conversation"`
	Total        int                           `json:"total"`
	Page         int                           `json:"page"`
	PageSize     int                           `json:"page_size"`
	Items        []ConversationMessageResponse `json:"items"`
}

// CloseConversationResponse é o retorno do encerramento de conversa.
type CloseConversationResponse struct {
	ID         int64  `json:"id"`
	Protocol   string `json:"protocol"`
	Status     string `json:"status"`
	FinishedAt string `json:"finished_at"`
}

// TransferTargetType é o tipo de destino de uma transferência.
type TransferTargetType string

const (
	// TransferToQueue transfere a conversa para uma fila.
	TransferToQueue TransferTargetType = "queue"
	// TransferToUser transfere a conversa para um agente.
	TransferToUser TransferTargetType = "user"
)

// TransferConversationRequest são os dados de transferência de conversa.
type TransferConversationRequest struct {
	// TargetType é o tipo de destino: queue ou user. Obrigatório.
	TargetType TransferTargetType `json:"target_type"`
	// QueueID é o ID da fila de destino (para TargetType=queue).
	QueueID *int `json:"queue_id,omitempty"`
	// UserID é o ID do agente de destino (para TargetType=user).
	UserID string `json:"user_id,omitempty"`
}

// TransferConversationResponse é o retorno da transferência.
type TransferConversationResponse struct {
	ID         int64  `json:"id"`
	Protocol   string `json:"protocol"`
	Status     string `json:"status"`
	TargetType string `json:"target_type"`
	QueueID    *int   `json:"queue_id,omitempty"`
	UserID     string `json:"user_id,omitempty"`
}

// OmniTempLinkResponse é um link temporário de download de arquivo de mensagem.
// Prefixado com o domínio (Omni) para não colidir com TelephonyTempLinkResponse.
type OmniTempLinkResponse struct {
	URL       string `json:"url"`
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	ExpiresAt string `json:"expires_at"`
}

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
	ListQueues(ctx context.Context, page, pageSize int) (*ListQueuesResponse, error)
	// ListAgents lista os agentes do tenant com presença em tempo real.
	ListAgents(ctx context.Context, page, pageSize int) (*ListAgentsResponse, error)
	// ListConversations lista as conversas abertas do tenant.
	ListConversations(ctx context.Context, params ListConversationsParams) (*ListConversationsResponse, error)
	// GetConversationHistory retorna as mensagens de uma conversa pelo protocolo, paginado.
	GetConversationHistory(ctx context.Context, protocol string, page, pageSize int) (*ConversationHistoryResponse, error)
	// CloseConversation encerra uma conversa ativa pelo protocolo.
	CloseConversation(ctx context.Context, protocol string) (*CloseConversationResponse, error)
	// TransferConversation transfere a conversa para uma fila ou um agente.
	TransferConversation(ctx context.Context, protocol string, req TransferConversationRequest) (*TransferConversationResponse, error)
	// CreateMessageFileTempLink gera link temporário de download do arquivo de uma mensagem.
	// expiresIn em segundos (0 = padrão do servidor, 3600).
	CreateMessageFileTempLink(ctx context.Context, messageID string, expiresIn int) (*OmniTempLinkResponse, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Implementação
// ─────────────────────────────────────────────────────────────────────────────

type omniClient struct{ http *httpClient }

func newOmniClient(hc *httpClient) OmniCase { return &omniClient{http: hc} }

func (o *omniClient) ListQueues(ctx context.Context, page, pageSize int) (*ListQueuesResponse, error) {
	var out ListQueuesResponse
	if err := o.http.get(ctx, pathOmniQueues, pageQuery(page, pageSize), &out); err != nil {
		return nil, fmt.Errorf("interage/omni.ListQueues: %w", err)
	}
	return &out, nil
}

func (o *omniClient) ListAgents(ctx context.Context, page, pageSize int) (*ListAgentsResponse, error) {
	var out ListAgentsResponse
	if err := o.http.get(ctx, pathOmniAgents, pageQuery(page, pageSize), &out); err != nil {
		return nil, fmt.Errorf("interage/omni.ListAgents: %w", err)
	}
	return &out, nil
}

func (o *omniClient) ListConversations(ctx context.Context, params ListConversationsParams) (*ListConversationsResponse, error) {
	q := pageQuery(params.Page, params.PageSize)
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.ChannelType != "" {
		q.Set("channel_type", params.ChannelType)
	}
	var out ListConversationsResponse
	if err := o.http.get(ctx, pathOmniConversations, q, &out); err != nil {
		return nil, fmt.Errorf("interage/omni.ListConversations: %w", err)
	}
	return &out, nil
}

func (o *omniClient) GetConversationHistory(ctx context.Context, protocol string, page, pageSize int) (*ConversationHistoryResponse, error) {
	var out ConversationHistoryResponse
	path := fmt.Sprintf(pathOmniConversationHistory, url.PathEscape(protocol))
	if err := o.http.get(ctx, path, pageQuery(page, pageSize), &out); err != nil {
		return nil, fmt.Errorf("interage/omni.GetConversationHistory: %w", err)
	}
	return &out, nil
}

func (o *omniClient) CloseConversation(ctx context.Context, protocol string) (*CloseConversationResponse, error) {
	var out CloseConversationResponse
	path := fmt.Sprintf(pathOmniConversationClose, url.PathEscape(protocol))
	if err := o.http.post(ctx, path, nil, nil, &out); err != nil {
		return nil, fmt.Errorf("interage/omni.CloseConversation: %w", err)
	}
	return &out, nil
}

func (o *omniClient) TransferConversation(ctx context.Context, protocol string, req TransferConversationRequest) (*TransferConversationResponse, error) {
	var out TransferConversationResponse
	path := fmt.Sprintf(pathOmniConversationTransfer, url.PathEscape(protocol))
	if err := o.http.post(ctx, path, nil, req, &out); err != nil {
		return nil, fmt.Errorf("interage/omni.TransferConversation: %w", err)
	}
	return &out, nil
}

func (o *omniClient) CreateMessageFileTempLink(ctx context.Context, messageID string, expiresIn int) (*OmniTempLinkResponse, error) {
	q := url.Values{}
	if expiresIn > 0 {
		q.Set("expires_in", strconv.Itoa(expiresIn))
	}
	var out OmniTempLinkResponse
	path := fmt.Sprintf(pathOmniMessageFileTempLink, url.PathEscape(messageID))
	if err := o.http.post(ctx, path, q, nil, &out); err != nil {
		return nil, fmt.Errorf("interage/omni.CreateMessageFileTempLink: %w", err)
	}
	return &out, nil
}
