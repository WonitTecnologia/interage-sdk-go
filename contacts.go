package interage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ─────────────────────────────────────────────────────────────────────────────
// Modelos
// ─────────────────────────────────────────────────────────────────────────────

// ContactIdentityResponse é uma identidade de canal do contato
// (telefone, username, chat_id, etc.).
type ContactIdentityResponse struct {
	// Channel: phone | whatsapp | sms | telegram | instagram | facebook | sip
	Channel string `json:"channel"`
	// IDType depende do canal: phone | username | chat_id | handle | extension | profile_id
	IDType string `json:"id_type"`
	// IDValue valor do identificador (ex.: número E.164 para canais de telefone).
	IDValue string `json:"id_value"`
	// IsPrimary marca a identidade principal do contato.
	IsPrimary bool `json:"is_primary"`
}

// ContactResponse é um contato da central de contatos do tenant.
type ContactResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CPF         string `json:"cpf,omitempty"`
	Email       string `json:"email,omitempty"`
	Company     string `json:"company,omitempty"`
	ReplaceName bool   `json:"replace_name"`
	// Visibility: public | private
	Visibility string `json:"visibility"`
	// CustomInfo valores dos campos personalizados do tenant (JSON bruto).
	CustomInfo json.RawMessage `json:"custom_info,omitempty"`
	// Identities formas de contato (telefones, usernames, etc.).
	Identities []ContactIdentityResponse `json:"identities"`
	CreatedAt  string                    `json:"created_at"`
	UpdatedAt  string                    `json:"updated_at"`
}

// ListContactsResponse é o envelope paginado da listagem de contatos.
type ListContactsResponse struct {
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Items    []ContactResponse `json:"items"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Interface
// ─────────────────────────────────────────────────────────────────────────────

// ListContactsParams filtra a listagem de contatos.
type ListContactsParams struct {
	// Search é a busca geral: nome, empresa, identidades e campos personalizados.
	Search string
	// Name filtra por nome/empresa (ignorado se Search informado).
	Name string
	// Page é a página (padrão: 1).
	Page int
	// PageSize é a quantidade por página (padrão: 10, máximo: 100).
	PageSize int
}

// ContactsCase expõe a central de contatos do tenant.
type ContactsCase interface {
	// ListContacts lista os contatos do tenant, paginado.
	ListContacts(ctx context.Context, params ListContactsParams) (*ListContactsResponse, error)
	// GetContact retorna um contato pelo UUID.
	// Retorna ErrNotFound quando o contato não existe.
	GetContact(ctx context.Context, id string) (*ContactResponse, error)
	// GetContactByPhone retorna um contato pelo número de telefone
	// (normalizado para dígitos; busca nos canais phone, whatsapp e sms).
	// Retorna ErrNotFound quando não encontra.
	GetContactByPhone(ctx context.Context, phone string) (*ContactResponse, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Implementação
// ─────────────────────────────────────────────────────────────────────────────

type contactsClient struct{ http *httpClient }

func newContactsClient(hc *httpClient) ContactsCase { return &contactsClient{http: hc} }

func (c *contactsClient) ListContacts(ctx context.Context, params ListContactsParams) (*ListContactsResponse, error) {
	q := pageQuery(params.Page, params.PageSize)
	if params.Search != "" {
		q.Set("search", params.Search)
	}
	if params.Name != "" {
		q.Set("name", params.Name)
	}
	var out ListContactsResponse
	if err := c.http.get(ctx, pathContacts, q, &out); err != nil {
		return nil, fmt.Errorf("interage/contacts.ListContacts: %w", err)
	}
	return &out, nil
}

func (c *contactsClient) GetContact(ctx context.Context, id string) (*ContactResponse, error) {
	var out ContactResponse
	path := fmt.Sprintf(pathContactByID, url.PathEscape(id))
	if err := c.http.get(ctx, path, nil, &out); err != nil {
		return nil, fmt.Errorf("interage/contacts.GetContact: %w", err)
	}
	return &out, nil
}

func (c *contactsClient) GetContactByPhone(ctx context.Context, phone string) (*ContactResponse, error) {
	q := url.Values{}
	q.Set("phone", phone)
	var out ContactResponse
	if err := c.http.get(ctx, pathContactByPhone, q, &out); err != nil {
		return nil, fmt.Errorf("interage/contacts.GetContactByPhone: %w", err)
	}
	return &out, nil
}
