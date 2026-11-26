package interage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout        = 30 * time.Second
	defaultConnectTimeout = 10 * time.Second
	userAgent             = "interage-sdk-go/" + Version
)

// httpClient é o transporte interno compartilhado por todos os domínios.
// Centraliza autenticação, montagem de requisições, unwrap do envelope
// {code,status,message,data} e tratamento de erros HTTP.
type httpClient struct {
	baseURL    string
	token      string
	http       *http.Client
	onResponse func(ResponseInfo)
}

func newHTTPClient(baseURL, token string, timeout, connectTimeout time.Duration, hc *http.Client, onResponse func(ResponseInfo)) *httpClient {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if connectTimeout <= 0 {
		connectTimeout = defaultConnectTimeout
	}
	if hc == nil {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.DialContext = (&net.Dialer{Timeout: connectTimeout, KeepAlive: 30 * time.Second}).DialContext
		hc = &http.Client{
			Timeout:   timeout,
			Transport: transport,
			// A API nunca redireciona. Seguir um redirect (ex.: http → https no proxy)
			// trocaria o POST por GET; o 3xx volta como erro, com o destino na mensagem.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}
	}
	return &httpClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		http:       hc,
		onResponse: onResponse,
	}
}

// envelope é o formato padrão de resposta da API: response.HttpResponse.
type envelope struct {
	Code    int             `json:"code"`
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// RateLimit é o consumo da cota informado pela API nos headers X-RateLimit-*.
// A API não manda os headers em 401, 403, nos limites por IP nem quando a cota
// não pôde ser consultada — nesses casos o ponteiro vem nil.
type RateLimit struct {
	// Limit é o tamanho da cota da classe da rota (padrão ou ação).
	Limit int
	// Remaining é quantas requisições ainda cabem agora.
	Remaining int
}

// ResponseInfo descreve uma resposta recebida. É entregue a Options.OnResponse
// em toda resposta HTTP (sucesso ou erro).
type ResponseInfo struct {
	Method     string
	Path       string
	StatusCode int
	// RateLimit vem dos headers X-RateLimit-*; nil quando a API não os enviou.
	RateLimit *RateLimit
	// IdempotencyReplayed é true quando a API devolveu a resposta guardada de uma
	// requisição anterior com a mesma Idempotency-Key (header Idempotency-Replayed).
	IdempotencyReplayed bool
}

// requestOptions são headers por requisição.
type requestOptions struct {
	idempotencyKey string
}

// responseMeta é o que o domínio precisa saber da resposta além do data.
type responseMeta struct {
	statusCode int
	replayed   bool
}

func parseRateLimit(h http.Header) *RateLimit {
	limit, err1 := strconv.Atoi(h.Get("X-RateLimit-Limit"))
	remaining, err2 := strconv.Atoi(h.Get("X-RateLimit-Remaining"))
	if err1 != nil || err2 != nil {
		return nil
	}
	return &RateLimit{Limit: limit, Remaining: remaining}
}

// do executa a requisição, trata erros HTTP e desembrulha o campo `data`
// do envelope no ponteiro `out` (quando ambos existem).
func (c *httpClient) do(ctx context.Context, method, path string, query url.Values, body io.Reader, contentType string, opts requestOptions, out any) (responseMeta, error) {
	var meta responseMeta
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return meta, fmt.Errorf("interage: erro ao montar requisição: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", "Token "+c.token)
	if opts.idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", opts.idempotencyKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return meta, fmt.Errorf("interage: erro na requisição HTTP: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return meta, fmt.Errorf("interage: erro ao ler resposta: %w", err)
	}

	rateLimit := parseRateLimit(resp.Header)
	meta.statusCode = resp.StatusCode
	meta.replayed = strings.EqualFold(resp.Header.Get("Idempotency-Replayed"), "true")
	if c.onResponse != nil {
		c.onResponse(ResponseInfo{
			Method: method, Path: path, StatusCode: resp.StatusCode,
			RateLimit: rateLimit, IdempotencyReplayed: meta.replayed,
		})
	}

	if resp.StatusCode >= 400 {
		return meta, parseAPIError(resp.StatusCode, resp.Header, respBody, rateLimit)
	}
	if resp.StatusCode >= 300 {
		return meta, &UnexpectedResponseError{
			StatusCode: resp.StatusCode,
			Reason:     "redirecionamento para " + resp.Header.Get("Location") + " — confira a baseURL (use https://)",
		}
	}

	if out == nil {
		return meta, nil
	}

	// Daqui em diante a requisição foi aceita (2xx): numa ação (envio, ligação,
	// campanha) ela pode já ter sido executada. O erro avisa isso, para quem chama
	// não repetir sem Idempotency-Key.
	if len(bytes.TrimSpace(respBody)) == 0 {
		return meta, &UnexpectedResponseError{StatusCode: resp.StatusCode, Reason: "corpo vazio"}
	}
	var env envelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return meta, &UnexpectedResponseError{StatusCode: resp.StatusCode, Reason: "envelope não é JSON válido", Err: err}
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return meta, &UnexpectedResponseError{StatusCode: resp.StatusCode, Reason: "resposta sem o campo data"}
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return meta, &UnexpectedResponseError{StatusCode: resp.StatusCode, Reason: "campo data em formato inesperado", Err: err}
	}
	return meta, nil
}

// UnexpectedResponseError indica uma resposta fora do contrato da API: 3xx, ou 2xx
// com corpo vazio ou que não é o envelope JSON esperado. Com 2xx, a ação pode ter
// sido executada — repita só com a mesma Idempotency-Key (envios e campanhas) ou
// depois de conferir o estado (ex.: histórico de ligações).
type UnexpectedResponseError struct {
	StatusCode int
	Reason     string
	Err        error
}

func (e *UnexpectedResponseError) Error() string {
	msg := fmt.Sprintf("interage: resposta inesperada da API (HTTP %d): %s", e.StatusCode, e.Reason)
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}

func (e *UnexpectedResponseError) Unwrap() []error {
	if e.Err == nil {
		return []error{ErrUnexpectedResponse}
	}
	return []error{ErrUnexpectedResponse, e.Err}
}

// AsUnexpectedResponseError extrai o *UnexpectedResponseError de uma cadeia de erros.
func AsUnexpectedResponseError(err error) (*UnexpectedResponseError, bool) {
	var u *UnexpectedResponseError
	if errors.As(err, &u) {
		return u, true
	}
	return nil, false
}

// get executa um GET com query params opcionais.
func (c *httpClient) get(ctx context.Context, path string, query url.Values, out any) error {
	_, err := c.do(ctx, http.MethodGet, path, query, nil, "", requestOptions{}, out)
	return err
}

// post executa um POST com corpo JSON opcional.
func (c *httpClient) post(ctx context.Context, path string, query url.Values, in, out any) error {
	_, err := c.postWith(ctx, path, query, in, requestOptions{}, out)
	return err
}

// postWith é o post com headers por requisição (Idempotency-Key).
func (c *httpClient) postWith(ctx context.Context, path string, query url.Values, in any, opts requestOptions, out any) (responseMeta, error) {
	body, contentType, err := jsonBody(in)
	if err != nil {
		return responseMeta{}, err
	}
	return c.do(ctx, http.MethodPost, path, query, body, contentType, opts, out)
}

// patch executa um PATCH com corpo JSON opcional.
func (c *httpClient) patch(ctx context.Context, path string, in, out any) error {
	body, contentType, err := jsonBody(in)
	if err != nil {
		return err
	}
	_, err = c.do(ctx, http.MethodPatch, path, nil, body, contentType, requestOptions{}, out)
	return err
}

// delete executa um DELETE.
func (c *httpClient) delete(ctx context.Context, path string, out any) error {
	_, err := c.do(ctx, http.MethodDelete, path, nil, nil, "", requestOptions{}, out)
	return err
}

// postMultipart executa um POST multipart/form-data com campos de texto e um arquivo.
func (c *httpClient) postMultipart(ctx context.Context, path string, fields map[string]string, fileField, fileName string, fileContent []byte, opts requestOptions, out any) (responseMeta, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return responseMeta{}, fmt.Errorf("interage: erro ao montar formulário: %w", err)
		}
	}
	fw, err := w.CreateFormFile(fileField, fileName)
	if err != nil {
		return responseMeta{}, fmt.Errorf("interage: erro ao anexar arquivo: %w", err)
	}
	if _, err := fw.Write(fileContent); err != nil {
		return responseMeta{}, fmt.Errorf("interage: erro ao escrever arquivo: %w", err)
	}
	if err := w.Close(); err != nil {
		return responseMeta{}, fmt.Errorf("interage: erro ao finalizar formulário: %w", err)
	}
	return c.do(ctx, http.MethodPost, path, nil, &buf, w.FormDataContentType(), opts, out)
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
