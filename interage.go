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

// Options configura opções adicionais do cliente. Todos os campos são opcionais.
type Options struct {
	// Timeout do http.Client interno. Padrão: 30s. Ignorado se HTTPClient for informado.
	Timeout time.Duration
	// HTTPClient substitui o http.Client interno (proxy, transporte customizado, etc.).
	HTTPClient *http.Client
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
	// Telephony — ramais, histórico de ligações, click-to-call e gravações.
	Telephony TelephonyCase
}

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

	var timeout time.Duration
	var custom *http.Client
	if opts != nil {
		timeout = opts.Timeout
		custom = opts.HTTPClient
	}

	hc := newHTTPClient(baseURL, token, timeout, custom)
	return &Client{
		Campaigns: newCampaignsClient(hc),
		Messages:  newMessagesClient(hc),
		Omni:      newOmniClient(hc),
		Telephony: newTelephonyClient(hc),
	}, nil
}
