package interage

import (
	"context"
	"iter"
)

// InstancesService cobre os canais (/instances).
type InstancesService struct{ c *httpClient }

// ListInstancesParams filtra a lista de instâncias.
type ListInstancesParams struct {
	// ChannelType: whatsapp, instagram, facebook ou telegram.
	ChannelType string `query:"channel_type"`
	// Active: só as ativas (true) ou só as inativas (false); nil = todas.
	Active *bool `query:"active"`
	Paging
}

// List devolve uma página de instâncias.
func (s *InstancesService) List(ctx context.Context, p ListInstancesParams) (*Page[Instance], error) {
	return lista[Instance](ctx, s.c, rotaInstancias.caminho(), montarQuery(p))
}

// All percorre todas as instâncias do filtro.
func (s *InstancesService) All(ctx context.Context, p ListInstancesParams) iter.Seq2[Instance, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[Instance], error) {
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// TemplatesService cobre os templates de mensagem (/templates).
type TemplatesService struct{ c *httpClient }

// ListTemplatesParams filtra a lista de templates.
type ListTemplatesParams struct {
	InstanceID string `query:"instance_id"`
	// Status: o da Meta (ex.: APPROVED).
	Status string `query:"status"`
	// Category: a da Meta (ex.: UTILITY).
	Category string `query:"category"`
	// Type: tipo do template (ex.: TEXT, IMAGE).
	Type string `query:"type"`
	Paging
}

// List devolve uma página de templates.
func (s *TemplatesService) List(ctx context.Context, p ListTemplatesParams) (*Page[Template], error) {
	return lista[Template](ctx, s.c, rotaTemplates.caminho(), montarQuery(p))
}

// All percorre todos os templates do filtro.
func (s *TemplatesService) All(ctx context.Context, p ListTemplatesParams) iter.Seq2[Template, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[Template], error) {
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// QueuesService cobre as filas de chat e de voz (/queues).
type QueuesService struct{ c *httpClient }

// ListQueuesParams filtra a lista de filas.
type ListQueuesParams struct {
	// Channel: chat ou voice.
	Channel string `query:"channel"`
	Active  *bool  `query:"active"`
	Paging
}

// List devolve uma página de filas.
func (s *QueuesService) List(ctx context.Context, p ListQueuesParams) (*Page[Queue], error) {
	return lista[Queue](ctx, s.c, rotaFilas.caminho(), montarQuery(p))
}

// All percorre todas as filas do filtro.
func (s *QueuesService) All(ctx context.Context, p ListQueuesParams) iter.Seq2[Queue, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[Queue], error) {
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// UsersService cobre os usuários (/users).
type UsersService struct{ c *httpClient }

// ListUsersParams filtra a lista de usuários.
type ListUsersParams struct {
	// Search: trecho do nome ou do e-mail.
	Search string `query:"search"`
	// ChatStatus e VoiceStatus: online, paused ou offline.
	ChatStatus  string `query:"chat_status"`
	VoiceStatus string `query:"voice_status"`
	// QueueID: fila de chat ou de voz (GET /queues).
	QueueID string `query:"queue_id"`
	Paging
}

// List devolve uma página de usuários.
func (s *UsersService) List(ctx context.Context, p ListUsersParams) (*Page[User], error) {
	return lista[User](ctx, s.c, rotaUsuarios.caminho(), montarQuery(p))
}

// All percorre todos os usuários do filtro.
func (s *UsersService) All(ctx context.Context, p ListUsersParams) iter.Seq2[User, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[User], error) {
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// Get devolve o usuário, com as filas dinâmicas de chat e de voz.
func (s *UsersService) Get(ctx context.Context, id string) (*UserDetail, error) {
	var out UserDetail
	if err := s.c.get(ctx, rotaUsuario.caminho(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
