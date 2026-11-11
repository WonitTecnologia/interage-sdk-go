# interage-sdk-go

SDK Go oficial da **API pública de clientes** da plataforma **Interage+** (Wonit).

Todas as respostas são **tipadas** — você nunca precisa fazer parsing manual de JSON.

---

## Instalação

```bash
go get github.com/WonitTecnologia/interage-sdk-go
```

## Início rápido

```go
package main

import (
	"context"
	"fmt"
	"log"

	interage "github.com/WonitTecnologia/interage-sdk-go"
)

func main() {
	// Domínio do seu tenant (https:// é adicionado automaticamente).
	cli, err := interage.NewClient("SEU-TENANT.wonit.cloud", "sk_xxxxxxxxxxxx", nil)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	campanhas, err := cli.Campaigns.List(ctx, interage.ListCampaignsParams{})
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range campanhas.Items {
		fmt.Println(c.Name, c.Status, c.TotalContacts)
	}
}
```

### Configuração

`NewClient(baseURL, token string, opts *interage.Options)`:

| Campo | Obrigatório | Descrição |
|---|---|---|
| `baseURL` | sim | Domínio do tenant (ex.: `SEU-TENANT.wonit.cloud`). O SDK adiciona `https://` automaticamente. |
| `token` | sim | Token de API `sk_<valor>` (painel administrativo → Tokens de API) |
| `Options.Timeout` | não | Tempo máximo de cada requisição (padrão: 30s) |
| `Options.ConnectTimeout` | não | Tempo máximo para abrir a conexão (padrão: 10s) |
| `Options.HTTPClient` | não | `*http.Client` customizado (proxy, transporte próprio, etc.). Ignora `Timeout`/`ConnectTimeout`; configure `CheckRedirect` para devolver `http.ErrUseLastResponse` |
| `Options.Insecure` | não | Força HTTP (sem TLS). Padrão: `false`. Use para dev local. |
| `Options.OnResponse` | não | `func(interage.ResponseInfo)` chamada a cada resposta: status, cota (`X-RateLimit-*`) e `Idempotency-Replayed` |

O SDK envia o token no header `Authorization: Token sk_...`. A API também aceita
`X-API-Key: sk_...` e `?token=sk_...` na URL — este último só para quem não
consegue mandar header, como um link de click to call em outro sistema; para esse
uso, crie um token liberado **somente** para `/pabx/telefonia/originate`.

> As permissões de cada rota (leitura, listagem, criação, alteração, remoção) são
> configuradas **por token** no painel. Token ausente, inválido, expirado ou
> revogado responde **401**. **403** é token sem a rota ou sem a ação, cliente
> desativado ou módulo não licenciado — `APIError.Message` diz qual.

O cliente interno **não segue redirecionamentos** (um 3xx viraria GET): ele volta
como `ErrUnexpectedResponse` com o destino na mensagem — em geral, `baseURL` com
`http://` em vez de `https://`.

---

## Domínios

| Campo do client | Interface | Cobre |
|---|---|---|
| `cli.Campaigns` | `CampaignsCase` | Campanhas de disparo WhatsApp (criar com CSV, listar, detalhar, iniciar, pausar, cancelar, remover) |
| `cli.Messages` | `MessagesCase` | Instâncias, templates HSM, envio de template/mensagem, status de entrega |
| `cli.Omni` | `OmniCase` | Filas, agentes, conversas (listar, histórico, encerrar, transferir, arquivos), contatos em lote |
| `cli.Contacts` | `ContactsCase` | Central de contatos (listar, buscar por UUID ou telefone) |
| `cli.Telephony` | `TelephonyCase` | Ramais, histórico de ligações, chamadas ativas, click-to-call, gravações |

Todos os modelos (requests/responses) vivem no **pacote principal** — um único import:

```go
import interage "github.com/WonitTecnologia/interage-sdk-go"

// interage.CreateCampaignRequest, interage.SendTemplateRequest,
// interage.TransferConversationRequest, interage.OriginateCallRequest, ...
```

Cada arquivo de domínio (`campaigns.go`, `messages.go`, `omni.go`, `contacts.go`, `telephony.go`) é
organizado pelas seções **Modelos / Interface / Implementação**, separadas pelo divisor `─────`.
Quando um nome de tipo colide entre domínios, o tipo recebe o **domínio como prefixo**
(ex.: `TelephonyTempLinkResponse` e `OmniTempLinkResponse`).

---

## Campaigns — campanhas de disparo WhatsApp

### Criar campanha com CSV

O CSV deve ter delimitador `,` ou `;` e conter uma coluna de telefone
(`phone`, `telefone`, `numero`, `número`, `celular` ou `whatsapp`). Colunas opcionais: `name`/`nome`,
`email`, `company`/`empresa`, `cpf`/`documento`; as demais viram campos adicionais. Até 20 MB
(acima disso, `ErrPayloadTooLarge`) e 200.000 linhas.

```go
csv, _ := os.ReadFile("contatos.csv")

resp, err := cli.Campaigns.Create(ctx, interage.CreateCampaignRequest{
	Name:            "Black Friday",
	InstanceID:      "<instance_id>",           // de cli.Messages.ListInstances
	TemplateID:      "<gupshup_id>",            // de cli.Messages.ListTemplates
	TemplateParams:  []string{"João", "20/01"}, // variáveis do template, na ordem
	CollisionPolicy: interage.CollisionIgnore, // ignore | overwrite | update_empty
	FileName:        "contatos.csv",
	FileContent:     csv,
	IdempotencyKey:  "black-friday-2026", // opcional: repetir não cria outra campanha
})
// resp.CampaignID, resp.MailingID, resp.Status ("pending" — importação assíncrona)
// resp.Replayed == true: a campanha já existia com esta IdempotencyKey
```

**`CollisionPolicy`** — o que fazer quando o contato do CSV já existe na base
(identificado por canal + número; em qualquer opção o número entra na campanha):

| Valor | Comportamento |
|---|---|
| `interage.CollisionIgnore` | Mantém os dados atuais do contato |
| `interage.CollisionOverwrite` | Sobrescreve nome/email/empresa com o CSV |
| `interage.CollisionUpdateEmpty` | Preenche só os campos vazios |

Campos opcionais: `Description`, `StartAt`/`EndAt` (RFC3339), `AutoStart *bool`, `Settings map[string]any`,
`ReplyWithoutContext *bool`, `ReplyWindowHours int` (1 a 72; padrão da API: 24) e `IdempotencyKey`.
`Settings` aceita só `delay_ms`, `max_retries`, `retry_delay_seconds`, `dispatch_days` e
`dispatch_periods` (até 3 horários) — outra chave devolve 400. `TemplateParams` não aceita valor vazio.

Por padrão só conta como resposta à campanha a mensagem que cita o disparo ou clica num
botão do template. Com `ReplyWithoutContext`, a primeira mensagem do contato dentro de
`ReplyWindowHours` após o disparo também conta, desde que abra uma conversa nova.

### Listar, detalhar e controlar

```go
lista, err := cli.Campaigns.List(ctx, interage.ListCampaignsParams{
	Search: "black", Status: "ready", Page: 1, PageSize: 20,
})

camp, err := cli.Campaigns.Get(ctx, "<campaign_uuid>")

camp, err = cli.Campaigns.Start(ctx, "<campaign_uuid>")  // ready|paused|scheduled → running
camp, err = cli.Campaigns.Pause(ctx, "<campaign_uuid>")  // running → paused
camp, err = cli.Campaigns.Cancel(ctx, "<campaign_uuid>") // → canceled (não reinicia)

err = cli.Campaigns.Delete(ctx, "<campaign_uuid>") // bloqueada se running/processing
```

Ciclo de vida: `pending` → `processing` (importando CSV) → `ready` → `running` ⇄ `paused` → `completed`.
`Cancel` é permitido em qualquer status não-final; campanha cancelada não pode ser iniciada.

`Get` traz `camp.Import` com o resultado da importação do CSV (`TotalRows`, `FailedRows`
e, se a importação inteira falhou, `ErrorMessage`) — a listagem e as transições não trazem.
`camp.InstanceID` e `camp.AutoStart` vêm em todas.

---

## Messages — instâncias, templates e envio

```go
// Instâncias (fonte do instance_id)
inst, err := cli.Messages.ListInstances(ctx, 1, 50)

// Templates HSM (use o GupshupID como template_id)
tpls, err := cli.Messages.ListTemplates(ctx, interage.ListTemplatesParams{Status: "APPROVED"})

// Enviar template (sem atendimento: só env.MessageID, do provedor)
env, err := cli.Messages.SendTemplate(ctx, interage.SendTemplateRequest{
	To:             "5547999999999",
	TemplateID:     "<gupshup_id>",
	InstanceID:     "<instance_id>",
	Params:         []string{"João"},
	IdempotencyKey: "pedido-123-boas-vindas", // opcional, recomendado
})

// Template abrindo atendimento para um agente, numa fila
env, err = cli.Messages.SendTemplate(ctx, interage.SendTemplateRequest{
	To: "5547999999999", TemplateID: "<gupshup_id>", InstanceID: "<instance_id>",
	AssignedUserID: "<user_uuid>", QueueID: interage.Ptr[int32](3),
})
// env.InternalMessageID, env.ConversationID e env.Protocol só vêm com AssignedUserID

// Enviar mensagem em sessão ativa (24h)
msg, err := cli.Messages.SendMessage(ctx, interage.SendMessageRequest{
	Protocol: "<protocolo>", Type: "text", Text: "Olá!",
})

// Localização: ponteiros, para 0 também ser enviado
msg, err = cli.Messages.SendMessage(ctx, interage.SendMessageRequest{
	Protocol: "<protocolo>", Type: "location",
	Latitude: interage.Ptr(-26.91), Longitude: interage.Ptr(-48.66), LocationName: "Loja Centro",
})

// Status de entrega, pelo id interno
st, err := cli.Messages.GetMessageStatus(ctx, strconv.FormatInt(msg.InternalMessageID, 10))
```

**Reenvio seguro (`IdempotencyKey`).** Em `SendTemplate`, `SendMessage` e `Campaigns.Create`,
a chave vai no header `Idempotency-Key` (até 128 caracteres). Por 24 horas, repetir a mesma
chave devolve a resposta original, sem novo envio — `resp.Replayed` fica `true`. Gere uma chave
por envio e use a mesma ao repetir depois de um timeout ou de `ErrUnexpectedResponse`. Mesma
chave com outro conteúdo: `ErrUnprocessable` (`IDEMPOTENCY_KEY_REUSED`); envio ainda em
andamento: `ErrConflict`. Só o sucesso fica guardado.

**Provedor recusou (502).** O erro casa com `ErrSendFailed` (e `ErrInternalServer`); nada saiu.
`APIError.InternalMessageID` traz a mensagem gravada com a falha, quando houve registro —
consulte o motivo em `GetMessageStatus`.

---

## Omni — filas, agentes e conversas

```go
filas, err := cli.Omni.ListQueues(ctx, 1, 20)
agentes, err := cli.Omni.ListAgents(ctx, 1, 20)

// status: bot | ai_agent | queue | attending | coex
convs, err := cli.Omni.ListConversations(ctx, interage.ListConversationsParams{Status: "queue"})

// Mensagens com MediaFileName, Edited, RemovedAt e Reactions
hist, err := cli.Omni.GetConversationHistory(ctx, "<protocolo>", 1, 50)

fim, err := cli.Omni.CloseConversation(ctx, "<protocolo>")
// fila que exige classificação:
fim, err = cli.Omni.CloseConversationWith(ctx, "<protocolo>", interage.CloseConversationRequest{
	Classification: "Resolvido",
})

tr, err := cli.Omni.TransferConversation(ctx, "<protocolo>", interage.TransferConversationRequest{
	TargetType: interage.TransferToQueue,
	QueueID:    interage.Ptr(3),
})
// Para um agente, QueueID opcional troca também a fila:
tr, err = cli.Omni.TransferConversation(ctx, "<protocolo>", interage.TransferConversationRequest{
	TargetType: interage.TransferToUser, UserID: "<user_uuid>", QueueID: interage.Ptr(3),
})

link, err := cli.Omni.CreateMessageFileTempLink(ctx, "<message_id>", 3600)
```

### Criar contatos em lote

Cria até 500 contatos de uma vez na central de contatos. Cada item é processado
de forma independente — erros de um não impedem os demais (ver `Items` na resposta).

```go
resp, err := cli.Omni.BatchCreateContacts(ctx, interage.BatchCreateContactsRequest{
	CollisionPolicy: interage.CollisionIgnore, // opcional — padrão: ignore
	Contacts: []interage.BatchContactItem{
		{
			Name:  "João da Silva",
			Email: "joao@empresa.com",
			Identities: []interage.BatchContactIdentity{
				{Channel: "whatsapp", IDType: "phone", IDValue: "5511999998888", IsPrimary: true},
			},
			Labels:     []int64{1, 3}, // números das etiquetas, não os nomes
			CustomInfo: map[string]string{"codigo_cliente": "C-123"},
		},
	},
})
// resp.Created/Updated/Existing/Errors + resp.Items[i].Status/ContactID/Message
```

`CollisionPolicy` (mesma semântica das campanhas): `ignore` mantém o contato
existente, `overwrite` sobrescreve os campos informados, `update_empty` preenche
só os campos vazios. A colisão é detectada pelas identidades (canal + valor).

**Limites por contato:** `Name` 255 chars, `CPF` 14, `Email`/`Company` 255,
máx. 10 `Identities` (`IDValue` 255 chars), máx. 20 `Labels`, `CustomInfo` 10 KB.

Identidades aceitas: `phone`→`phone`; `whatsapp`→`phone`|`username`;
`telegram`→`chat_id`|`username`; `instagram`/`facebook`→`handle`|`profile_id`;
`sip`→`extension`. `CPF` vazio não é enviado; para apagar o CPF de um contato
existente com `CollisionOverwrite`, marque `ClearCPF: true`.

`Labels` recebe o **número** de cada etiqueta — o número sequencial exibido na
gestão de etiquetas da central (Admin → Contatos → Etiquetas), não o nome. A
etiqueta precisa **já existir e estar ativa** no tenant — etiqueta inexistente
gera erro no item (`Items[i].Status == "error"`) e aquele contato não é criado
nem atualizado.

`CustomInfo` aceita só texto — a API recusa número, booleano ou objeto: números e
datas vão como string (ex.: `"data_nascimento": "1990-05-20"`). As chaves são os
`key` dos campos personalizados cadastrados e ativos na central; chave desconhecida
é erro no item.

---

## Contacts — central de contatos

```go
lista, err := cli.Contacts.ListContacts(ctx, interage.ListContactsParams{
	Search: "joão", Page: 1, PageSize: 20,
})

contato, err := cli.Contacts.GetContact(ctx, "<contact_uuid>")

contato, err = cli.Contacts.GetContactByPhone(ctx, "5547999999999")
// ErrNotFound quando não existe contato com o número
```

Para percorrer a base inteira, use o cursor (ver [Paginação](#paginação)).

`GetContactByPhone` compara só os dígitos, nos canais `phone` e `whatsapp` (e `sms`
em cadastros antigos), com ou sem o DDI `55` e o 9º dígito. Número sem nenhum dígito
responde `ErrBadRequest`. Os telefones aparecem em `contato.Identities` e as
etiquetas em `contato.Labels` (`TenantSeq`, `Name`).

> A criação de contatos em lote está no domínio Omni: `cli.Omni.BatchCreateContacts`.

---

## Telephony — ramais, histórico e click-to-call

```go
ramais, err := cli.Telephony.ListExtensions(ctx, interage.ListExtensionsParams{Search: "10"})

hist, err := cli.Telephony.ListCallHistory(ctx, interage.ListCallHistoryParams{
	DateFrom: "2026-07-01", DateTo: "2026-07-09", // obrigatórios, máx. 3 meses
	CallResult: "ANSWERED",
	Extension:  "1001", // ramal exato; "10*" busca por prefixo
})

ativas, err := cli.Telephony.ListActiveCalls(ctx) // chamadas em curso (não paginada)

call, err := cli.Telephony.OriginateCall(ctx, interage.OriginateCallRequest{
	FromExtension: "1000",
	ToNumber:      "5547999999999",
})

grav, err := cli.Telephony.CreateRecordingTempLink(ctx, "<call_id>", 3600)
// ErrGone: gravação arquivada pela retenção — não há restauração
```

---

## Tratamento de erros

Todo erro HTTP vira um `*interage.APIError`, que faz `Unwrap()` para o(s) erro(s) sentinela:

```go
camp, err := cli.Campaigns.Get(ctx, id)
if err != nil {
	switch {
	case errors.Is(err, interage.ErrNotFound):
		// campanha não existe
	case errors.Is(err, interage.ErrUnauthorized):
		// token ausente, inválido, expirado ou revogado — corrigir a credencial
	case errors.Is(err, interage.ErrForbidden):
		// rota ou ação não liberada, cliente desativado ou módulo não licenciado (ver Message)
	case errors.Is(err, interage.ErrTooManyRequests), errors.Is(err, interage.ErrServiceUnavailable):
		// aguarde apiErr.RetryAfter (o SDK não repete sozinho)
	case errors.Is(err, interage.ErrUnexpectedResponse):
		// 3xx, ou 2xx vazio/sem JSON: numa ação, ela pode ter sido executada
	}

	if apiErr, ok := interage.AsAPIError(err); ok {
		log.Println(apiErr.StatusCode, apiErr.Status, apiErr.Message)
	}
}
```

| Sentinela | HTTP | Quando |
|---|---|---|
| `ErrBadRequest` | 400 | Parâmetros inválidos |
| `ErrUnauthorized` | 401 | Token ausente, inválido, expirado ou revogado |
| `ErrForbidden` | 403 | Token sem a rota ou sem a ação, cliente desativado ou módulo não licenciado |
| `ErrNotFound` | 404 | Recurso não encontrado |
| `ErrConflict` | 409 | Conflito de estado (ramal em chamada ou acionado nos últimos 35 s, mesma `IdempotencyKey` em andamento) |
| `ErrGone` | 410 | Gravação arquivada pela retenção — não volta |
| `ErrPayloadTooLarge` | 413 | CSV de campanha acima de 20 MB |
| `ErrUnprocessable` | 422 | Ação inválida no estado atual, janela de 24h expirada, instância inativa, `IdempotencyKey` com outro conteúdo |
| `ErrTooManyRequests` | 429 | Limite de requisições; espere `APIError.RetryAfter` |
| `ErrSendFailed` | 502 | Provedor recusou o envio; `APIError.InternalMessageID` |
| `ErrServiceUnavailable` | 503 | Indisponibilidade temporária; espere `APIError.RetryAfter`. Nas rotas de ação, nada foi executado |
| `ErrInternalServer` | 5xx | Qualquer 5xx (inclui 502 e 503) |
| `ErrUnexpectedResponse` | 3xx/2xx | Resposta fora do contrato (`*UnexpectedResponseError`) |

Se a resposta de erro não é o envelope da API (ex.: HTML de um proxy), `Message`
fica curta e o corpo (até 2 KB) vai para `APIError.RawBody`.

### Limite de requisições

Os limites valem para o **cliente (tenant)**, somando todos os tokens dele — criar
outro token não traz cota nova. Há duas cotas independentes: **ação** (originar
chamada, enviar mensagem ou template, criar e iniciar campanha, criar contatos em
lote — padrão 20 requisições recompostas ao longo de 60 s) e **padrão** (todas as
outras — 120 por 60 s). A cota se recompõe de forma contínua.

Cada resposta autenticada traz `X-RateLimit-Limit` e `X-RateLimit-Remaining`:
receba-os em `Options.OnResponse` (ou em `APIError.RateLimit`). Acima do limite,
429 com `Retry-After` em `APIError.RetryAfter`. Há ainda um teto por IP de origem,
e 20 tentativas por minuto com **token inexistente** bloqueiam o IP até a janela
fechar (tokens que autenticaram nos últimos 10 minutos continuam passando; token
expirado ou revogado não entra nessa conta). Nas rotas de ação, 503 indica que a
cota não pôde ser conferida: nada foi executado e dá para repetir.

```go
lista, err := cli.Contacts.ListContacts(ctx, params)
if apiErr, ok := interage.AsAPIError(err); ok && apiErr.RetryAfter > 0 {
	time.Sleep(apiErr.RetryAfter)
	lista, err = cli.Contacts.ListContacts(ctx, params)
}
```

O SDK não repete a requisição sozinho — a decisão de esperar e tentar de novo
fica com quem chama. Num POST de envio, repita só com a mesma `IdempotencyKey`.

---

## Webhooks

Cada entrega é um POST com `events[]` (até 50, processe em ordem) assinado em
`X-Interage-Signature: t=<unix>,v1=<hex>`. O SDK confere a assinatura (HMAC-SHA256
de `"<t>.<corpo bruto>"`, qualquer `v1` vale — durante a troca do segredo vêm dois —,
tolerância de 300 s, comparação em tempo constante) e decodifica os eventos:

```go
http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
	corpo, _ := io.ReadAll(r.Body) // corpo bruto: não reserialize o JSON
	lote, err := interage.ParseWebhook(r.Header.Get(interage.HeaderWebhookSignature), corpo,
		segredoAtual, segredoAnterior) // o anterior só durante a troca
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	for _, ev := range lote.Events {
		if jaProcessado(ev.ID) { // deduplique por ev.ID — Idempotency-Key muda a cada POST
			continue
		}
		switch ev.Type {
		case interage.WebhookMessageCreated: // ev.Conversation + ev.Message
		case interage.WebhookConversationClosed: // ev.Conversation
		case interage.WebhookCallEnded: // ev.Call
		case interage.WebhookCallRecordingAvailable: // ev.Call + ev.Recording.URL
		case interage.WebhookExtensionStatusChanged: // ev.Extension
		}
	}
	w.WriteHeader(http.StatusOK) // responda em menos de 5 s; processe em fila
})
```

`VerifyWebhookSignature` / `VerifyWebhookSignatureAt` só conferem; `DecodeWebhook` só decodifica.

---

## Paginação

As listagens aceitam `Page`/`PageSize` (padrão 10, máximo 100). Campanhas,
contatos e histórico de ligações também paginam por **cursor**, mais indicado
para percorrer listas grandes: passe o `NextCursor` da resposta como `Cursor`
da próxima chamada. `NextCursor` vazio indica a última página; com `Cursor`
informado, a API ignora `Page`.

```go
params := interage.ListContactsParams{PageSize: 100}
for {
	pagina, err := cli.Contacts.ListContacts(ctx, params)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range pagina.Items {
		fmt.Println(c.Name)
	}
	if pagina.NextCursor == "" {
		break
	}
	params.Cursor = pagina.NextCursor
}
```

---

## Exemplo completo — campanha de ponta a ponta

```go
cli, _ := interage.NewClient(baseURL, token, nil)
ctx := context.Background()

// 1. Descobrir instância e template
inst, _ := cli.Messages.ListInstances(ctx, 1, 10)
tpls, _ := cli.Messages.ListTemplates(ctx, interage.ListTemplatesParams{
	Status: "APPROVED", InstanceID: inst.Items[0].InstanceID,
})

// 2. Criar a campanha com o CSV
csv, _ := os.ReadFile("contatos.csv")
created, _ := cli.Campaigns.Create(ctx, interage.CreateCampaignRequest{
	Name:            "Campanha via SDK",
	InstanceID:      inst.Items[0].InstanceID,
	TemplateID:      tpls.Items[0].GupshupID,
	CollisionPolicy: interage.CollisionIgnore,
	FileName:        "contatos.csv",
	FileContent:     csv,
})

// 3. Aguardar a importação (pending → ready) e iniciar
for {
	camp, _ := cli.Campaigns.Get(ctx, created.CampaignID)
	if camp.Status == string(interage.CampaignStatusReady) {
		break
	}
	time.Sleep(2 * time.Second)
}
running, _ := cli.Campaigns.Start(ctx, created.CampaignID)
fmt.Println("Campanha em execução:", running.ID)
```

---

## Mudanças na v0.4.0

Incompatíveis (ajuste ao atualizar):

- `SendMessageRequest.Latitude`/`Longitude` viraram `*float64` (antes, `0` não era enviado e a API
  respondia 400). Use `interage.Ptr(-26.91)`.
- `APIError.Unwrap()` devolve `[]error` (502 e 503 casam com o sentinela específico e com
  `ErrInternalServer`). `errors.Is`/`errors.As` continuam iguais.
- `OmniCase` ganhou `CloseConversationWith` — quem implementa a interface (mocks) precisa do método.
- 2xx sem corpo, sem `data` ou fora do JSON agora é erro (`ErrUnexpectedResponse`); antes
  devolvia a struct vazia sem erro. O cliente interno não segue mais redirecionamentos.

Obsoletos (continuam compilando, sempre vazios): `CampaignResponse.ScheduledAt`,
`ContactResponse.Visibility`, `ActiveCallResponse.Channel` — a API não devolve mais esses campos.

Novos: `IdempotencyKey`/`Replayed`, `SendTemplateRequest.QueueID`, `CloseConversationRequest`,
`CampaignResponse.InstanceID`/`AutoStart`/`Import`, `ContactResponse.Labels`, mensagens do
histórico com `MediaFileName`/`Edited`/`RemovedAt`/`Reactions`, `ErrGone`, `ErrPayloadTooLarge`,
`ErrSendFailed`, `ErrServiceUnavailable`, `ErrUnexpectedResponse`, `APIError.Data`/
`InternalMessageID`/`RateLimit`/`RawBody`, `Options.ConnectTimeout`/`OnResponse`, webhooks
(`ParseWebhook` e tipos dos eventos) e `BatchContactItem.ClearCPF`.

---

## Licença

MIT — © Wonit Tecnologia da Informação
