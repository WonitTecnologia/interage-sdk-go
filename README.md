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
| `Options.Timeout` | não | Timeout do HTTP client interno (padrão: 30s) |
| `Options.HTTPClient` | não | `*http.Client` customizado (proxy, transporte próprio, etc.) |
| `Options.Insecure` | não | Força HTTP (sem TLS). Padrão: `false`. Use para dev local. |

> As permissões de cada rota (leitura, listagem, criação, alteração, remoção) são
> configuradas **por token** no painel. Token ausente, inválido, expirado ou
> revogado responde **401**; token válido sem a rota ou sem a ação responde **403**.

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
(`phone`, `telefone`, `numero`, `celular` ou `whatsapp`). Colunas opcionais: `name`, `email`, `company`.

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
})
// resp.CampaignID, resp.MailingID, resp.Status ("pending" — importação assíncrona)
```

**`CollisionPolicy`** — o que fazer quando o contato do CSV já existe na base
(identificado por canal + número; em qualquer opção o número entra na campanha):

| Valor | Comportamento |
|---|---|
| `interage.CollisionIgnore` | Mantém os dados atuais do contato |
| `interage.CollisionOverwrite` | Sobrescreve nome/email/empresa com o CSV |
| `interage.CollisionUpdateEmpty` | Preenche só os campos vazios |

Campos opcionais: `Description`, `StartAt`/`EndAt` (RFC3339), `AutoStart *bool`, `Settings map[string]any`.

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

---

## Messages — instâncias, templates e envio

```go
// Instâncias (fonte do instance_id)
inst, err := cli.Messages.ListInstances(ctx, 1, 50)

// Templates HSM (use o GupshupID como template_id)
tpls, err := cli.Messages.ListTemplates(ctx, interage.ListTemplatesParams{Status: "APPROVED"})

// Enviar template
env, err := cli.Messages.SendTemplate(ctx, interage.SendTemplateRequest{
	To:         "5547999999999",
	TemplateID: "<gupshup_id>",
	InstanceID: "<instance_id>",
	Params:     []string{"João"},
})

// Enviar mensagem em sessão ativa (24h)
msg, err := cli.Messages.SendMessage(ctx, interage.SendMessageRequest{
	Protocol: "<protocolo>", Type: "text", Text: "Olá!",
})

// Status de entrega
st, err := cli.Messages.GetMessageStatus(ctx, "<internal_message_id>")
```

---

## Omni — filas, agentes e conversas

```go
filas, err := cli.Omni.ListQueues(ctx, 1, 20)
agentes, err := cli.Omni.ListAgents(ctx, 1, 20)

convs, err := cli.Omni.ListConversations(ctx, interage.ListConversationsParams{Status: "in_progress"})

hist, err := cli.Omni.GetConversationHistory(ctx, "<protocolo>", 1, 50)

fim, err := cli.Omni.CloseConversation(ctx, "<protocolo>")

queueID := 3
tr, err := cli.Omni.TransferConversation(ctx, "<protocolo>", interage.TransferConversationRequest{
	TargetType: interage.TransferToQueue,
	QueueID:    &queueID,
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

`Labels` recebe o **número** de cada etiqueta — o número sequencial exibido na
gestão de etiquetas da central (Admin → Contatos → Etiquetas), não o nome. A
etiqueta precisa **já existir e estar ativa** no tenant — etiqueta inexistente
gera erro no item (`Items[i].Status == "error"`) e aquele contato não é criado
nem atualizado.

`CustomInfo` aceita só texto: números e datas vão como string
(ex.: `"data_nascimento": "1990-05-20"`). As chaves são os `key` dos campos
personalizados cadastrados na central.

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

`GetContactByPhone` normaliza o número para dígitos e busca nos canais `phone`,
`whatsapp` e `sms`. Os telefones cadastrados aparecem em `contato.Identities`.

> A criação de contatos em lote está no domínio Omni: `cli.Omni.BatchCreateContacts`.

---

## Telephony — ramais, histórico e click-to-call

```go
ramais, err := cli.Telephony.ListExtensions(ctx, interage.ListExtensionsParams{Search: "10"})

hist, err := cli.Telephony.ListCallHistory(ctx, interage.ListCallHistoryParams{
	DateFrom: "2026-07-01", DateTo: "2026-07-09", // obrigatórios, máx. 3 meses
	CallResult: "ANSWERED",
})

ativas, err := cli.Telephony.ListActiveCalls(ctx) // chamadas em curso (não paginada)

call, err := cli.Telephony.OriginateCall(ctx, interage.OriginateCallRequest{
	FromExtension: "1000",
	ToNumber:      "5547999999999",
})

grav, err := cli.Telephony.CreateRecordingTempLink(ctx, "<call_id>", 3600)
```

---

## Tratamento de erros

Todo erro HTTP vira um `*interage.APIError`, que faz `Unwrap()` para um erro sentinela:

```go
camp, err := cli.Campaigns.Get(ctx, id)
if err != nil {
	switch {
	case errors.Is(err, interage.ErrNotFound):
		// campanha não existe
	case errors.Is(err, interage.ErrUnauthorized):
		// token ausente, inválido, expirado ou revogado — corrigir a credencial
	case errors.Is(err, interage.ErrForbidden):
		// token válido sem a rota ou sem a ação (ex.: listagem) — liberar no painel
	case errors.Is(err, interage.ErrUnprocessable):
		// ação não permitida no estado atual (ex.: iniciar campanha cancelada)
	}

	if apiErr, ok := interage.AsAPIError(err); ok {
		log.Println(apiErr.StatusCode, apiErr.Status, apiErr.Message)
	}
}
```

Sentinelas disponíveis: `ErrBadRequest` (400), `ErrUnauthorized` (401), `ErrForbidden` (403),
`ErrNotFound` (404), `ErrConflict` (409), `ErrUnprocessable` (422),
`ErrTooManyRequests` (429), `ErrInternalServer` (5xx).

### Limite de requisições (429)

A API limita as requisições por token e por IP de origem. Ao exceder, responde
429 e informa quanto esperar — disponível em `APIError.RetryAfter`:

```go
lista, err := cli.Contacts.ListContacts(ctx, params)
if errors.Is(err, interage.ErrTooManyRequests) {
	apiErr, _ := interage.AsAPIError(err)
	time.Sleep(apiErr.RetryAfter) // zero quando a API não informou
	lista, err = cli.Contacts.ListContacts(ctx, params)
}
```

O SDK não repete a requisição sozinho — a decisão de esperar e tentar de novo
fica com quem chama.

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

## Licença

MIT — © Wonit Tecnologia da Informação
