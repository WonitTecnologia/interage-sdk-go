package interage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEnvioComIdempotencyKeyEQueueID(t *testing.T) {
	cli, got := fakeAPI(t, 200, map[string]string{"Idempotency-Replayed": "true"}, ok(`{"message_id":"g1"}`))
	resp, err := cli.Messages.SendTemplate(context.Background(), SendTemplateRequest{
		To: "5511999998888", TemplateID: "t", InstanceID: "i",
		AssignedUserID: "u", QueueID: Ptr[int32](7), IdempotencyKey: "envio-1",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got.header.Get("Idempotency-Key") != "envio-1" {
		t.Errorf("Idempotency-Key = %q", got.header.Get("Idempotency-Key"))
	}
	var body map[string]any
	_ = json.Unmarshal(got.body, &body)
	if body["queue_id"] != float64(7) {
		t.Errorf("queue_id = %v, want 7", body["queue_id"])
	}
	if _, vazou := body["IdempotencyKey"]; vazou {
		t.Error("IdempotencyKey não pode ir no corpo")
	}
	if !resp.Replayed {
		t.Error("Replayed deveria ser true com Idempotency-Replayed: true")
	}
}

func TestEnvioSemIdempotencyKeyNaoMandaHeader(t *testing.T) {
	cli, got := fakeAPI(t, 200, nil, ok(`{"internal_message_id":1}`))
	resp, err := cli.Messages.SendMessage(context.Background(), SendMessageRequest{Protocol: "p", Type: "text", Text: "oi"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if _, enviado := got.header["Idempotency-Key"]; enviado {
		t.Error("Idempotency-Key enviado sem ter sido informado")
	}
	if resp.Replayed {
		t.Error("Replayed deveria ser false")
	}
}

func TestLocalizacaoZeroEEnviada(t *testing.T) {
	cli, got := fakeAPI(t, 200, nil, ok(`{"internal_message_id":1}`))
	_, err := cli.Messages.SendMessage(context.Background(), SendMessageRequest{
		Protocol: "p", Type: "location", Latitude: Ptr(0.0), Longitude: Ptr(0.0),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !strings.Contains(string(got.body), `"latitude":0`) || !strings.Contains(string(got.body), `"longitude":0`) {
		t.Fatalf("latitude/longitude 0 não enviadas: %s", got.body)
	}
}

func TestCampanhaRepetidaPorIdempotencyKey(t *testing.T) {
	cli, got := fakeAPI(t, 200, nil, ok(`{"campaign_id":"c1","status":"pending"}`))
	req := baseCampaignRequest()
	req.IdempotencyKey = "camp-1"
	resp, err := cli.Campaigns.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got.header.Get("Idempotency-Key") != "camp-1" {
		t.Errorf("Idempotency-Key = %q", got.header.Get("Idempotency-Key"))
	}
	if !resp.Replayed {
		t.Error("200 na criação deveria marcar Replayed")
	}
}

func TestEncerrarConversaComESemClassificacao(t *testing.T) {
	cli, got := fakeAPI(t, 200, nil, ok(`{"id":1,"status":"finished"}`))
	if _, err := cli.Omni.CloseConversationWith(context.Background(), "2026-1", CloseConversationRequest{Classification: "Resolvido"}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if string(got.body) != `{"classification":"Resolvido"}` {
		t.Errorf("corpo = %s", got.body)
	}
	if _, err := cli.Omni.CloseConversation(context.Background(), "2026-1"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(got.body) != 0 {
		t.Errorf("encerrar sem classificação não deveria mandar corpo: %s", got.body)
	}
}

func TestLoteLimpaCPF(t *testing.T) {
	b, err := json.Marshal(BatchContactItem{Name: "a", ClearCPF: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"cpf":""`) {
		t.Errorf("ClearCPF não enviou cpf vazio: %s", b)
	}
	b, _ = json.Marshal(BatchContactItem{Name: "a"})
	if strings.Contains(string(b), "cpf") {
		t.Errorf("cpf enviado sem ter sido informado: %s", b)
	}
	b, _ = json.Marshal(BatchContactItem{Name: "a", CPF: "123", ClearCPF: true})
	if !strings.Contains(string(b), `"cpf":"123"`) {
		t.Errorf("CPF preenchido deveria prevalecer: %s", b)
	}
}

func TestModelosNovos(t *testing.T) {
	cli, _ := fakeAPI(t, 200, nil, ok(`{"id":"c1","instance_id":"i1","auto_start":true,"scheduled_at":"x",
		"import":{"file_name":"a.csv","status":"failed","total_rows":3,"failed_rows":3,"error_message":"sem coluna"}}`))
	c, err := cli.Campaigns.Get(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if c.InstanceID != "i1" || !c.AutoStart || c.Import == nil || c.Import.ErrorMessage != "sem coluna" || c.ScheduledAt != "" {
		t.Errorf("campanha = %+v", c)
	}

	cli, _ = fakeAPI(t, 200, nil, ok(`{"id":"x","labels":[{"tenant_seq":3,"name":"VIP"}],"visibility":"public"}`))
	ct, err := cli.Contacts.GetContact(context.Background(), "x")
	if err != nil {
		t.Fatal(err)
	}
	if len(ct.Labels) != 1 || ct.Labels[0].TenantSeq != 3 || ct.Visibility != "" {
		t.Errorf("contato = %+v", ct)
	}

	cli, _ = fakeAPI(t, 200, nil, ok(`{"conversation":{"id":1,"status":"coex"},"items":[{"id":5,"edited":true,
		"removed_at":"2026-05-27T10:05:00Z","media_file_name":"a.pdf","reactions":[{"emoji":"👍","sent_by":"client"}]}]}`))
	h, err := cli.Omni.GetConversationHistory(context.Background(), "p", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	m := h.Items[0]
	if !m.Edited || m.RemovedAt == "" || m.MediaFileName != "a.pdf" || len(m.Reactions) != 1 {
		t.Errorf("mensagem = %+v", m)
	}
}

func TestErrosNovos(t *testing.T) {
	cases := []struct {
		status int
		corpo  string
		want   []error
	}{
		{410, `{"code":410,"status":"GONE","message":"arquivada"}`, []error{ErrGone}},
		{413, `{"code":413,"status":"REQUEST_ENTITY_TOO_LARGE","message":"grande"}`, []error{ErrPayloadTooLarge}},
		{503, `{"code":503,"status":"SERVICE_UNAVAILABLE","message":"indisponível"}`, []error{ErrServiceUnavailable, ErrInternalServer}},
		{502, `{"code":502,"status":"SEND_FAILED","message":"recusado","data":{"internal_message_id":42}}`, []error{ErrSendFailed, ErrInternalServer}},
	}
	for _, tc := range cases {
		t.Run(strconv.Itoa(tc.status), func(t *testing.T) {
			cli, _ := fakeAPI(t, tc.status, map[string]string{"Retry-After": "5"}, tc.corpo)
			_, err := cli.Campaigns.Get(context.Background(), "abc")
			for _, w := range tc.want {
				if !errors.Is(err, w) {
					t.Errorf("err = %v, want %v", err, w)
				}
			}
			apiErr, _ := AsAPIError(err)
			if apiErr.RetryAfter != 5*time.Second {
				t.Errorf("RetryAfter = %v", apiErr.RetryAfter)
			}
			if tc.status == 502 && (apiErr.InternalMessageID == nil || *apiErr.InternalMessageID != 42) {
				t.Errorf("InternalMessageID = %v, want 42", apiErr.InternalMessageID)
			}
		})
	}
}

func TestErroComHTMLNaoViraMensagem(t *testing.T) {
	cli, _ := fakeAPI(t, 504, nil, `<html><body>Gateway Timeout</body></html>`)
	_, err := cli.Campaigns.Get(context.Background(), "abc")
	apiErr, okErr := AsAPIError(err)
	if !okErr || strings.Contains(apiErr.Message, "<html>") || !strings.Contains(apiErr.RawBody, "Gateway Timeout") {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}

func TestSucessoForaDoContrato(t *testing.T) {
	for _, corpo := range []string{"", "não é json", `{"code":200,"status":"OK","message":"ok"}`} {
		cli, _ := fakeAPI(t, 200, nil, corpo)
		_, err := cli.Telephony.OriginateCall(context.Background(), OriginateCallRequest{FromExtension: "1", ToNumber: "2"})
		if !errors.Is(err, ErrUnexpectedResponse) {
			t.Errorf("corpo %q: err = %v, want ErrUnexpectedResponse", corpo, err)
		}
	}
}

func TestRedirecionamentoNaoESeguido(t *testing.T) {
	var chamadas int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chamadas++
		http.Redirect(w, r, "https://outro.exemplo/x", http.StatusMovedPermanently)
	}))
	t.Cleanup(srv.Close)
	cli, _ := NewClient(srv.URL, "sk", nil)
	_, err := cli.Telephony.OriginateCall(context.Background(), OriginateCallRequest{FromExtension: "1", ToNumber: "2"})
	u, okErr := AsUnexpectedResponseError(err)
	if !okErr || u.StatusCode != 301 || !strings.Contains(u.Reason, "outro.exemplo") {
		t.Fatalf("err = %v", err)
	}
	if chamadas != 1 {
		t.Fatalf("chamadas = %d, want 1", chamadas)
	}
}

func TestOnResponseRecebeRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "120")
		w.Header().Set("X-RateLimit-Remaining", "119")
		_, _ = w.Write([]byte(ok(`{"total":0,"items":[]}`)))
	}))
	t.Cleanup(srv.Close)
	var mu sync.Mutex
	var info ResponseInfo
	cli, _ := NewClient(srv.URL, "sk", &Options{OnResponse: func(i ResponseInfo) { mu.Lock(); info = i; mu.Unlock() }})
	if _, err := cli.Omni.ListAgents(context.Background(), 0, 0); err != nil {
		t.Fatal(err)
	}
	if info.RateLimit == nil || info.RateLimit.Limit != 120 || info.RateLimit.Remaining != 119 || info.StatusCode != 200 {
		t.Fatalf("info = %+v", info)
	}
}

func assinar(segredo string, ts int64, corpo []byte) string {
	mac := hmac.New(sha256.New, []byte(segredo))
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "."))
	mac.Write(corpo)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerificacaoDaAssinaturaDoWebhook(t *testing.T) {
	corpo := []byte(`{"events":[{"id":"e1","type":"call.ended","created_at":"2026-10-27T15:00:00Z","call":{"call_id":"AC-1","status":"completed"}}]}`)
	agora := time.Unix(1761594000, 0)
	ts := agora.Unix()
	header := "t=" + strconv.FormatInt(ts, 10) + ",v1=" + assinar("novo", ts, corpo) + ",v1=" + assinar("antigo", ts, corpo)

	if err := VerifyWebhookSignatureAt(header, corpo, agora, 0, "antigo"); err != nil {
		t.Errorf("segredo anterior deveria conferir: %v", err)
	}
	if err := VerifyWebhookSignatureAt(header, corpo, agora, 0, "outro", "novo"); err != nil {
		t.Errorf("segredo atual deveria conferir: %v", err)
	}
	if err := VerifyWebhookSignatureAt(header, corpo, agora, 0, "errado"); !errors.Is(err, ErrInvalidWebhookSignature) {
		t.Errorf("segredo errado: err = %v", err)
	}
	if err := VerifyWebhookSignatureAt(header, corpo, agora.Add(301*time.Second), 0, "novo"); !errors.Is(err, ErrInvalidWebhookSignature) {
		t.Errorf("fora da tolerância: err = %v", err)
	}
	if err := VerifyWebhookSignatureAt(header, append(corpo, ' '), agora, 0, "novo"); !errors.Is(err, ErrInvalidWebhookSignature) {
		t.Errorf("corpo alterado: err = %v", err)
	}
	if err := VerifyWebhookSignatureAt("v1=abc", corpo, agora, 0, "novo"); !errors.Is(err, ErrInvalidWebhookSignature) {
		t.Errorf("sem t: err = %v", err)
	}

	lote, err := DecodeWebhook(corpo)
	if err != nil || len(lote.Events) != 1 || lote.Events[0].Type != WebhookCallEnded || lote.Events[0].Call.CallID != "AC-1" {
		t.Fatalf("lote = %+v, err = %v", lote, err)
	}
}
