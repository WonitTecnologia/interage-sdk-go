package interage

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Códigos de erro da API v2 (campo code do Problem). São estáveis: decida pelo código, não pelo
// texto de detail.
const (
	CodeTokenMissing           ErrorCode = "TOKEN_MISSING"
	CodeTokenConflict          ErrorCode = "TOKEN_CONFLICT"
	CodeTokenInQueryNotAllowed ErrorCode = "TOKEN_IN_QUERY_NOT_ALLOWED"
	CodeTokenInvalid           ErrorCode = "TOKEN_INVALID"
	CodeTokenExpired           ErrorCode = "TOKEN_EXPIRED"
	CodeTokenRevoked           ErrorCode = "TOKEN_REVOKED"
	CodeTokenVersionMismatch   ErrorCode = "TOKEN_VERSION_MISMATCH"
	CodeRouteNotAllowed        ErrorCode = "ROUTE_NOT_ALLOWED"
	CodeActionNotAllowed       ErrorCode = "ACTION_NOT_ALLOWED"
	CodeTenantInactive         ErrorCode = "TENANT_INACTIVE"
	CodeModuleNotLicensed      ErrorCode = "MODULE_NOT_LICENSED"
	CodeValidationError        ErrorCode = "VALIDATION_ERROR"
	CodeNotFound               ErrorCode = "NOT_FOUND"
	CodeMethodNotAllowed       ErrorCode = "METHOD_NOT_ALLOWED"
	CodeRateLimitExceeded      ErrorCode = "RATE_LIMIT_EXCEEDED"
	CodeIPRateLimitExceeded    ErrorCode = "IP_RATE_LIMIT_EXCEEDED"
	CodeTooManyInvalidTokens   ErrorCode = "TOO_MANY_INVALID_TOKENS"
	CodeRecordingNotFound      ErrorCode = "RECORDING_NOT_FOUND"
	CodeRecordingArchived      ErrorCode = "RECORDING_ARCHIVED"
	CodeFileNotFound           ErrorCode = "FILE_NOT_FOUND"
	CodeFileArchived           ErrorCode = "FILE_ARCHIVED"
	CodeConversationFinished   ErrorCode = "CONVERSATION_FINISHED"
	CodeConversationInCoex     ErrorCode = "CONVERSATION_IN_COEX"
	CodeSendInProgress         ErrorCode = "SEND_IN_PROGRESS"
	CodeProviderRejected       ErrorCode = "PROVIDER_REJECTED"
	// CodeProviderUnavailable: falha de comunicação com o provedor do canal (503). O envio não
	// foi confirmado e pode ter chegado ao contato: consulte a mensagem (extra message_id),
	// quando houver, antes de repetir.
	CodeProviderUnavailable       ErrorCode = "PROVIDER_UNAVAILABLE"
	CodeConversationAlreadyActive ErrorCode = "CONVERSATION_ALREADY_ACTIVE"
	CodeSessionWindowExpired      ErrorCode = "SESSION_WINDOW_EXPIRED"
	CodeChannelNotSupported       ErrorCode = "CHANNEL_NOT_SUPPORTED"
	CodeInstanceInactive          ErrorCode = "INSTANCE_INACTIVE"
	CodeIdempotencyInProgress     ErrorCode = "IDEMPOTENCY_IN_PROGRESS"
	CodeIdempotencyKeyReused      ErrorCode = "IDEMPOTENCY_KEY_REUSED"
	CodeInvalidCampaignStatus     ErrorCode = "INVALID_CAMPAIGN_STATUS"
	CodeCampaignStatusChanged     ErrorCode = "CAMPAIGN_STATUS_CHANGED"
	CodePayloadTooLarge           ErrorCode = "PAYLOAD_TOO_LARGE"
	CodeLicenseLimitReached       ErrorCode = "LICENSE_LIMIT_REACHED"
	CodeExtensionAlreadyExists    ErrorCode = "EXTENSION_ALREADY_EXISTS"
	CodeExtensionInUse            ErrorCode = "EXTENSION_IN_USE"
	CodeContactAlreadyExists      ErrorCode = "CONTACT_ALREADY_EXISTS"
	CodeExtensionBusy             ErrorCode = "EXTENSION_BUSY"
	CodeExtensionOffline          ErrorCode = "EXTENSION_OFFLINE"
	CodeDNDNotAllowed             ErrorCode = "DND_NOT_ALLOWED"
	CodeForwardingNotAllowed      ErrorCode = "FORWARDING_NOT_ALLOWED"
	CodeInternalError             ErrorCode = "INTERNAL_ERROR"
	CodeServiceUnavailable        ErrorCode = "SERVICE_UNAVAILABLE"
)

// Error é o erro devolvido pela API (RFC 9457, application/problem+json).
//
//	var e *interage.Error
//	if errors.As(err, &e) && e.Code == interage.CodeValidationError {
//		for _, f := range e.Errors { fmt.Println(f.Field, f.Message) }
//	}
type Error struct {
	// Status é o status HTTP.
	Status int
	// Code é o código estável do erro (CodeNotFound, CodeValidationError...).
	Code ErrorCode
	// Title é o título do código, em pt-BR.
	Title string
	// Detail explica o caso, em pt-BR.
	Detail string
	// Errors lista todos os campos recusados (VALIDATION_ERROR).
	Errors []FieldError
	// Extras são os membros a mais do código (ex.: protocol no CONVERSATION_ALREADY_ACTIVE).
	Extras map[string]json.RawMessage
	// RetryAfter é quanto esperar antes de tentar de novo (header Retry-After), em 429 e 503.
	RetryAfter time.Duration
	// RateLimit é o limite informado nos headers X-RateLimit-*, quando vieram.
	RateLimit *RateLimit
}

// RateLimit são os headers de limite da resposta.
type RateLimit struct {
	Limit     int
	Remaining int
	Reset     time.Duration // até o limite voltar
	Scope     string        // tenant (cota do cliente) ou ip
}

func (e *Error) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("interage: %d %s: %s", e.Status, e.Code, e.Detail)
	}
	return fmt.Sprintf("interage: %d %s", e.Status, e.Code)
}

// IsCode informa se err é um *Error com o código.
func IsCode(err error, code ErrorCode) bool {
	var e *Error
	return errors.As(err, &e) && e.Code == code
}

// Extra decodifica um membro extra do erro (ex.: e.Extra("protocol", &protocolo)).
func (e *Error) Extra(nome string, destino any) bool {
	b, ok := e.Extras[nome]
	return ok && json.Unmarshal(b, destino) == nil
}

var membrosDoProblema = map[string]bool{"type": true, "title": true, "status": true, "detail": true, "code": true, "errors": true, "instance": true}

func erroDaResposta(resp *http.Response, corpo []byte) error {
	e := &Error{Status: resp.StatusCode, RetryAfter: lerRetryAfter(resp.Header.Get("Retry-After")), RateLimit: lerRateLimit(resp.Header)}
	var p Problem
	if json.Unmarshal(corpo, &p) == nil && p.Code != "" {
		e.Code, e.Title, e.Detail, e.Errors = p.Code, p.Title, p.Detail, p.Errors
		var todos map[string]json.RawMessage
		if json.Unmarshal(corpo, &todos) == nil {
			for k, v := range todos {
				if !membrosDoProblema[k] {
					if e.Extras == nil {
						e.Extras = map[string]json.RawMessage{}
					}
					e.Extras[k] = v
				}
			}
		}
		return e
	}
	// Corpo fora do RFC 9457 (proxy, página de erro): status e o começo do corpo.
	e.Title = http.StatusText(resp.StatusCode)
	e.Detail = strings.TrimSpace(string(corpo))
	if len(e.Detail) > 300 {
		e.Detail = e.Detail[:300]
	}
	return e
}

func lerRetryAfter(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if s, err := strconv.Atoi(v); err == nil && s >= 0 {
		return time.Duration(s) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

func lerRateLimit(h http.Header) *RateLimit {
	limite, err := strconv.Atoi(h.Get("X-RateLimit-Limit"))
	if err != nil {
		return nil
	}
	r := &RateLimit{Limit: limite, Scope: h.Get("X-RateLimit-Scope")}
	r.Remaining, _ = strconv.Atoi(h.Get("X-RateLimit-Remaining"))
	if s, err := strconv.Atoi(h.Get("X-RateLimit-Reset")); err == nil {
		r.Reset = time.Duration(s) * time.Second
	}
	return r
}
