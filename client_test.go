package interage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// captured guarda o que o servidor falso recebeu na última requisição.
type captured struct {
	method string
	path   string
	query  map[string]string
	header http.Header
	body   []byte
}

// fakeAPI sobe um servidor que responde com status/body fixos e registra a requisição.
func fakeAPI(t *testing.T, status int, headers map[string]string, body string) (*Client, *captured) {
	t.Helper()
	got := &captured{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.query = map[string]string{}
		for k, v := range r.URL.Query() {
			got.query[k] = v[0]
		}
		got.header = r.Header.Clone()
		got.body, _ = io.ReadAll(r.Body)
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	cli, err := NewClient(srv.URL, "sk_teste", nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return cli, got
}

func ok(data string) string {
	return `{"code":200,"status":"OK","message":"ok","data":` + data + `}`
}

func TestHeadersDeAutenticacaoEUserAgent(t *testing.T) {
	cli, got := fakeAPI(t, 200, nil, ok(`{"total":0,"items":[]}`))
	if _, err := cli.Messages.ListInstances(context.Background(), 0, 0); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if h := got.header.Get("Authorization"); h != "Token sk_teste" {
		t.Errorf("Authorization = %q, want %q", h, "Token sk_teste")
	}
	if h := got.header.Get("User-Agent"); h != "interage-sdk-go/"+Version {
		t.Errorf("User-Agent = %q, want %q", h, "interage-sdk-go/"+Version)
	}
	if _, inQuery := got.query["token"]; inQuery {
		t.Error("o token não pode ir na URL")
	}
}

// O servidor espera labels como números (tenant_seq) e custom_info só com texto —
// nomes de etiqueta ou valores não-texto são recusados com 400 na leitura do JSON.
func TestBatchCreateContacts_FormatoDeLabelsECustomInfo(t *testing.T) {
	cli, got := fakeAPI(t, 200, nil, ok(`{"total":1,"created":1,"items":[{"index":0,"status":"created"}]}`))
	_, err := cli.Omni.BatchCreateContacts(context.Background(), BatchCreateContactsRequest{
		CollisionPolicy: CollisionIgnore,
		Contacts: []BatchContactItem{{
			Name:       "Maria",
			Identities: []BatchContactIdentity{{Channel: "whatsapp", IDType: "phone", IDValue: "5511999998888", IsPrimary: true}},
			Labels:     []int64{1, 3},
			CustomInfo: map[string]string{"codigo_cliente": "C-123"},
		}},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	var body struct {
		Contacts []struct {
			Labels     []json.RawMessage          `json:"labels"`
			CustomInfo map[string]json.RawMessage `json:"custom_info"`
		} `json:"contacts"`
	}
	if err := json.Unmarshal(got.body, &body); err != nil {
		t.Fatalf("corpo inválido: %v", err)
	}
	labels := body.Contacts[0].Labels
	if len(labels) != 2 || string(labels[0]) != "1" || string(labels[1]) != "3" {
		t.Errorf("labels = %s, want números [1 3]", got.body)
	}
	if v := string(body.Contacts[0].CustomInfo["codigo_cliente"]); v != `"C-123"` {
		t.Errorf("custom_info.codigo_cliente = %s, want string", v)
	}
}

func TestPaginacaoPorCursor(t *testing.T) {
	page := ok(`{"total":30,"page":1,"page_size":10,"next_cursor":"prox-123","items":[]}`)
	ctx := context.Background()

	cases := []struct {
		name string
		call func(cli *Client) (string, error)
	}{
		{"campanhas", func(cli *Client) (string, error) {
			r, err := cli.Campaigns.List(ctx, ListCampaignsParams{Cursor: "cur-abc"})
			if err != nil {
				return "", err
			}
			return r.NextCursor, nil
		}},
		{"contatos", func(cli *Client) (string, error) {
			r, err := cli.Contacts.ListContacts(ctx, ListContactsParams{Cursor: "cur-abc"})
			if err != nil {
				return "", err
			}
			return r.NextCursor, nil
		}},
		{"historico", func(cli *Client) (string, error) {
			r, err := cli.Telephony.ListCallHistory(ctx, ListCallHistoryParams{DateFrom: "2026-09-01", DateTo: "2026-09-28", Cursor: "cur-abc"})
			if err != nil {
				return "", err
			}
			return r.NextCursor, nil
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cli, got := fakeAPI(t, 200, nil, page)
			next, err := tc.call(cli)
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if got.query["cursor"] != "cur-abc" {
				t.Errorf("cursor enviado = %q, want %q", got.query["cursor"], "cur-abc")
			}
			if next != "prox-123" {
				t.Errorf("NextCursor = %q, want %q", next, "prox-123")
			}
		})
	}
}

func TestPaginacaoSemCursorNaoEnviaParametro(t *testing.T) {
	cli, got := fakeAPI(t, 200, nil, ok(`{"total":0,"items":[]}`))
	if _, err := cli.Contacts.ListContacts(context.Background(), ListContactsParams{Page: 2}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if _, has := got.query["cursor"]; has {
		t.Error("cursor vazio não deve ir na query")
	}
	if got.query["page"] != "2" {
		t.Errorf("page = %q, want 2", got.query["page"])
	}
}

func TestLimiteDeRequisicoes(t *testing.T) {
	cli, _ := fakeAPI(t, 429, map[string]string{"Retry-After": "12"},
		`{"code":429,"status":"TOO_MANY_REQUESTS","message":"Limite de requisições excedido para este token de API."}`)

	_, err := cli.Omni.ListAgents(context.Background(), 0, 0)
	if !errors.Is(err, ErrTooManyRequests) {
		t.Fatalf("err = %v, want ErrTooManyRequests", err)
	}
	apiErr, isAPIErr := AsAPIError(err)
	if !isAPIErr {
		t.Fatal("esperava *APIError na cadeia")
	}
	if apiErr.RetryAfter != 12*time.Second {
		t.Errorf("RetryAfter = %v, want 12s", apiErr.RetryAfter)
	}
}

func TestMapeamentoDeErros(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{400, ErrBadRequest},
		{401, ErrUnauthorized},
		{403, ErrForbidden},
		{404, ErrNotFound},
		{409, ErrConflict},
		{422, ErrUnprocessable},
		{429, ErrTooManyRequests},
		{500, ErrInternalServer},
		{503, ErrInternalServer},
	}
	for _, tc := range cases {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			cli, _ := fakeAPI(t, tc.status, nil, `{"code":0,"status":"X","message":"falhou"}`)
			_, err := cli.Campaigns.Get(context.Background(), "abc")
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
			if apiErr, _ := AsAPIError(err); apiErr == nil || apiErr.Message != "falhou" {
				t.Errorf("mensagem da API não preservada: %v", err)
			}
		})
	}
}

func TestCampanhaCamposDeResposta(t *testing.T) {
	cli, got := fakeAPI(t, 200, nil, ok(`{"id":"c1","status":"running","reply_without_context":true,"reply_window_hours":48}`))
	c, err := cli.Campaigns.Get(context.Background(), "c1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got.path != "/api/public/whatsapp/campanhas/c1" {
		t.Errorf("path = %q", got.path)
	}
	if !c.ReplyWithoutContext || c.ReplyWindowHours != 48 {
		t.Errorf("ReplyWithoutContext=%v ReplyWindowHours=%d, want true/48", c.ReplyWithoutContext, c.ReplyWindowHours)
	}
}

// campaignForm decodifica o multipart capturado pelo fakeAPI.
func campaignForm(t *testing.T, got *captured) map[string]string {
	t.Helper()
	_, params, err := mime.ParseMediaType(got.header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("Content-Type inválido: %v", err)
	}
	form, err := multipart.NewReader(bytes.NewReader(got.body), params["boundary"]).ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("multipart inválido: %v", err)
	}
	out := map[string]string{}
	for k, v := range form.Value {
		out[k] = v[0]
	}
	return out
}

func baseCampaignRequest() CreateCampaignRequest {
	return CreateCampaignRequest{
		Name: "c", InstanceID: "i", TemplateID: "t", CollisionPolicy: CollisionIgnore,
		FileName: "c.csv", FileContent: []byte("phone\n5511999998888\n"),
	}
}

func TestCreateCampaignEnviaRespostaSemReply(t *testing.T) {
	cli, got := fakeAPI(t, 201, nil, ok(`{"campaign_id":"c1","status":"pending"}`))
	req := baseCampaignRequest()
	on := true
	req.ReplyWithoutContext = &on
	req.ReplyWindowHours = 48
	if _, err := cli.Campaigns.Create(context.Background(), req); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	form := campaignForm(t, got)
	if form["reply_without_context"] != "true" || form["reply_window_hours"] != "48" {
		t.Fatalf("campos de resposta sem reply = %q/%q, want true/48", form["reply_without_context"], form["reply_window_hours"])
	}
}

func TestCreateCampaignOmiteRespostaSemReplyNaoInformada(t *testing.T) {
	cli, got := fakeAPI(t, 201, nil, ok(`{"campaign_id":"c1","status":"pending"}`))
	if _, err := cli.Campaigns.Create(context.Background(), baseCampaignRequest()); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	form := campaignForm(t, got)
	for _, k := range []string{"reply_without_context", "reply_window_hours"} {
		if _, sent := form[k]; sent {
			t.Errorf("%s enviado sem ter sido informado", k)
		}
	}
}

func TestCreateCampaignRecusaJanelaForaDoIntervalo(t *testing.T) {
	cli, _ := fakeAPI(t, 201, nil, ok(`{}`))
	req := baseCampaignRequest()
	req.ReplyWindowHours = 73
	if _, err := cli.Campaigns.Create(context.Background(), req); err == nil {
		t.Fatal("ReplyWindowHours=73 deveria falhar antes de chamar a API")
	}
}

func TestGetMessageStatusRecusaIDNaoNumerico(t *testing.T) {
	cli, got := fakeAPI(t, 200, nil, ok(`{"id":1,"status":"sent"}`))
	if _, err := cli.Messages.GetMessageStatus(context.Background(), "abc"); err == nil {
		t.Fatal("messageID não numérico deveria falhar")
	}
	if got.method != "" {
		t.Fatal("a API não deveria ser chamada com messageID inválido")
	}
	if _, err := cli.Messages.GetMessageStatus(context.Background(), "1042"); err != nil {
		t.Fatalf("messageID numérico: erro inesperado %v", err)
	}
	if got.query["message_id"] != "1042" {
		t.Fatalf("message_id = %q, want 1042", got.query["message_id"])
	}
}
