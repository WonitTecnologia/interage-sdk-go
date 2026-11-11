package interage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Erros sentinela do SDK. Use errors.Is para verificar a categoria do erro:
//
//	if errors.Is(err, interage.ErrNotFound) { ... }
var (
	// ErrInvalidToken indica que o token informado no NewClient está vazio.
	ErrInvalidToken = errors.New("interage: token não pode ser vazio")
	// ErrInvalidBaseURL indica que a baseURL informada no NewClient está vazia.
	ErrInvalidBaseURL = errors.New("interage: baseURL não pode ser vazia")
	// ErrBadRequest — 400: parâmetros inválidos.
	ErrBadRequest = errors.New("interage: requisição inválida")
	// ErrUnauthorized — 401: token ausente, inválido, expirado ou revogado.
	// A credencial precisa ser corrigida (gerar/renovar o token no painel).
	ErrUnauthorized = errors.New("interage: não autorizado — token ausente, inválido, expirado ou revogado")
	// ErrForbidden — 403: a rota não foi liberada para o token, falta a ação exigida
	// (leitura, listagem, criação, alteração ou remoção), o cliente está desativado ou
	// o módulo da rota não está licenciado. Leia APIError.Message para saber qual.
	// Trocar de token não resolve.
	ErrForbidden = errors.New("interage: acesso negado — rota ou ação não liberada para o token, cliente desativado ou módulo não licenciado")
	// ErrNotFound — 404: recurso não encontrado.
	ErrNotFound = errors.New("interage: recurso não encontrado")
	// ErrConflict — 409: conflito de estado (ex.: ramal já em chamada, requisição com a
	// mesma Idempotency-Key ainda em andamento).
	ErrConflict = errors.New("interage: conflito de estado")
	// ErrGone — 410: o recurso existiu e não volta mais (ex.: gravação arquivada pela
	// política de retenção). Não adianta repetir.
	ErrGone = errors.New("interage: recurso não está mais disponível")
	// ErrPayloadTooLarge — 413: corpo maior que o aceito (ex.: CSV de campanha acima de 20 MB).
	ErrPayloadTooLarge = errors.New("interage: requisição maior que o permitido")
	// ErrUnprocessable — 422: ação não permitida no estado atual (ex.: transição de status
	// inválida, janela de 24h expirada, Idempotency-Key reutilizada com outro conteúdo).
	ErrUnprocessable = errors.New("interage: ação não permitida no estado atual")
	// ErrTooManyRequests — 429: limite de requisições excedido (cota do cliente, somando todos os tokens dele, ou teto por IP).
	// Aguarde APIError.RetryAfter antes de tentar de novo.
	ErrTooManyRequests = errors.New("interage: limite de requisições excedido")
	// ErrInternalServer — 5xx: erro da API. Os erros de 502 (ErrSendFailed) e 503
	// (ErrServiceUnavailable) também casam com este.
	ErrInternalServer = errors.New("interage: erro interno do servidor")
	// ErrSendFailed — 502 SEND_FAILED: o provedor (WhatsApp) recusou o envio. Nada saiu;
	// APIError.InternalMessageID, quando presente, permite consultar o motivo em
	// Messages.GetMessageStatus.
	ErrSendFailed = errors.New("interage: o provedor recusou o envio")
	// ErrServiceUnavailable — 503: indisponibilidade temporária. Nas rotas de ação, a
	// ação não foi executada. Aguarde APIError.RetryAfter e repita.
	ErrServiceUnavailable = errors.New("interage: serviço temporariamente indisponível")
	// ErrUnexpectedResponse — resposta fora do contrato (3xx, 2xx vazio ou que não é
	// JSON). Ver UnexpectedResponseError.
	ErrUnexpectedResponse = errors.New("interage: resposta inesperada da API")
)

// maxRawBody é o quanto do corpo de uma resposta de erro fora do padrão fica em APIError.RawBody.
const maxRawBody = 2048

// APIError carrega o payload completo de erro retornado pela API.
// Faz Unwrap() para o(s) erro(s) sentinela correspondente(s) ao status HTTP,
// então errors.Is continua funcionando normalmente.
type APIError struct {
	// StatusCode é o status HTTP da resposta.
	StatusCode int `json:"-"`
	// Code é o código retornado no corpo da resposta.
	Code int `json:"code"`
	// Status é o rótulo do erro (ex.: "BAD_REQUEST", "SEND_FAILED", "IDEMPOTENCY_KEY_REUSED").
	Status string `json:"status"`
	// Message é a mensagem legível retornada pela API (pt-BR).
	Message string `json:"message"`
	// Data é o campo data do erro, quando a API o envia (ex.: no 502 SEND_FAILED).
	Data json.RawMessage `json:"data,omitempty"`
	// InternalMessageID é o id da mensagem gravada com falha no 502 SEND_FAILED,
	// quando houve registro. Use em Messages.GetMessageStatus para ver o motivo.
	InternalMessageID *int64 `json:"-"`
	// RetryAfter é quanto esperar antes de tentar de novo, lido do header
	// Retry-After. Preenchido em 429 e 503; zero quando a API não informou.
	RetryAfter time.Duration `json:"-"`
	// RateLimit é o consumo da cota (headers X-RateLimit-*); nil quando a API não os enviou.
	RateLimit *RateLimit `json:"-"`
	// RawBody é o corpo da resposta (até 2 KB) quando ele não é o envelope JSON da
	// API — ex.: página HTML de um proxy.
	RawBody string `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("interage: API retornou %d (%s): %s", e.StatusCode, e.Status, e.Message)
}

// Unwrap mapeia o status HTTP para o(s) erro(s) sentinela correspondente(s).
func (e *APIError) Unwrap() []error {
	switch {
	case e.StatusCode == http.StatusBadRequest:
		return []error{ErrBadRequest}
	case e.StatusCode == http.StatusUnauthorized:
		return []error{ErrUnauthorized}
	case e.StatusCode == http.StatusForbidden:
		return []error{ErrForbidden}
	case e.StatusCode == http.StatusNotFound:
		return []error{ErrNotFound}
	case e.StatusCode == http.StatusConflict:
		return []error{ErrConflict}
	case e.StatusCode == http.StatusGone:
		return []error{ErrGone}
	case e.StatusCode == http.StatusRequestEntityTooLarge:
		return []error{ErrPayloadTooLarge}
	case e.StatusCode == http.StatusUnprocessableEntity:
		return []error{ErrUnprocessable}
	case e.StatusCode == http.StatusTooManyRequests:
		return []error{ErrTooManyRequests}
	case e.StatusCode == http.StatusBadGateway && e.Status == "SEND_FAILED":
		return []error{ErrSendFailed, ErrInternalServer}
	case e.StatusCode == http.StatusServiceUnavailable:
		return []error{ErrServiceUnavailable, ErrInternalServer}
	case e.StatusCode >= 500:
		return []error{ErrInternalServer}
	default:
		return nil
	}
}

// AsAPIError extrai o *APIError de uma cadeia de erros, quando presente.
// Útil para inspecionar o payload completo (Status, Message):
//
//	if apiErr, ok := interage.AsAPIError(err); ok {
//	    log.Println(apiErr.StatusCode, apiErr.Message)
//	}
func AsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// parseAPIError monta o APIError a partir do corpo e dos headers da resposta.
func parseAPIError(statusCode int, header http.Header, body []byte, rateLimit *RateLimit) error {
	apiErr := &APIError{
		StatusCode: statusCode,
		RetryAfter: parseRetryAfter(header.Get("Retry-After")),
		RateLimit:  rateLimit,
	}
	if err := json.Unmarshal(body, apiErr); err != nil || (apiErr.Status == "" && apiErr.Message == "") {
		// Não é o envelope da API (ex.: HTML de um proxy): mensagem curta, corpo à parte.
		raw := bytes.TrimSpace(body)
		if len(raw) > maxRawBody {
			raw = raw[:maxRawBody]
		}
		apiErr.RawBody = string(raw)
	}
	if apiErr.Status == "" {
		apiErr.Status = http.StatusText(statusCode)
	}
	if apiErr.Message == "" {
		apiErr.Message = "resposta HTTP " + strconv.Itoa(statusCode) + " sem o envelope de erro da API"
	}
	if len(apiErr.Data) > 0 && string(apiErr.Data) != "null" {
		var d struct {
			InternalMessageID *int64 `json:"internal_message_id"`
		}
		if json.Unmarshal(apiErr.Data, &d) == nil {
			apiErr.InternalMessageID = d.InternalMessageID
		}
	}
	return apiErr
}

// parseRetryAfter lê o header Retry-After em segundos (formato usado pela API).
// Valor ausente ou inválido resulta em zero.
func parseRetryAfter(v string) time.Duration {
	secs, err := strconv.Atoi(v)
	if err != nil || secs <= 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}
