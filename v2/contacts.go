package interage

import (
	"context"
	"iter"
)

// ContactsService cobre os contatos (/contacts).
type ContactsService struct{ c *httpClient }

// ListContactsParams filtra a lista de contatos.
type ListContactsParams struct {
	// Search: busca geral (nome, empresa, e-mail, telefone).
	Search string `query:"search"`
	// Name: nome ou empresa (contém).
	Name string `query:"name"`
	// Phone: número de telefone.
	Phone string `query:"phone"`
	// Group: id do grupo de contatos.
	Group int `query:"group"`
	Paging
}

// List devolve uma página de contatos.
func (s *ContactsService) List(ctx context.Context, p ListContactsParams) (*Page[Contact], error) {
	return lista[Contact](ctx, s.c, rotaContatos.caminho(), montarQuery(p))
}

// All percorre todos os contatos do filtro, página a página.
func (s *ContactsService) All(ctx context.Context, p ListContactsParams) iter.Seq2[Contact, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[Contact], error) {
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// Get devolve o contato pelo id.
func (s *ContactsService) Get(ctx context.Context, id string) (*Contact, error) {
	var out Contact
	if err := s.c.get(ctx, rotaContato.caminho(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create cria o contato. Identidade já usada em outro contato: CONTACT_ALREADY_EXISTS.
func (s *ContactsService) Create(ctx context.Context, in ContactCreate) (*Contact, error) {
	var out Contact
	if err := s.c.json(ctx, rotaCriarContato.method, rotaCriarContato.caminho(), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update altera o contato: só os campos com valor (Set) ou Null mudam.
func (s *ContactsService) Update(ctx context.Context, id string, in ContactUpdate) (*Contact, error) {
	var out Contact
	if err := s.c.json(ctx, rotaAlterarContato.method, rotaAlterarContato.caminho(id), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete exclui o contato.
func (s *ContactsService) Delete(ctx context.Context, id string) error {
	return s.c.json(ctx, rotaExcluirContato.method, rotaExcluirContato.caminho(id), nil, nil, nil, nil)
}

// Batch cria ou atualiza até 500 contatos de uma vez, com a política de colisão.
func (s *ContactsService) Batch(ctx context.Context, in ContactBatchRequest) (*ContactBatchResult, error) {
	var out ContactBatchResult
	if err := s.c.json(ctx, rotaLoteContatos.method, rotaLoteContatos.caminho(), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}
