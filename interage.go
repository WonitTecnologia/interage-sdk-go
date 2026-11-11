// Package interage é o SDK Go oficial da API pública de clientes da
// plataforma Interage+ (Wonit).
//
// Uso básico:
//
//	cli, err := interage.NewClient("SEU-TENANT.wonit.cloud", "sk_xxxxxxxx", nil)
//	if err != nil { ... }
//
//	campanhas, err := cli.Campaigns.List(ctx, interage.ListCampaignsParams{})
//
// Todos os modelos (requests/responses) vivem neste pacote — import único.
// Cada arquivo de domínio é dividido pelas seções Modelos / Interface / Implementação.
// Em colisão de nome entre domínios, o tipo recebe o domínio como prefixo
// (ex.: TelephonyTempLinkResponse e OmniTempLinkResponse).
package interage

import (
	"net/http"
	"strings"
	"time"
)

// Version é a versão do SDK, enviada no User-Agent (interage-sdk-go/<Version>).
const Version = "0.4.0"

// Options configura opções adicionais do cliente. Todos os campos são opcionais.
type Options struct {
	// Timeout do http.Client interno (requisição inteira). Padrão: 30s. Ignorado se HTTPClient for informado.
	Timeout time.Duration
	// ConnectTimeout limita só o estabelecimento da conexão TCP. Padrão: 10s.
	// Ignorado se HTTPClient for informado.
	ConnectTimeout time.Duration
	// HTTPClient substitui o http.Client interno (proxy, transporte customizado, etc.).
	// O cliente interno não segue redirecionamentos; com um HTTPClient próprio,
	// configure CheckRedirect para devolver http.ErrUseLastResponse.
	HTTPClient *http.Client
	// OnResponse, quando informado, é chamado a cada resposta HTTP com o status, o
	// consumo da cota (X-RateLimit-Limit / X-RateLimit-Remaining) e o
	// Idempotency-Replayed. Precisa ser seguro para uso concorrente.
	OnResponse func(ResponseInfo)
	// Insecure força HTTP (sem TLS) ao invés de HTTPS. Padrão: false.
	// Útil para ambientes de desenvolvimento local.
	Insecure bool
}

// Client é o ponto de entrada do SDK. Cada campo cobre um domínio da API pública.
type Client struct {
	// Campaigns — campanhas de disparo WhatsApp (criar com CSV, listar, iniciar, pausar, cancelar, remover).
	Campaigns CampaignsCase
	// Messages — instâncias, templates HSM e envio de mensagens WhatsApp.
	Messages MessagesCase
	// Omni — filas, agentes e conversas do atendimento omnichannel.
	Omni OmniCase
	// Contacts — central de contatos do tenant (listar, buscar por UUID ou telefone).
	Contacts ContactsCase
	// Telephony — ramais, histórico de ligações, click-to-call e gravações.
	Telephony TelephonyCase
}

// Ptr devolve um ponteiro para v — útil nos campos opcionais (ex.: Latitude, QueueID).
func Ptr[T any](v T) *T { return &v }

// NewClient cria o cliente do SDK.
//
// baseURL é o domínio do seu tenant (ex.: SEU-TENANT.wonit.cloud). Se informado
// sem scheme (ex.: sem https://), o SDK adiciona https:// automaticamente.
// Use Options.Insecure para forçar http:// (ambientes de desenvolvimento).
//
// token é o token de API no formato sk_<valor>, gerado no painel administrativo.
// As permissões de cada rota (leitura, listagem, criação, alteração, remoção)
// são configuradas por token no painel.
func NewClient(baseURL, token string, opts *Options) (*Client, error) {
	if baseURL == "" {
		return nil, ErrInvalidBaseURL
	}
	if token == "" {
		return nil, ErrInvalidToken
	}

	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		scheme := "https"
		if opts != nil && opts.Insecure {
			scheme = "http"
		}
		baseURL = scheme + "://" + baseURL
	}

	var timeout, connectTimeout time.Duration
	var custom *http.Client
	var onResponse func(ResponseInfo)
	if opts != nil {
		timeout = opts.Timeout
		connectTimeout = opts.ConnectTimeout
		custom = opts.HTTPClient
		onResponse = opts.OnResponse
	}

	hc := newHTTPClient(baseURL, token, timeout, connectTimeout, custom, onResponse)
	return &Client{
		Campaigns: newCampaignsClient(hc),
		Messages:  newMessagesClient(hc),
		Omni:      newOmniClient(hc),
		Contacts:  newContactsClient(hc),
		Telephony: newTelephonyClient(hc),
	}, nil
}
