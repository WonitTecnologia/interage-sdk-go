// Package omni contém os modelos do domínio omnichannel
// (filas, agentes, conversas e links temporários de arquivos).
package omni

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

// TempLinkResponse é um link temporário de download de arquivo.
type TempLinkResponse struct {
	URL       string `json:"url"`
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	ExpiresAt string `json:"expires_at"`
}
