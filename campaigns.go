package interage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// ─────────────────────────────────────────────────────────────────────────────
// Modelos
// ─────────────────────────────────────────────────────────────────────────────

// CampaignStatus é o status de uma campanha.
type CampaignStatus string

const (
	CampaignStatusPending    CampaignStatus = "pending"    // criada; CSV em importação
	CampaignStatusProcessing CampaignStatus = "processing" // importando contatos
	CampaignStatusReady      CampaignStatus = "ready"      // pronta para iniciar
	CampaignStatusScheduled  CampaignStatus = "scheduled"  // agendada (auto_start futuro)
	CampaignStatusRunning    CampaignStatus = "running"    // em execução
	CampaignStatusPaused     CampaignStatus = "paused"     // pausada (retomável)
	CampaignStatusCompleted  CampaignStatus = "completed"  // concluída
	CampaignStatusCanceled   CampaignStatus = "canceled"   // cancelada (não reinicia)
	CampaignStatusFailed     CampaignStatus = "failed"     // falhou
)

// CollisionPolicy define o que fazer quando o contato do CSV já existe na base.
type CollisionPolicy string

const (
	// CollisionIgnore mantém os dados atuais do contato; não sobrescreve nada.
	CollisionIgnore CollisionPolicy = "ignore"
	// CollisionOverwrite sobrescreve nome/email/empresa com os valores do CSV.
	CollisionOverwrite CollisionPolicy = "overwrite"
	// CollisionUpdateEmpty preenche apenas os campos vazios do contato.
	CollisionUpdateEmpty CollisionPolicy = "update_empty"
)

// CreateCampaignRequest são os dados para criar uma campanha com CSV anexo.
//
// O CSV deve ter delimitador `,` ou `;` (detectado automaticamente) e conter
// uma coluna de telefone (phone, telefone, numero, celular ou whatsapp).
// Colunas opcionais reconhecidas: name, email, company.
type CreateCampaignRequest struct {
	// Name é o nome da campanha. Obrigatório.
	Name string
	// InstanceID é o ID da instância WhatsApp de envio (Messages.ListInstances). Obrigatório.
	InstanceID string
	// TemplateID é o gupshup_id do template HSM aprovado (Messages.ListTemplates). Obrigatório.
	TemplateID string
	// CollisionPolicy define o tratamento de contatos já existentes. Obrigatório.
	CollisionPolicy CollisionPolicy
	// FileName é o nome do arquivo CSV (ex.: contatos.csv). Obrigatório.
	FileName string
	// FileContent é o conteúdo do arquivo CSV. Obrigatório.
	FileContent []byte

	// TemplateParams são os valores das variáveis do template, na ordem em que aparecem.
	TemplateParams []string
	// Description é a descrição da campanha.
	Description string
	// StartAt é a data/hora de início do disparo (RFC3339, ex.: 2026-01-20T09:00:00-03:00).
	StartAt string
	// EndAt é a data/hora limite do disparo (RFC3339). Vazio = dispara até concluir.
	EndAt string
	// AutoStart, quando true, inicia a campanha automaticamente em StartAt.
	AutoStart *bool
	// Settings são configurações extras (ex.: {"delay_ms": 1000}).
	Settings map[string]any
}

// CreateCampaignResponse é o retorno da criação de campanha.
// A importação do CSV é assíncrona; a campanha nasce com status pending.
type CreateCampaignResponse struct {
	CampaignID string `json:"campaign_id"`
	Status     string `json:"status"`
	MailingID  string `json:"mailing_id"`
	FileName   string `json:"file_name"`
}

// CampaignResponse é a visão de uma campanha (listagem e detalhe).
type CampaignResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Channel        string `json:"channel"`
	Status         string `json:"status"`
	TemplateID     string `json:"template_id,omitempty"`
	ScheduledAt    string `json:"scheduled_at,omitempty"`
	StartAt        string `json:"start_at,omitempty"`
	EndAt          string `json:"end_at,omitempty"`
	TotalContacts  int    `json:"total_contacts"`
	SentCount      int    `json:"sent_count"`
	DeliveredCount int    `json:"delivered_count"`
	ReadCount      int    `json:"read_count"`
	RepliedCount   int    `json:"replied_count"`
	FailedCount    int    `json:"failed_count"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ListCampaignsResponse é o envelope paginado da listagem de campanhas.
type ListCampaignsResponse struct {
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Items    []CampaignResponse `json:"items"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Interface
// ─────────────────────────────────────────────────────────────────────────────

// ListCampaignsParams filtra a listagem de campanhas.
type ListCampaignsParams struct {
	// Search busca por nome da campanha.
	Search string
	// Status filtra por status (pending, ready, running, paused, completed, canceled, failed...).
	Status string
	// Page é a página (padrão: 1).
	Page int
	// PageSize é a quantidade por página (padrão: 10, máximo: 100).
	PageSize int
}

// CampaignsCase expõe as operações de campanhas de disparo WhatsApp.
type CampaignsCase interface {
	// Create cria uma campanha com o CSV de contatos anexo (multipart).
	// A importação dos contatos é assíncrona; a campanha nasce com status pending
	// e passa a ready quando o CSV termina de importar.
	Create(ctx context.Context, req CreateCampaignRequest) (*CreateCampaignResponse, error)
	// List lista as campanhas do tenant, paginado.
	List(ctx context.Context, params ListCampaignsParams) (*ListCampaignsResponse, error)
	// Get retorna uma campanha pelo UUID.
	Get(ctx context.Context, id string) (*CampaignResponse, error)
	// Start inicia a campanha (ready|paused|scheduled → running).
	Start(ctx context.Context, id string) (*CampaignResponse, error)
	// Pause pausa uma campanha em execução (running → paused).
	Pause(ctx context.Context, id string) (*CampaignResponse, error)
	// Cancel cancela a campanha (→ canceled). Cancelada não reinicia.
	Cancel(ctx context.Context, id string) (*CampaignResponse, error)
	// Delete remove a campanha. Em execução/processamento não pode ser removida.
	Delete(ctx context.Context, id string) error
}

// ─────────────────────────────────────────────────────────────────────────────
// Implementação
// ─────────────────────────────────────────────────────────────────────────────

type campaignsClient struct{ http *httpClient }

func newCampaignsClient(hc *httpClient) CampaignsCase { return &campaignsClient{http: hc} }

func (c *campaignsClient) Create(ctx context.Context, req CreateCampaignRequest) (*CreateCampaignResponse, error) {
	if req.Name == "" {
		return nil, errors.New("interage/campaigns.Create: Name é obrigatório")
	}
	if req.InstanceID == "" {
		return nil, errors.New("interage/campaigns.Create: InstanceID é obrigatório")
	}
	if req.TemplateID == "" {
		return nil, errors.New("interage/campaigns.Create: TemplateID é obrigatório")
	}
	if req.CollisionPolicy == "" {
		return nil, errors.New("interage/campaigns.Create: CollisionPolicy é obrigatório (ignore, overwrite ou update_empty)")
	}
	if req.FileName == "" || len(req.FileContent) == 0 {
		return nil, errors.New("interage/campaigns.Create: FileName e FileContent são obrigatórios")
	}

	fields := map[string]string{
		"name":             req.Name,
		"channel":          "whatsapp",
		"instance_id":      req.InstanceID,
		"template_id":      req.TemplateID,
		"collision_policy": string(req.CollisionPolicy),
	}
	if len(req.TemplateParams) > 0 {
		b, err := json.Marshal(req.TemplateParams)
		if err != nil {
			return nil, fmt.Errorf("interage/campaigns.Create: erro ao serializar TemplateParams: %w", err)
		}
		fields["template_params"] = string(b)
	}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	if req.StartAt != "" {
		fields["start_at"] = req.StartAt
	}
	if req.EndAt != "" {
		fields["end_at"] = req.EndAt
	}
	if req.AutoStart != nil {
		fields["auto_start"] = fmt.Sprintf("%t", *req.AutoStart)
	}
	if len(req.Settings) > 0 {
		b, err := json.Marshal(req.Settings)
		if err != nil {
			return nil, fmt.Errorf("interage/campaigns.Create: erro ao serializar Settings: %w", err)
		}
		fields["settings"] = string(b)
	}

	var out CreateCampaignResponse
	if err := c.http.postMultipart(ctx, pathCampaigns, fields, "file", req.FileName, req.FileContent, &out); err != nil {
		return nil, fmt.Errorf("interage/campaigns.Create: %w", err)
	}
	return &out, nil
}

func (c *campaignsClient) List(ctx context.Context, params ListCampaignsParams) (*ListCampaignsResponse, error) {
	q := pageQuery(params.Page, params.PageSize)
	if params.Search != "" {
		q.Set("search", params.Search)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	var out ListCampaignsResponse
	if err := c.http.get(ctx, pathCampaigns, q, &out); err != nil {
		return nil, fmt.Errorf("interage/campaigns.List: %w", err)
	}
	return &out, nil
}

func (c *campaignsClient) Get(ctx context.Context, id string) (*CampaignResponse, error) {
	var out CampaignResponse
	if err := c.http.get(ctx, fmt.Sprintf(pathCampaignByID, url.PathEscape(id)), nil, &out); err != nil {
		return nil, fmt.Errorf("interage/campaigns.Get: %w", err)
	}
	return &out, nil
}

func (c *campaignsClient) Start(ctx context.Context, id string) (*CampaignResponse, error) {
	var out CampaignResponse
	if err := c.http.patch(ctx, fmt.Sprintf(pathCampaignStart, url.PathEscape(id)), nil, &out); err != nil {
		return nil, fmt.Errorf("interage/campaigns.Start: %w", err)
	}
	return &out, nil
}

func (c *campaignsClient) Pause(ctx context.Context, id string) (*CampaignResponse, error) {
	var out CampaignResponse
	if err := c.http.patch(ctx, fmt.Sprintf(pathCampaignPause, url.PathEscape(id)), nil, &out); err != nil {
		return nil, fmt.Errorf("interage/campaigns.Pause: %w", err)
	}
	return &out, nil
}

func (c *campaignsClient) Cancel(ctx context.Context, id string) (*CampaignResponse, error) {
	var out CampaignResponse
	if err := c.http.patch(ctx, fmt.Sprintf(pathCampaignCancel, url.PathEscape(id)), nil, &out); err != nil {
		return nil, fmt.Errorf("interage/campaigns.Cancel: %w", err)
	}
	return &out, nil
}

func (c *campaignsClient) Delete(ctx context.Context, id string) error {
	if err := c.http.delete(ctx, fmt.Sprintf(pathCampaignByID, url.PathEscape(id)), nil); err != nil {
		return fmt.Errorf("interage/campaigns.Delete: %w", err)
	}
	return nil
}
