package interage

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	// ErrUnauthorized — 401: token ausente, inválido ou sem a permissão exigida pela rota.
	ErrUnauthorized = errors.New("interage: não autorizado — token inválido ou sem permissão")
	// ErrForbidden — 403: rota não liberada para o token (verifique as permissões em api_routes).
	ErrForbidden = errors.New("interage: acesso negado — rota não liberada para o token")
	// ErrNotFound — 404: recurso não encontrado.
	ErrNotFound = errors.New("interage: recurso não encontrado")
	// ErrConflict — 409: conflito de estado (ex.: ramal já em chamada).
	ErrConflict = errors.New("interage: conflito de estado")
	// ErrUnprocessable — 422: ação não permitida no estado atual (ex.: transição de status inválida).
	ErrUnprocessable = errors.New("interage: ação não permitida no estado atual")
	// ErrInternalServer — 5xx: erro interno da API.
	ErrInternalServer = errors.New("interage: erro interno do servidor")
)

// APIError carrega o payload completo de erro retornado pela API.
// Faz Unwrap() para o erro sentinela correspondente ao status HTTP,
// então errors.Is continua funcionando normalmente.
type APIError struct {
	// StatusCode é o status HTTP da resposta.
	StatusCode int `json:"-"`
	// Code é o código retornado no corpo da resposta.
	Code int `json:"code"`
	// Status é o rótulo do erro (ex.: "BAD_REQUEST", "UNAUTHORIZED").
	Status string `json:"status"`
	// Message é a mensagem legível retornada pela API (pt-BR).
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("interage: API retornou %d (%s): %s", e.StatusCode, e.Status, e.Message)
}

// Unwrap mapeia o status HTTP para o erro sentinela correspondente.
func (e *APIError) Unwrap() error {
	switch {
	case e.StatusCode == http.StatusBadRequest:
		return ErrBadRequest
	case e.StatusCode == http.StatusUnauthorized:
		return ErrUnauthorized
	case e.StatusCode == http.StatusForbidden:
		return ErrForbidden
	case e.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case e.StatusCode == http.StatusConflict:
		return ErrConflict
	case e.StatusCode == http.StatusUnprocessableEntity:
		return ErrUnprocessable
	case e.StatusCode >= 500:
		return ErrInternalServer
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

// parseAPIError monta o APIError a partir do corpo da resposta.
func parseAPIError(statusCode int, body []byte) error {
	apiErr := &APIError{StatusCode: statusCode}
	_ = json.Unmarshal(body, apiErr)
	if apiErr.Status == "" {
		apiErr.Status = http.StatusText(statusCode)
	}
	if apiErr.Message == "" {
		apiErr.Message = string(body)
	}
	return apiErr
}
