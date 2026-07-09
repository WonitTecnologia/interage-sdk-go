package interage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/WonitTecnologia/interage-sdk-go/models/campaigns"
)

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
	Create(ctx context.Context, req campaigns.CreateCampaignRequest) (*campaigns.CreateCampaignResponse, error)
	// List lista as campanhas do tenant, paginado.
	List(ctx context.Context, params ListCampaignsParams) (*campaigns.ListCampaignsResponse, error)
	// Get retorna uma campanha pelo UUID.
	Get(ctx context.Context, id string) (*campaigns.CampaignResponse, error)
	// Start inicia a campanha (ready|paused|scheduled → running).
	Start(ctx context.Context, id string) (*campaigns.CampaignResponse, error)
	// Pause pausa uma campanha em execução (running → paused).
	Pause(ctx context.Context, id string) (*campaigns.CampaignResponse, error)
	// Cancel cancela a campanha (→ canceled). Cancelada não reinicia.
	Cancel(ctx context.Context, id string) (*campaigns.CampaignResponse, error)
	// Delete remove a campanha. Em execução/processamento não pode ser removida.
	Delete(ctx context.Context, id string) error
}

// ─────────────────────────────────────────────────────────────────────────────
// Implementação
// ─────────────────────────────────────────────────────────────────────────────

type campaignsClient struct{ http *httpClient }

func newCampaignsClient(hc *httpClient) CampaignsCase { return &campaignsClient{http: hc} }

func (c *campaignsClient) Create(ctx context.Context, req campaigns.CreateCampaignRequest) (*campaigns.CreateCampaignResponse, error) {
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

	var out campaigns.CreateCampaignResponse
	if err := c.http.postMultipart(ctx, pathCampaigns, fields, "file", req.FileName, req.FileContent, &out); err != nil {
		return nil, fmt.Errorf("interage/campaigns.Create: %w", err)
	}
	return &out, nil
}

func (c *campaignsClient) List(ctx context.Context, params ListCampaignsParams) (*campaigns.ListCampaignsResponse, error) {
	q := pageQuery(params.Page, params.PageSize)
	if params.Search != "" {
		q.Set("search", params.Search)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	var out campaigns.ListCampaignsResponse
	if err := c.http.get(ctx, pathCampaigns, q, &out); err != nil {
		return nil, fmt.Errorf("interage/campaigns.List: %w", err)
	}
	return &out, nil
}

func (c *campaignsClient) Get(ctx context.Context, id string) (*campaigns.CampaignResponse, error) {
	var out campaigns.CampaignResponse
	if err := c.http.get(ctx, fmt.Sprintf(pathCampaignByID, url.PathEscape(id)), nil, &out); err != nil {
		return nil, fmt.Errorf("interage/campaigns.Get: %w", err)
	}
	return &out, nil
}

func (c *campaignsClient) Start(ctx context.Context, id string) (*campaigns.CampaignResponse, error) {
	var out campaigns.CampaignResponse
	if err := c.http.patch(ctx, fmt.Sprintf(pathCampaignStart, url.PathEscape(id)), nil, &out); err != nil {
		return nil, fmt.Errorf("interage/campaigns.Start: %w", err)
	}
	return &out, nil
}

func (c *campaignsClient) Pause(ctx context.Context, id string) (*campaigns.CampaignResponse, error) {
	var out campaigns.CampaignResponse
	if err := c.http.patch(ctx, fmt.Sprintf(pathCampaignPause, url.PathEscape(id)), nil, &out); err != nil {
		return nil, fmt.Errorf("interage/campaigns.Pause: %w", err)
	}
	return &out, nil
}

func (c *campaignsClient) Cancel(ctx context.Context, id string) (*campaigns.CampaignResponse, error) {
	var out campaigns.CampaignResponse
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
