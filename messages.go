package interage

import (
	"context"
	"fmt"
	"net/url"
)

// ─────────────────────────────────────────────────────────────────────────────
// Modelos
// ─────────────────────────────────────────────────────────────────────────────

// InstanceResponse é uma instância de canal WhatsApp do tenant.
type InstanceResponse struct {
	ID          string `json:"id"`
	InstanceID  string `json:"instance_id"`
	Name        string `json:"name"`
	ChannelType string `json:"channel_type"`
	Source      string `json:"source"`
	Active      bool   `json:"active"`
	Receptive   bool   `json:"receptive"`
}

// ListInstancesResponse é o envelope paginado da listagem de instâncias.
type ListInstancesResponse struct {
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Items    []InstanceResponse `json:"items"`
}

// TemplateResponse é um template HSM aprovado.
type TemplateResponse struct {
	ID             int    `json:"id"`
	GupshupID      string `json:"gupshup_id"`
	ElementName    string `json:"element_name"`
	Content        string `json:"content"`
	Header         string `json:"header,omitempty"`
	Footer         string `json:"footer,omitempty"`
	Category       string `json:"category"`
	TemplateType   string `json:"template_type"`
	Status         string `json:"status"`
	LanguageCode   string `json:"language_code"`
	MediaURL       string `json:"media_url,omitempty"`
	ParamsCount    int    `json:"params_count"`
	Visibility     string `json:"visibility"`
	InstanceID     string `json:"instance_id"`
	InstanceName   string `json:"instance_name"`
	InstanceSource string `json:"instance_source"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ListTemplatesResponse é o envelope paginado da listagem de templates.
type ListTemplatesResponse struct {
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Items    []TemplateResponse `json:"items"`
}

// SendTemplateRequest são os dados para envio de um template HSM.
type SendTemplateRequest struct {
	// To é o número do destinatário (ex.: 5547999999999). Obrigatório.
	To string `json:"to"`
	// TemplateID é o gupshup_id do template (ListTemplates). Obrigatório.
	TemplateID string `json:"template_id"`
	// InstanceID é o instance_id da instância do template. Obrigatório.
	InstanceID string `json:"instance_id"`
	// Params são os valores das variáveis do template, na ordem.
	Params []string `json:"params,omitempty"`
	// ContactName é o nome do contato (para criação/atualização do cadastro).
	ContactName string `json:"contact_name,omitempty"`
	// MediaURL é a URL da mídia (para templates com header de mídia).
	MediaURL string `json:"media_url,omitempty"`
	// MediaType é o tipo da mídia: image | video | document.
	MediaType string `json:"media_type,omitempty"`
	// MediaFilename é o nome do arquivo da mídia (para document).
	MediaFilename string `json:"media_filename,omitempty"`
	// AssignedUserID, quando informado, cria um atendimento outbound atribuído ao agente.
	AssignedUserID string `json:"assigned_user_id,omitempty"`
}

// SendTemplateResponse é o retorno do envio de template.
type SendTemplateResponse struct {
	MessageID         string `json:"message_id"`
	InternalMessageID *int64 `json:"internal_message_id,omitempty"`
	ConversationID    *int64 `json:"conversation_id,omitempty"`
	Protocol          string `json:"protocol,omitempty"`
}

// SendMessageRequest são os dados para envio de mensagem não-template
// sobre uma conversa com sessão ativa de 24h.
type SendMessageRequest struct {
	// Protocol é o protocolo da conversa ativa. Obrigatório.
	Protocol string `json:"protocol"`
	// Type é o tipo da mensagem: text | image | video | file | audio | sticker | location. Obrigatório.
	Type string `json:"type"`
	// Text é o conteúdo textual (para type=text) ou legenda.
	Text string `json:"text,omitempty"`
	// URL é a URL da mídia (para tipos de mídia).
	URL string `json:"url,omitempty"`
	// Caption é a legenda da mídia.
	Caption string `json:"caption,omitempty"`
	// Filename é o nome do arquivo (para type=file).
	Filename string `json:"filename,omitempty"`
	// Latitude/Longitude/Address/LocationName são usados em type=location.
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
	Address      string  `json:"address,omitempty"`
	LocationName string  `json:"location_name,omitempty"`
}

// SendMessageResponse é o retorno do envio de mensagem.
type SendMessageResponse struct {
	MessageID         string `json:"message_id"`
	InternalMessageID int64  `json:"internal_message_id"`
}

// MessageStatusResponse é o status de entrega de uma mensagem enviada.
type MessageStatusResponse struct {
	ID           int64  `json:"id"`
	Status       string `json:"status"`
	SentAt       string `json:"sent_at,omitempty"`
	DeliveredAt  string `json:"delivered_at,omitempty"`
	ReadAt       string `json:"read_at,omitempty"`
	FailedAt     string `json:"failed_at,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorDetails string `json:"error_details,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Interface
// ─────────────────────────────────────────────────────────────────────────────

// ListTemplatesParams filtra a listagem de templates HSM.
type ListTemplatesParams struct {
	// Status filtra por status do template (ex.: APPROVED).
	Status string
	// Category filtra por categoria (ex.: MARKETING, UTILITY).
	Category string
	// Type filtra por tipo do template (ex.: TEXT, IMAGE).
	Type string
	// InstanceID filtra templates de uma instância específica.
	InstanceID string
	// Page é a página (padrão: 1).
	Page int
	// PageSize é a quantidade por página (padrão: 10, máximo: 100).
	PageSize int
}

// MessagesCase expõe instâncias, templates HSM e envio de mensagens WhatsApp.
type MessagesCase interface {
	// ListInstances lista as instâncias de canal WhatsApp do tenant.
	// Use o campo instance_id nas criações de campanha e envio de template.
	ListInstances(ctx context.Context, page, pageSize int) (*ListInstancesResponse, error)
	// ListTemplates lista os templates HSM do tenant.
	// Use o gupshup_id como template_id nos envios e campanhas.
	ListTemplates(ctx context.Context, params ListTemplatesParams) (*ListTemplatesResponse, error)
	// SendTemplate envia um template HSM aprovado para um destinatário.
	SendTemplate(ctx context.Context, req SendTemplateRequest) (*SendTemplateResponse, error)
	// SendMessage envia uma mensagem não-template sobre uma conversa com sessão ativa de 24h.
	SendMessage(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error)
	// GetMessageStatus consulta o status de entrega pelo internal_message_id.
	GetMessageStatus(ctx context.Context, messageID string) (*MessageStatusResponse, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Implementação
// ─────────────────────────────────────────────────────────────────────────────

type messagesClient struct{ http *httpClient }

func newMessagesClient(hc *httpClient) MessagesCase { return &messagesClient{http: hc} }

func (m *messagesClient) ListInstances(ctx context.Context, page, pageSize int) (*ListInstancesResponse, error) {
	var out ListInstancesResponse
	if err := m.http.get(ctx, pathInstances, pageQuery(page, pageSize), &out); err != nil {
		return nil, fmt.Errorf("interage/messages.ListInstances: %w", err)
	}
	return &out, nil
}

func (m *messagesClient) ListTemplates(ctx context.Context, params ListTemplatesParams) (*ListTemplatesResponse, error) {
	q := pageQuery(params.Page, params.PageSize)
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.Category != "" {
		q.Set("category", params.Category)
	}
	if params.Type != "" {
		q.Set("type", params.Type)
	}
	if params.InstanceID != "" {
		q.Set("instance_id", params.InstanceID)
	}
	var out ListTemplatesResponse
	if err := m.http.get(ctx, pathTemplates, q, &out); err != nil {
		return nil, fmt.Errorf("interage/messages.ListTemplates: %w", err)
	}
	return &out, nil
}

func (m *messagesClient) SendTemplate(ctx context.Context, req SendTemplateRequest) (*SendTemplateResponse, error) {
	var out SendTemplateResponse
	if err := m.http.post(ctx, pathTemplateSend, nil, req, &out); err != nil {
		return nil, fmt.Errorf("interage/messages.SendTemplate: %w", err)
	}
	return &out, nil
}

func (m *messagesClient) SendMessage(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error) {
	var out SendMessageResponse
	if err := m.http.post(ctx, pathMessageSend, nil, req, &out); err != nil {
		return nil, fmt.Errorf("interage/messages.SendMessage: %w", err)
	}
	return &out, nil
}

func (m *messagesClient) GetMessageStatus(ctx context.Context, messageID string) (*MessageStatusResponse, error) {
	q := url.Values{}
	q.Set("message_id", messageID)
	var out MessageStatusResponse
	if err := m.http.get(ctx, pathMessageStatus, q, &out); err != nil {
		return nil, fmt.Errorf("interage/messages.GetMessageStatus: %w", err)
	}
	return &out, nil
}
