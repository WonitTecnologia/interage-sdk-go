package interage

import (
	"context"
	"fmt"
	"net/url"

	"github.com/WonitTecnologia/interage-sdk-go/models/messages"
)

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
	ListInstances(ctx context.Context, page, pageSize int) (*messages.ListInstancesResponse, error)
	// ListTemplates lista os templates HSM do tenant.
	// Use o gupshup_id como template_id nos envios e campanhas.
	ListTemplates(ctx context.Context, params ListTemplatesParams) (*messages.ListTemplatesResponse, error)
	// SendTemplate envia um template HSM aprovado para um destinatário.
	SendTemplate(ctx context.Context, req messages.SendTemplateRequest) (*messages.SendTemplateResponse, error)
	// SendMessage envia uma mensagem não-template sobre uma conversa com sessão ativa de 24h.
	SendMessage(ctx context.Context, req messages.SendMessageRequest) (*messages.SendMessageResponse, error)
	// GetMessageStatus consulta o status de entrega pelo internal_message_id.
	GetMessageStatus(ctx context.Context, messageID string) (*messages.MessageStatusResponse, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Implementação
// ─────────────────────────────────────────────────────────────────────────────

type messagesClient struct{ http *httpClient }

func newMessagesClient(hc *httpClient) MessagesCase { return &messagesClient{http: hc} }

func (m *messagesClient) ListInstances(ctx context.Context, page, pageSize int) (*messages.ListInstancesResponse, error) {
	var out messages.ListInstancesResponse
	if err := m.http.get(ctx, pathInstances, pageQuery(page, pageSize), &out); err != nil {
		return nil, fmt.Errorf("interage/messages.ListInstances: %w", err)
	}
	return &out, nil
}

func (m *messagesClient) ListTemplates(ctx context.Context, params ListTemplatesParams) (*messages.ListTemplatesResponse, error) {
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
	var out messages.ListTemplatesResponse
	if err := m.http.get(ctx, pathTemplates, q, &out); err != nil {
		return nil, fmt.Errorf("interage/messages.ListTemplates: %w", err)
	}
	return &out, nil
}

func (m *messagesClient) SendTemplate(ctx context.Context, req messages.SendTemplateRequest) (*messages.SendTemplateResponse, error) {
	var out messages.SendTemplateResponse
	if err := m.http.post(ctx, pathTemplateSend, nil, req, &out); err != nil {
		return nil, fmt.Errorf("interage/messages.SendTemplate: %w", err)
	}
	return &out, nil
}

func (m *messagesClient) SendMessage(ctx context.Context, req messages.SendMessageRequest) (*messages.SendMessageResponse, error) {
	var out messages.SendMessageResponse
	if err := m.http.post(ctx, pathMessageSend, nil, req, &out); err != nil {
		return nil, fmt.Errorf("interage/messages.SendMessage: %w", err)
	}
	return &out, nil
}

func (m *messagesClient) GetMessageStatus(ctx context.Context, messageID string) (*messages.MessageStatusResponse, error) {
	q := url.Values{}
	q.Set("message_id", messageID)
	var out messages.MessageStatusResponse
	if err := m.http.get(ctx, pathMessageStatus, q, &out); err != nil {
		return nil, fmt.Errorf("interage/messages.GetMessageStatus: %w", err)
	}
	return &out, nil
}
