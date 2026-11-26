package interage

import "context"

// Identify devolve quem é o token (tenant, nome, validade, permissões e limite). É sempre
// liberada, mesmo sem nenhuma operação marcada no token: serve para testar a conexão.
func (c *Client) Identify(ctx context.Context) (*Identification, error) {
	var out Identification
	if err := c.http.get(ctx, rotaIdentificacao.caminho(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
