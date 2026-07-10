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
> configuradas **por token** no painel. Sem a permissão, a API responde 401/403.

---

## Domínios

| Campo do client | Interface | Cobre |
|---|---|---|
| `cli.Campaigns` | `CampaignsCase` | Campanhas de disparo WhatsApp (criar com CSV, listar, detalhar, iniciar, pausar, cancelar, remover) |
| `cli.Messages` | `MessagesCase` | Instâncias, templates HSM, envio de template/mensagem, status de entrega |
| `cli.Omni` | `OmniCase` | Filas, agentes, conversas (listar, histórico, encerrar, transferir, arquivos) |
| `cli.Telephony` | `TelephonyCase` | Ramais, histórico de ligações, click-to-call, gravações |

Todos os modelos (requests/responses) vivem no **pacote principal** — um único import:

```go
import interage "github.com/WonitTecnologia/interage-sdk-go"

// interage.CreateCampaignRequest, interage.SendTemplateRequest,
// interage.TransferConversationRequest, interage.OriginateCallRequest, ...
```

Cada arquivo de domínio (`campaigns.go`, `messages.go`, `omni.go`, `telephony.go`) é
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

---

## Telephony — ramais, histórico e click-to-call

```go
ramais, err := cli.Telephony.ListExtensions(ctx, interage.ListExtensionsParams{Search: "10"})

hist, err := cli.Telephony.ListCallHistory(ctx, interage.ListCallHistoryParams{
	DateFrom: "2026-07-01", DateTo: "2026-07-09", // obrigatórios, máx. 3 meses
	CallResult: "ANSWERED",
})

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
		// token inválido ou sem a permissão exigida
	case errors.Is(err, interage.ErrUnprocessable):
		// ação não permitida no estado atual (ex.: iniciar campanha cancelada)
	}

	if apiErr, ok := interage.AsAPIError(err); ok {
		log.Println(apiErr.StatusCode, apiErr.Status, apiErr.Message)
	}
}
```

Sentinelas disponíveis: `ErrBadRequest` (400), `ErrUnauthorized` (401), `ErrForbidden` (403),
`ErrNotFound` (404), `ErrConflict` (409), `ErrUnprocessable` (422), `ErrInternalServer` (5xx).

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
