// Package interage é o SDK Go oficial da API pública v2 da plataforma Interage+ (Wonit),
// em /api/public/v2.
//
//	cli, err := interage.NewClient("SEU-TENANT.wonit.cloud", "sk_xxxxxxxx", nil)
//	if err != nil { ... }
//
//	ramal, err := cli.Extensions.Get(ctx, "1001")
//
//	for conversa, err := range cli.Conversations.All(ctx, interage.ListConversationsParams{
//		From: inicio, To: fim,
//	}) {
//		if err != nil { ... }
//		fmt.Println(conversa.Protocol)
//	}
//
// O token precisa ser da v2 (criado em Admin › Configurações › Geral › API Tokens); token da
// v1 recebe 403 TOKEN_VERSION_MISMATCH. Os erros são *Error (RFC 9457), com Code estável.
//
// A v1 da API (/api/public) continua no módulo github.com/WonitTecnologia/interage-sdk-go e
// deixa de funcionar em 15/03/2027.
package interage

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

// Version é a versão do SDK, enviada no User-Agent (interage-sdk-go/<Version>).
const Version = "2.0.0"

// Options configura o cliente. Todos os campos são opcionais.
type Options struct {
	// Timeout do http.Client interno. Padrão: 30s. Ignorado se HTTPClient for informado.
	Timeout time.Duration
	// HTTPClient substitui o http.Client interno (proxy, transporte próprio etc.).
	HTTPClient *http.Client
	// Insecure usa http:// quando a baseURL vem sem esquema (desenvolvimento local).
	Insecure bool
}

// Client é o ponto de entrada do SDK. Cada campo cobre um recurso da API.
type Client struct {
	http *httpClient

	// Extensions — ramais: lista, detalhe, criar, alterar, excluir, não perturbe, desvios e senha SIP.
	Extensions *ExtensionsService
	// Calls — ligações: click to call, histórico, em curso, detalhe e link da gravação.
	Calls *CallsService
	// Contacts — contatos: lista, detalhe, criar, alterar, excluir e lote.
	Contacts *ContactsService
	// Instances — canais (instâncias) do omnichannel.
	Instances *InstancesService
	// Templates — templates de mensagem (HSM) do WhatsApp.
	Templates *TemplatesService
	// Queues — filas de chat e de voz.
	Queues *QueuesService
	// Users — usuários, com o estado no chat e na voz.
	Users *UsersService
	// Conversations — conversas: em andamento, histórico, detalhe, mensagens, iniciar, enviar, encerrar e transferir.
	Conversations *ConversationsService
	// Messages — template avulso (POST /messages), fora de conversa.
	Messages *MessagesService
	// Campaigns — campanhas de disparo: criar (grupos ou CSV), listar, iniciar, pausar, cancelar, remover e resultado por contato.
	Campaigns *CampaignsService
}

// Erros de configuração do NewClient.
var (
	ErrInvalidBaseURL = errors.New("interage: baseURL não pode ser vazia")
	ErrInvalidToken   = errors.New("interage: token não pode ser vazio")
)

// NewClient cria o cliente. baseURL é o domínio do tenant (ex.: SEU-TENANT.wonit.cloud); sem
// esquema, o SDK usa https:// (ou http:// com Options.Insecure). token é o token de API da
// v2 (sk_...).
func NewClient(baseURL, token string, opts *Options) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, ErrInvalidBaseURL
	}
	if strings.TrimSpace(token) == "" {
		return nil, ErrInvalidToken
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		esquema := "https://"
		if opts != nil && opts.Insecure {
			esquema = "http://"
		}
		baseURL = esquema + baseURL
	}
	var timeout time.Duration
	var custom *http.Client
	if opts != nil {
		timeout, custom = opts.Timeout, opts.HTTPClient
	}
	hc := newHTTPClient(baseURL+"/api/public/v2", token, timeout, custom)
	return &Client{
		http:          hc,
		Extensions:    &ExtensionsService{hc},
		Calls:         &CallsService{hc},
		Contacts:      &ContactsService{hc},
		Instances:     &InstancesService{hc},
		Templates:     &TemplatesService{hc},
		Queues:        &QueuesService{hc},
		Users:         &UsersService{hc},
		Conversations: &ConversationsService{hc},
		Messages:      &MessagesService{hc},
		Campaigns:     &CampaignsService{hc},
	}, nil
}
