package interage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	userAgent      = "interage-sdk-go"
)

// httpClient é o transporte interno compartilhado por todos os domínios.
// Centraliza autenticação, montagem de requisições, unwrap do envelope
// {code,status,message,data} e tratamento de erros HTTP.
type httpClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func newHTTPClient(baseURL, token string, timeout time.Duration, hc *http.Client) *httpClient {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if hc == nil {
		hc = &http.Client{Timeout: timeout}
	}
	return &httpClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    hc,
	}
}

// envelope é o formato padrão de resposta da API: response.HttpResponse.
type envelope struct {
	Code    int             `json:"code"`
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// do executa a requisição, trata erros HTTP e desembrulha o campo `data`
// do envelope no ponteiro `out` (quando ambos existem).
func (c *httpClient) do(ctx context.Context, method, path string, query url.Values, body io.Reader, contentType string, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return fmt.Errorf("interage: erro ao montar requisição: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", "Token "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("interage: erro na requisição HTTP: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("interage: erro ao ler resposta: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, respBody)
	}

	if out == nil {
		return nil
	}

	var env envelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return fmt.Errorf("interage: erro ao decodificar envelope da resposta: %w", err)
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("interage: erro ao decodificar dados da resposta: %w", err)
	}
	return nil
}

// get executa um GET com query params opcionais.
func (c *httpClient) get(ctx context.Context, path string, query url.Values, out any) error {
	return c.do(ctx, http.MethodGet, path, query, nil, "", out)
}

// post executa um POST com corpo JSON opcional.
func (c *httpClient) post(ctx context.Context, path string, query url.Values, in, out any) error {
	body, contentType, err := jsonBody(in)
	if err != nil {
		return err
	}
	return c.do(ctx, http.MethodPost, path, query, body, contentType, out)
}

// patch executa um PATCH com corpo JSON opcional.
func (c *httpClient) patch(ctx context.Context, path string, in, out any) error {
	body, contentType, err := jsonBody(in)
	if err != nil {
		return err
	}
	return c.do(ctx, http.MethodPatch, path, nil, body, contentType, out)
}

// delete executa um DELETE.
func (c *httpClient) delete(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil, "", out)
}

// postMultipart executa um POST multipart/form-data com campos de texto e um arquivo.
func (c *httpClient) postMultipart(ctx context.Context, path string, fields map[string]string, fileField, fileName string, fileContent []byte, out any) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return fmt.Errorf("interage: erro ao montar formulário: %w", err)
		}
	}
	fw, err := w.CreateFormFile(fileField, fileName)
	if err != nil {
		return fmt.Errorf("interage: erro ao anexar arquivo: %w", err)
	}
	if _, err := fw.Write(fileContent); err != nil {
		return fmt.Errorf("interage: erro ao escrever arquivo: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("interage: erro ao finalizar formulário: %w", err)
	}
	return c.do(ctx, http.MethodPost, path, nil, &buf, w.FormDataContentType(), out)
}

// jsonBody serializa o corpo JSON quando presente.
func jsonBody(in any) (io.Reader, string, error) {
	if in == nil {
		return nil, "", nil
	}
	b, err := json.Marshal(in)
	if err != nil {
		return nil, "", fmt.Errorf("interage: erro ao serializar corpo da requisição: %w", err)
	}
	return bytes.NewReader(b), "application/json; charset=utf-8", nil
}

// pageQuery monta query de paginação, incluindo apenas valores > 0.
func pageQuery(page, pageSize int) url.Values {
	q := url.Values{}
	if page > 0 {
		q.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		q.Set("page_size", fmt.Sprintf("%d", pageSize))
	}
	return q
}
