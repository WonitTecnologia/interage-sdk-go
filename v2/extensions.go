package interage

import (
	"context"
	"iter"
	"net/url"
)

// ExtensionsService cobre os ramais (/extensions).
type ExtensionsService struct{ c *httpClient }

// ListExtensionsParams filtra a lista de ramais.
type ListExtensionsParams struct {
	// Search: número (começa com) ou nome (contém).
	Search string `query:"search"`
	// Status: offline, in_call, dnd ou online.
	Status string `query:"status"`
	Paging
}

// List devolve uma página de ramais.
func (s *ExtensionsService) List(ctx context.Context, p ListExtensionsParams) (*Page[Extension], error) {
	return lista[Extension](ctx, s.c, rotaRamais.caminho(), montarQuery(p))
}

// All percorre todos os ramais do filtro, página a página.
func (s *ExtensionsService) All(ctx context.Context, p ListExtensionsParams) iter.Seq2[Extension, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[Extension], error) {
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// Get devolve o detalhe do ramal pelo número.
func (s *ExtensionsService) Get(ctx context.Context, number string) (*ExtensionDetail, error) {
	var out ExtensionDetail
	if err := s.c.get(ctx, rotaRamal.caminho(number), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create cria o ramal. A senha SIP só aparece nesta resposta (SipCredentials).
func (s *ExtensionsService) Create(ctx context.Context, in ExtensionCreate) (*ExtensionCreated, error) {
	var out ExtensionCreated
	if err := s.c.json(ctx, rotaCriarRamal.method, rotaCriarRamal.caminho(), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update altera o ramal: só os campos com valor (Set) ou Null mudam.
func (s *ExtensionsService) Update(ctx context.Context, number string, in ExtensionUpdate) (*ExtensionDetail, error) {
	var out ExtensionDetail
	if err := s.c.json(ctx, rotaAlterarRamal.method, rotaAlterarRamal.caminho(number), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete exclui o ramal. force exclui mesmo vinculado a usuário ou fila (EXTENSION_IN_USE sem ele).
func (s *ExtensionsService) Delete(ctx context.Context, number string, force bool) error {
	q := url.Values{}
	if force {
		q.Set("force", "true")
	}
	return s.c.json(ctx, rotaExcluirRamal.method, rotaExcluirRamal.caminho(number), q, nil, nil, nil)
}

// SetDND liga ou desliga o não perturbe.
func (s *ExtensionsService) SetDND(ctx context.Context, number string, enabled bool) (*ExtensionDetail, error) {
	var out ExtensionDetail
	in := ExtensionDndRequest{Enabled: &enabled}
	if err := s.c.json(ctx, rotaDndRamal.method, rotaDndRamal.caminho(number), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetForwarding altera os desvios: Set(destino) liga, Null desliga, ausente fica como está.
func (s *ExtensionsService) SetForwarding(ctx context.Context, number string, in ExtensionForwardingRequest) (*ExtensionDetail, error) {
	var out ExtensionDetail
	if err := s.c.json(ctx, rotaDesviosRamal.method, rotaDesviosRamal.caminho(number), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResetPassword gera uma senha SIP nova. A anterior para na hora; a nova só aparece aqui.
func (s *ExtensionsService) ResetPassword(ctx context.Context, number string) (*SipCredentials, error) {
	var out ExtensionPasswordResponse
	if err := s.c.json(ctx, rotaSenhaRamal.method, rotaSenhaRamal.caminho(number), nil, nil, &out, nil); err != nil {
		return nil, err
	}
	return &out.SIPCredentials, nil
}
