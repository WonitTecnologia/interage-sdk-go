package interage

import (
	"context"
	"encoding/json"
	"errors"
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

// MessageReactionResponse é uma reação (emoji) a uma mensagem, uma por remetente.
type MessageReactionResponse struct {
	Emoji string `json:"emoji"`
	// SentBy: client | bot | ai_agent | agent | agent_coex.
	SentBy string `json:"sent_by"`
}

// ConversationMessageResponse é uma mensagem do histórico da conversa.
type ConversationMessageResponse struct {
	ID        int64  `json:"id"`
	Direction string `json:"direction"`
	// SentBy: client | bot | ai_agent | agent | agent_coex.
	SentBy      string `json:"sent_by,omitempty"`
	MessageType string `json:"message_type"`
	Content     string `json:"content,omitempty"`
	Caption     string `json:"caption,omitempty"`
	// MediaURL é a referência do arquivo, sem credencial. Para baixar, gere o link em
	// CreateMessageFileTempLink.
	MediaURL      string `json:"media_url,omitempty"`
	MediaMime     string `json:"media_mime,omitempty"`
	MediaFileName string `json:"media_file_name,omitempty"`
	Status        string `json:"status"`
	// Edited indica conteúdo editado pelo remetente.
	Edited bool `json:"edited"`
	// RemovedAt vem quando o remetente apagou a mensagem para todos; o conteúdo
	// original permanece.
	RemovedAt string `json:"removed_at,omitempty"`
	// Reactions são as reações à mensagem, uma por remetente.
	Reactions []MessageReactionResponse `json:"reactions,omitempty"`
	CreatedAt string                    `json:"created_at"`
}

// ConversationHistoryResponse é o histórico paginado de uma conversa.
type ConversationHistoryResponse struct {
	Conversation ConversationSummaryResponse   `json:"conversation"`
	Total        int                           `json:"total"`
	Page         int                           `json:"page"`
	PageSize     int                           `json:"page_size"`
	Items        []ConversationMessageResponse `json:"items"`
}

// CloseConversationRequest é o corpo opcional do encerramento de conversa.
type CloseConversationRequest struct {
	// Classification é a classificação do atendimento: obrigatória se a fila exigir e,
	// com opções cadastradas na fila, precisa ser uma delas (senão a API responde 422).
	Classification string `json:"classification,omitempty"`
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
	// QueueID é o ID da fila de destino. Obrigatório com TargetType=queue; com
	// TargetType=user, opcional: troca a fila da conversa ao atribuir ao agente.
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

// BatchContactIdentity é uma identidade de canal de um contato do lote.
// Combinações aceitas: phone→phone; whatsapp→phone|username; telegram→chat_id|username;
// instagram→handle|profile_id; facebook→handle|profile_id; sip→extension.
type BatchContactIdentity struct {
	Channel   string `json:"channel"`
	IDType    string `json:"id_type"`
	IDValue   string `json:"id_value"`
	IsPrimary bool   `json:"is_primary,omitempty"`
}

// BatchContactItem é um contato individual da criação em lote.
type BatchContactItem struct {
	// Name é o nome do contato. Obrigatório.
	Name string `json:"name"`
	// Identities são as formas de contato. Obrigatório ao menos uma.
	Identities []BatchContactIdentity `json:"identities"`
	// CPF do contato (opcional, com ou sem pontuação). Vazio não é enviado; para
	// limpar o CPF de um contato existente com CollisionOverwrite, use ClearCPF.
	CPF string `json:"cpf,omitempty"`
	// ClearCPF envia cpf vazio: com CollisionOverwrite, apaga o CPF do contato existente.
	// Ignorado quando CPF está preenchido.
	ClearCPF bool `json:"-"`
	// Email do contato (opcional).
	Email string `json:"email,omitempty"`
	// Company é a empresa do contato (opcional).
	Company string `json:"company,omitempty"`
	// CustomInfo são os valores dos campos personalizados do tenant
	// (chaves = key dos campos cadastrados e ativos na central; chave desconhecida é
	// erro no item). A API só aceita texto — números e datas vão como string
	// (ex.: "1990-05-20"). Máx. 10 KB.
	CustomInfo map[string]string `json:"custom_info,omitempty"`
	// Labels são os números das etiquetas a atribuir (máx. 20) — o número
	// sequencial exibido na gestão de etiquetas da central, não o nome.
	// ATENÇÃO: a etiqueta precisa existir e estar ativa no tenant — etiqueta
	// inexistente gera erro no item e nada é criado/atualizado para aquele contato.
	Labels []int64 `json:"labels,omitempty"`
	// ReplaceName permite substituir o nome pelo perfil do canal (ex: WhatsApp).
	ReplaceName *bool `json:"replace_name,omitempty"`
}

// MarshalJSON envia "cpf": "" quando ClearCPF está marcado e CPF vazio.
func (b BatchContactItem) MarshalJSON() ([]byte, error) {
	type item BatchContactItem
	if !b.ClearCPF || b.CPF != "" {
		return json.Marshal(item(b))
	}
	return json.Marshal(struct {
		item
		CPF string `json:"cpf"`
	}{item: item(b), CPF: ""})
}

// BatchCreateContactsRequest é o payload da criação de contatos em lote.
type BatchCreateContactsRequest struct {
	// CollisionPolicy define o tratamento quando uma identidade já existe em
	// outro contato do tenant. Opcional (padrão da API: CollisionIgnore).
	CollisionPolicy CollisionPolicy `json:"collision_policy,omitempty"`
	// Contacts lista de contatos a criar (mínimo 1, máximo 500).
	Contacts []BatchContactItem `json:"contacts"`
}

// BatchContactResultItem é o resultado do processamento de um contato do lote.
// Status: created | existing | updated | error.
type BatchContactResultItem struct {
	Index     int    `json:"index"`
	Status    string `json:"status"`
	ContactID string `json:"contact_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

// BatchCreateContactsResponse é o envelope do resultado da criação em lote.
type BatchCreateContactsResponse struct {
	Total    int                      `json:"total"`
	Created  int                      `json:"created"`
	Updated  int                      `json:"updated"`
	Existing int                      `json:"existing"`
	Errors   int                      `json:"errors"`
	Items    []BatchContactResultItem `json:"items"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Interface
// ─────────────────────────────────────────────────────────────────────────────

// ListConversationsParams filtra a listagem de conversas.
type ListConversationsParams struct {
	// Status filtra por status da conversa: bot | ai_agent | queue | attending | coex
	// (vazio = todas as ativas). coex = atendida pelo celular (coexistência).
	Status string
	// ChannelType filtra por tipo de canal: whatsapp | telegram | instagram | facebook | webchat | email.
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
	// CloseConversation encerra uma conversa ativa pelo protocolo, sem classificação.
	// Se a fila exige classificação, use CloseConversationWith.
	CloseConversation(ctx context.Context, protocol string) (*CloseConversationResponse, error)
	// CloseConversationWith encerra a conversa informando a classificação do atendimento.
	CloseConversationWith(ctx context.Context, protocol string, req CloseConversationRequest) (*CloseConversationResponse, error)
	// TransferConversation transfere a conversa para uma fila ou um agente.
	TransferConversation(ctx context.Context, protocol string, req TransferConversationRequest) (*TransferConversationResponse, error)
	// CreateMessageFileTempLink gera link temporário de download do arquivo de uma mensagem.
	// expiresIn em segundos (0 = padrão do servidor, 3600).
	CreateMessageFileTempLink(ctx context.Context, messageID string, expiresIn int) (*OmniTempLinkResponse, error)
	// BatchCreateContacts cria até 500 contatos de uma vez na central de contatos.
	// Cada item é processado de forma independente — ver Items na resposta.
	BatchCreateContacts(ctx context.Context, req BatchCreateContactsRequest) (*BatchCreateContactsResponse, error)
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
	return o.CloseConversationWith(ctx, protocol, CloseConversationRequest{})
}

func (o *omniClient) CloseConversationWith(ctx context.Context, protocol string, req CloseConversationRequest) (*CloseConversationResponse, error) {
	var out CloseConversationResponse
	path := fmt.Sprintf(pathOmniConversationClose, url.PathEscape(protocol))
	var body any
	if req.Classification != "" {
		body = req
	}
	if err := o.http.post(ctx, path, nil, body, &out); err != nil {
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

func (o *omniClient) BatchCreateContacts(ctx context.Context, req BatchCreateContactsRequest) (*BatchCreateContactsResponse, error) {
	if len(req.Contacts) == 0 {
		return nil, errors.New("interage/omni.BatchCreateContacts: informe ao menos um contato")
	}
	if len(req.Contacts) > 500 {
		return nil, errors.New("interage/omni.BatchCreateContacts: lote excede o máximo de 500 contatos")
	}
	var out BatchCreateContactsResponse
	if err := o.http.post(ctx, pathOmniContactsBatch, nil, req, &out); err != nil {
		return nil, fmt.Errorf("interage/omni.BatchCreateContacts: %w", err)
	}
	return &out, nil
}
