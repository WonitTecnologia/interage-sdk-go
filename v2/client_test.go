package interage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func servidor(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	cli, err := NewClient(srv.URL, "sk_teste", nil)
	if err != nil {
		t.Fatal(err)
	}
	return cli
}

func TestAutenticacaoECaminho(t *testing.T) {
	cli := servidor(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk_teste" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/api/public/v2/extensions/10%2F1" && r.URL.RawPath != "/api/public/v2/extensions/10%2F1" {
			t.Errorf("caminho = %q (%q)", r.URL.Path, r.URL.RawPath)
		}
		w.Write([]byte(`{"data":{"number":"10/1","name":"Recepção","forwarding":{"always":null}}}`))
	})
	r, err := cli.Extensions.Get(context.Background(), "10/1")
	if err != nil || r.Number != "10/1" || r.Name != "Recepção" {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestNewClientValida(t *testing.T) {
	if _, err := NewClient("", "sk", nil); !errors.Is(err, ErrInvalidBaseURL) {
		t.Fatal(err)
	}
	if _, err := NewClient("x.wonit.cloud", " ", nil); !errors.Is(err, ErrInvalidToken) {
		t.Fatal(err)
	}
	cli, _ := NewClient("x.wonit.cloud/", "sk", nil)
	if cli.http.base != "https://x.wonit.cloud/api/public/v2" {
		t.Fatal(cli.http.base)
	}
}

// Filtros na query (data em RFC 3339 com fuso, zero fica de fora) e All seguindo o cursor.
func TestListaFiltrosEPaginacao(t *testing.T) {
	fuso := time.FixedZone("-03", -3*3600)
	de := time.Date(2026, 11, 1, 0, 0, 0, 0, fuso)
	var pedidos []string
	cli := servidor(t, func(w http.ResponseWriter, r *http.Request) {
		pedidos = append(pedidos, r.URL.RawQuery)
		q := r.URL.Query()
		if q.Get("from") != "2026-11-01T00:00:00-03:00" || q.Get("direction") != "inbound" || q.Has("number") {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		if q.Get("cursor") == "" {
			w.Write([]byte(`{"data":[{"id":"A"},{"id":"B"}],"pagination":{"next_cursor":"c2","page_size":2,"total":3}}`))
			return
		}
		w.Write([]byte(`{"data":[{"id":"C"}],"pagination":{"next_cursor":null,"page_size":2,"total":3}}`))
	})
	var ids []string
	for ligacao, err := range cli.Calls.All(context.Background(), ListCallsParams{From: de, To: de.AddDate(0, 0, 1), Direction: "inbound"}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, ligacao.ID)
	}
	if strings.Join(ids, ",") != "A,B,C" || len(pedidos) != 2 || !strings.Contains(pedidos[1], "cursor=c2") {
		t.Fatalf("ids %v, pedidos %v", ids, pedidos)
	}
}

func TestErroRFC9457(t *testing.T) {
	cli := servidor(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.Header().Set("Retry-After", "7")
		w.Header().Set("X-RateLimit-Limit", "120")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Scope", "tenant")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"title":"Conversa já ativa","status":409,"detail":"Já existe conversa.","code":"CONVERSATION_ALREADY_ACTIVE","protocol":"WT123","errors":[{"field":"to","message":"x"}]}`))
	})
	_, err := cli.Conversations.Start(context.Background(), ConversationStartRequest{})
	var e *Error
	if !errors.As(err, &e) || e.Status != 409 || e.Code != CodeConversationAlreadyActive || len(e.Errors) != 1 {
		t.Fatalf("%#v", err)
	}
	var protocolo string
	if !e.Extra("protocol", &protocolo) || protocolo != "WT123" {
		t.Fatalf("extra protocol = %q", protocolo)
	}
	if e.RetryAfter != 7*time.Second || e.RateLimit == nil || e.RateLimit.Limit != 120 || e.RateLimit.Scope != "tenant" {
		t.Fatalf("limite: %v %+v", e.RetryAfter, e.RateLimit)
	}
	if !IsCode(err, CodeConversationAlreadyActive) || IsCode(err, CodeNotFound) {
		t.Fatal("IsCode")
	}
}

func TestErroForaDoFormato(t *testing.T) {
	cli := servidor(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("<html>bad gateway</html>"))
	})
	err := cli.Contacts.Delete(context.Background(), "x")
	var e *Error
	if !errors.As(err, &e) || e.Status != 502 || e.Code != "" || !strings.Contains(e.Detail, "bad gateway") {
		t.Fatalf("%#v", err)
	}
}

// Opt: ausente não vai, Null vai null, Set vai o valor — inclusive lista vazia.
func TestOptTresEstados(t *testing.T) {
	var corpo map[string]json.RawMessage
	cli := servidor(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("método %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&corpo)
		w.Write([]byte(`{"data":{"id":"c1"}}`))
	})
	_, err := cli.Contacts.Update(context.Background(), "c1", ContactUpdate{
		Email:   Set("joao@empresa.com"),
		Company: Null[string](),
		Labels:  Set([]int64{}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(corpo["email"]) != `"joao@empresa.com"` || string(corpo["company"]) != "null" || string(corpo["labels"]) != "[]" {
		t.Fatalf("corpo = %v", corpo)
	}
	if _, ok := corpo["name"]; ok {
		t.Fatal("campo ausente foi no corpo")
	}
	var o Opt[string]
	json.Unmarshal([]byte("null"), &o)
	if !o.IsNull() || o.IsZero() {
		t.Fatal("Unmarshal null")
	}
}

func TestCampanhaCSVComChave(t *testing.T) {
	cli := servidor(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Idempotency-Key") != "chave-1" {
			t.Errorf("Idempotency-Key = %q", r.Header.Get("Idempotency-Key"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		var c CampaignCreateRequest
		json.Unmarshal([]byte(r.FormValue("campaign")), &c)
		f, _, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		csv, _ := io.ReadAll(f)
		if c.Name != "Black Friday" || string(csv) != "telefone\n5511999998888\n" {
			t.Errorf("campanha %+v csv %q", c, csv)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"id":"cp1","name":"Black Friday"}}`))
	})
	r, err := cli.Campaigns.CreateFromCSV(context.Background(), CampaignCreateRequest{Name: "Black Friday", CollisionPolicy: "ignore"},
		"contatos.csv", bytes.NewBufferString("telefone\n5511999998888\n"), &CreateCampaignOptions{IdempotencyKey: "chave-1"})
	if err != nil || r.ID != "cp1" {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestDeleteSemCorpo(t *testing.T) {
	cli := servidor(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Query().Get("force") != "true" {
			t.Errorf("%s %s", r.Method, r.URL)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := cli.Extensions.Delete(context.Background(), "1001", true); err != nil {
		t.Fatal(err)
	}
}

// Vetor calculado à parte (HMAC-SHA256 de "<t>.<corpo>"), no formato que a API assina.
func TestAssinaturaDoWebhook(t *testing.T) {
	corpo := []byte(`{"events":[{"id":"e1","type":"ping","created_at":"2026-11-18T10:00:00-03:00"}]}`)
	const vetor = "fd5f13fc343e421ec1a2b02f771e6793eb2ce027d1cd88d950f67da3fb5738a0"
	if err := VerifyWebhookSignature("whsec_teste", "t=1795000000,v1="+vetor, corpo, -1); err != nil {
		t.Fatalf("vetor recusado: %v", err)
	}
	// Na troca do segredo, dois v1: vale qualquer um.
	if err := VerifyWebhookSignature("whsec_teste", "t=1795000000,v1=00ff,v1="+vetor, corpo, -1); err != nil {
		t.Fatal(err)
	}
	if err := VerifyWebhookSignature("outro", "t=1795000000,v1="+vetor, corpo, -1); !errors.Is(err, ErrWebhookSignature) {
		t.Fatal(err)
	}
	if err := VerifyWebhookSignature("whsec_teste", "t=1795000000,v1="+vetor, corpo, 0); !errors.Is(err, ErrWebhookExpired) {
		t.Fatalf("antiga deveria vencer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(corpo))
	req.Header.Set(HeaderWebhookSignature, "t=1795000000,v1="+vetor)
	lote, err := ParseWebhook(req, "whsec_teste", -1)
	if err != nil || len(lote.Events) != 1 || lote.Events[0].Type != EventPing {
		t.Fatalf("%+v %v", lote, err)
	}
}
