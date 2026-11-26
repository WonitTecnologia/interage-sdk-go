package interage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	userAgent      = "interage-sdk-go/" + Version
)

// httpClient é o transporte comum a todos os recursos: autenticação, corpo JSON, o envelope
// {"data": ...} das respostas e os erros RFC 9457.
type httpClient struct {
	base  string // https://<tenant>/api/public/v2
	token string
	http  *http.Client
}

func newHTTPClient(base, token string, timeout time.Duration, hc *http.Client) *httpClient {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if hc == nil {
		hc = &http.Client{Timeout: timeout}
	}
	return &httpClient{base: base, token: token, http: hc}
}

// pedido é uma requisição: método, caminho (relativo a /api/public/v2), query, corpo e headers.
type pedido struct {
	method      string
	path        string
	query       url.Values
	body        io.Reader
	contentType string
	headers     map[string]string
}

// envelopeDados é o formato de toda resposta de sucesso com corpo.
type envelopeDados struct {
	Data       json.RawMessage `json:"data"`
	Pagination *Pagination     `json:"pagination"`
}

// do executa o pedido e decodifica data em out (quando out não é nil). Devolve a paginação,
// quando a resposta tem.
func (c *httpClient) do(ctx context.Context, p pedido, out any) (*Pagination, error) {
	u := c.base + p.path
	if len(p.query) > 0 {
		u += "?" + p.query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, p.method, u, p.body)
	if err != nil {
		return nil, fmt.Errorf("interage: erro ao montar a requisição: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", "Bearer "+c.token)
	if p.contentType != "" {
		req.Header.Set("Content-Type", p.contentType)
	}
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("interage: erro na requisição HTTP: %w", err)
	}
	defer resp.Body.Close()
	corpo, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("interage: erro ao ler a resposta: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, erroDaResposta(resp, corpo)
	}
	if out == nil || len(bytes.TrimSpace(corpo)) == 0 {
		return nil, nil
	}
	var env envelopeDados
	if err := json.Unmarshal(corpo, &env); err != nil {
		return nil, fmt.Errorf("interage: resposta fora do formato esperado: %w", err)
	}
	if len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return nil, fmt.Errorf("interage: erro ao decodificar a resposta: %w", err)
		}
	}
	return env.Pagination, nil
}

func (c *httpClient) json(ctx context.Context, method, path string, query url.Values, in, out any, headers map[string]string) error {
	p := pedido{method: method, path: path, query: query, headers: headers}
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("interage: erro ao serializar o corpo: %w", err)
		}
		p.body, p.contentType = bytes.NewReader(b), "application/json"
	}
	_, err := c.do(ctx, p, out)
	return err
}

func (c *httpClient) get(ctx context.Context, path string, query url.Values, out any) error {
	return c.json(ctx, http.MethodGet, path, query, nil, out, nil)
}

// lista faz um GET de lista e devolve a página.
func lista[T any](ctx context.Context, c *httpClient, path string, query url.Values) (*Page[T], error) {
	var itens []T
	pag, err := c.do(ctx, pedido{method: http.MethodGet, path: path, query: query}, &itens)
	if err != nil {
		return nil, err
	}
	pg := &Page[T]{Items: itens}
	if pag != nil {
		pg.NextCursor, pg.Total, pg.PageSize = pag.NextCursor, pag.Total, pag.PageSize
	}
	if pg.Items == nil {
		pg.Items = []T{}
	}
	return pg, nil
}

// seg escapa um segmento do caminho (número do ramal, protocolo, id).
func seg(v string) string { return url.PathEscape(strings.TrimSpace(v)) }
