package interage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"mime/multipart"
	"net/http"
)

// CampaignsService cobre as campanhas de disparo (/campaigns).
type CampaignsService struct{ c *httpClient }

// ListCampaignsParams filtra a lista de campanhas.
type ListCampaignsParams struct {
	// Status: pending, processing, ready, scheduled, running, paused, completed, canceled ou failed.
	Status string `query:"status"`
	// Search: nome ou id.
	Search string `query:"search"`
	Paging
}

// ListCampaignContactsParams filtra o resultado por contato.
type ListCampaignContactsParams struct {
	// Status: pending, sent, delivered, read, replied, failed ou skipped.
	Status string `query:"status"`
	Paging
}

// CreateCampaignOptions são as opções da criação.
type CreateCampaignOptions struct {
	// IdempotencyKey: com a mesma chave, repetir a criação devolve a campanha já criada (200)
	// em vez de criar outra. Use uma chave por campanha (ex.: um uuid) e repita-a ao tentar de novo.
	IdempotencyKey string
}

func (o *CreateCampaignOptions) headers() map[string]string {
	if o == nil || o.IdempotencyKey == "" {
		return nil
	}
	return map[string]string{"Idempotency-Key": o.IdempotencyKey}
}

// List devolve uma página de campanhas.
func (s *CampaignsService) List(ctx context.Context, p ListCampaignsParams) (*Page[Campaign], error) {
	return lista[Campaign](ctx, s.c, rotaCampanhas.caminho(), montarQuery(p))
}

// All percorre todas as campanhas do filtro.
func (s *CampaignsService) All(ctx context.Context, p ListCampaignsParams) iter.Seq2[Campaign, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[Campaign], error) {
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// Get devolve o detalhe da campanha.
func (s *CampaignsService) Get(ctx context.Context, id string) (*CampaignDetail, error) {
	var out CampaignDetail
	if err := s.c.get(ctx, rotaCampanha.caminho(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create cria a campanha com os contatos de grupos (GroupIDs).
func (s *CampaignsService) Create(ctx context.Context, in CampaignCreateRequest, opts *CreateCampaignOptions) (*CampaignDetail, error) {
	var out CampaignDetail
	if err := s.c.json(ctx, rotaCriarCampanha.method, rotaCriarCampanha.caminho(), nil, in, &out, opts.headers()); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateFromCSV cria a campanha com os contatos de um CSV (até 20 MB e 200.000 linhas).
// in.CollisionPolicy é obrigatório e in.GroupIDs não vai.
func (s *CampaignsService) CreateFromCSV(ctx context.Context, in CampaignCreateRequest, fileName string, csv io.Reader, opts *CreateCampaignOptions) (*CampaignDetail, error) {
	campanha, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("interage: erro ao serializar a campanha: %w", err)
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if err := w.WriteField("campaign", string(campanha)); err != nil {
		return nil, fmt.Errorf("interage: erro ao montar o formulário: %w", err)
	}
	fw, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("interage: erro ao anexar o CSV: %w", err)
	}
	if _, err := io.Copy(fw, csv); err != nil {
		return nil, fmt.Errorf("interage: erro ao ler o CSV: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("interage: erro ao fechar o formulário: %w", err)
	}
	var out CampaignDetail
	p := pedido{method: http.MethodPost, path: rotaCriarCampanha.caminho(), body: &buf, contentType: w.FormDataContentType(), headers: opts.headers()}
	if _, err := s.c.do(ctx, p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *CampaignsService) acao(ctx context.Context, r rota, id string) (*CampaignDetail, error) {
	var out CampaignDetail
	if err := s.c.json(ctx, r.method, r.caminho(id), nil, nil, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// Start inicia o disparo.
func (s *CampaignsService) Start(ctx context.Context, id string) (*CampaignDetail, error) {
	return s.acao(ctx, rotaIniciarCampanha, id)
}

// Pause pausa o disparo.
func (s *CampaignsService) Pause(ctx context.Context, id string) (*CampaignDetail, error) {
	return s.acao(ctx, rotaPausarCampanha, id)
}

// Cancel cancela a campanha (definitivo).
func (s *CampaignsService) Cancel(ctx context.Context, id string) (*CampaignDetail, error) {
	return s.acao(ctx, rotaCancelarCampanha, id)
}

// Delete remove a campanha.
func (s *CampaignsService) Delete(ctx context.Context, id string) error {
	return s.c.json(ctx, rotaExcluirCampanha.method, rotaExcluirCampanha.caminho(id), nil, nil, nil, nil)
}

// ListContacts devolve uma página do resultado por contato (status, horários, erro, resposta).
func (s *CampaignsService) ListContacts(ctx context.Context, id string, p ListCampaignContactsParams) (*Page[CampaignContact], error) {
	return lista[CampaignContact](ctx, s.c, rotaContatosCampanha.caminho(id), montarQuery(p))
}

// AllContacts percorre todo o resultado por contato.
func (s *CampaignsService) AllContacts(ctx context.Context, id string, p ListCampaignContactsParams) iter.Seq2[CampaignContact, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[CampaignContact], error) {
		p.Cursor = cursor
		return s.ListContacts(ctx, id, p)
	})
}
