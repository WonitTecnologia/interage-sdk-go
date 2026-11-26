# interage-sdk-go/v2

SDK Go oficial da **API pública v2** da plataforma **Interage+** (Wonit), em `/api/public/v2`.

Todas as respostas são tipadas, as listas têm iterador que segue a paginação sozinho e os erros
trazem um código estável.

> A **v1** da API (`/api/public`) continua no módulo `github.com/WonitTecnologia/interage-sdk-go`
> e **deixa de funcionar em 15/03/2027**. Veja [Migrar da v1](#migrar-da-v1).

---

## Instalação

```bash
go get github.com/WonitTecnologia/interage-sdk-go/v2
```

Requer Go 1.24 ou mais novo.

## Token

Crie um token em **Admin › Configurações › Geral › API Tokens** e libere as operações que a
integração usa. Todo token novo é da v2. Um token da v1 recebe `403 TOKEN_VERSION_MISMATCH`.

## Início rápido

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	interage "github.com/WonitTecnologia/interage-sdk-go/v2"
)

func main() {
	// Domínio do seu tenant (https:// é adicionado automaticamente).
	cli, err := interage.NewClient("SEU-TENANT.wonit.cloud", "sk_xxxxxxxxxxxx", nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Confere o token: tenant, validade, permissões e limite.
	id, err := cli.Identify(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("conectado em", id.Tenant.Domain)

	// Ligações dos últimos 7 dias, todas as páginas.
	fim := time.Now()
	for ligacao, err := range cli.Calls.All(ctx, interage.ListCallsParams{From: fim.AddDate(0, 0, -7), To: fim}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(ligacao.ID, ligacao.StartedAt)
	}
}
```

## Recursos

| Campo do cliente | Rotas | Métodos |
|---|---|---|
| `Identify` | `GET /` | identificação do token |
| `Extensions` | `/extensions` | `List`, `All`, `Get`, `Create`, `Update`, `Delete`, `SetDND`, `SetForwarding`, `ResetPassword` |
| `Calls` | `/calls` | `List`, `All`, `Get`, `ListActive`, `AllActive`, `Originate` (click to call), `RecordingLink` |
| `Contacts` | `/contacts` | `List`, `All`, `Get`, `Create`, `Update`, `Delete`, `Batch` |
| `Instances` | `/instances` | `List`, `All` |
| `Templates` | `/templates` | `List`, `All` |
| `Queues` | `/queues` | `List`, `All` |
| `Users` | `/users` | `List`, `All`, `Get` |
| `Conversations` | `/conversations` | `List`, `All` (histórico), `ListActive`, `AllActive`, `Get`, `Start`, `Close`, `Transfer`, `ListMessages`, `AllMessages`, `GetMessage`, `SendMessage`, `FileLink` |
| `Messages` | `/messages` | `SendTemplate` (template avulso, fora de conversa) |
| `Campaigns` | `/campaigns` | `List`, `All`, `Get`, `Create`, `CreateFromCSV`, `Start`, `Pause`, `Cancel`, `Delete`, `ListContacts`, `AllContacts` |

A referência de cada campo está no `openapi.json` da v2 (`https://SEU-TENANT.wonit.cloud/api/public/v2/docs/openapi.json`)
e em [docs.wonit.cloud](https://docs.wonit.cloud/api-docs).

## Paginação

`List` devolve uma página (`Items`, `NextCursor`, `Total`). `All` percorre todas, pedindo a
próxima só quando a anterior acaba. Para no primeiro erro, que vem como o último par:

```go
for contato, err := range cli.Contacts.All(ctx, interage.ListContactsParams{Phone: "5511999998888"}) {
	if err != nil {
		return err
	}
	fmt.Println(contato.Name)
}
```

Para controlar a paginação à mão, passe `Paging{PageSize: 100, Cursor: pagina.NextCursor}`.

## Erros

Todo erro da API é um `*interage.Error` (RFC 9457): `Status`, `Code` (estável, em inglês),
`Title` e `Detail` (pt-BR), e `Errors` com **todos** os campos recusados de uma vez.

```go
_, err := cli.Conversations.Start(ctx, pedido)
var e *interage.Error
switch {
case interage.IsCode(err, interage.CodeConversationAlreadyActive):
	var protocolo string
	e, _ = err.(*interage.Error)
	e.Extra("protocol", &protocolo) // a conversa que já está ativa
case errors.As(err, &e) && e.Code == interage.CodeValidationError:
	for _, campo := range e.Errors {
		fmt.Println(campo.Field, campo.Message)
	}
case interage.IsCode(err, interage.CodeRateLimitExceeded):
	time.Sleep(e.RetryAfter) // e.RateLimit traz os headers X-RateLimit-*
}
```

Decida pelo `Code`, não pelo texto de `Detail`. A lista completa está nas constantes `Code*`.

O SDK não repete requisições sozinho. Nos envios (`Messages.SendTemplate`,
`Conversations.Start`, `Conversations.SendMessage`), `CodeProviderRejected` (502) é recusa do
provedor e não adianta repetir; `CodeProviderUnavailable` (503) é falha de comunicação: o envio
**não foi confirmado e pode ter chegado**. Consulte a mensagem pelo extra `message_id`, quando
vier, antes de enviar de novo.

## Alteração parcial (`Opt`)

No `Update` de contato e de ramal e no `SetForwarding`, a API distingue **campo ausente** (fica
como está), **null** (limpa ou desliga) e **valor**. Esses campos são `interage.Opt[T]`:

```go
cli.Contacts.Update(ctx, id, interage.ContactUpdate{
	Email:   interage.Set("joao@empresa.com"), // troca
	Company: interage.Null[string](),          // limpa
	Labels:  interage.Set([]int64{}),          // tira todas as etiquetas
	// Name não informado: não muda
})

cli.Extensions.SetForwarding(ctx, "1001", interage.ExtensionForwardingRequest{
	Always:   interage.Set("11999990000"), // liga o desvio sempre
	NoAnswer: interage.Null[string](),     // desliga o desvio de não atendida
})
```

## Campanhas

```go
// Contatos de grupos.
cp, err := cli.Campaigns.Create(ctx, interage.CampaignCreateRequest{
	Name: "Black Friday", InstanceID: instancia, GroupIDs: []int32{3},
	Template: interage.CampaignCreateRequestTemplate{ID: template},
}, &interage.CreateCampaignOptions{IdempotencyKey: uuid.NewString()})

// Contatos de um CSV (até 20 MB e 200.000 linhas); CollisionPolicy é obrigatório.
f, _ := os.Open("contatos.csv")
cp, err = cli.Campaigns.CreateFromCSV(ctx, pedido, "contatos.csv", f, &interage.CreateCampaignOptions{IdempotencyKey: chave})
```

Com a mesma `IdempotencyKey`, repetir a criação (ex.: depois de um timeout) devolve a campanha já
criada, em vez de disparar duas vezes. Guarde a chave e repita-a ao tentar de novo.

## Webhooks

Os webhooks (Admin › Integrações › Webhooks) entregam os eventos no formato da v2. O SDK confere a
assinatura `X-Interage-Signature` e decodifica o lote:

```go
http.HandleFunc("/interage", func(w http.ResponseWriter, r *http.Request) {
	lote, err := interage.ParseWebhook(r, os.Getenv("INTERAGE_WEBHOOK_SECRET"), 0)
	if err != nil {
		http.Error(w, "assinatura inválida", http.StatusUnauthorized)
		return
	}
	for _, ev := range lote.Events { // na ordem do array
		switch ev.Type {
		case interage.EventMessageCreated:
			fmt.Println(ev.Conversation.Protocol, *ev.Message.Content)
		case interage.EventCallEnded:
			fmt.Println(ev.Call.ID)
		case interage.EventExtensionStatusChanged:
			fmt.Println(ev.Extension.Number, ev.Extension.PreviousStatus, "→", ev.Extension.Status)
		}
	}
	w.WriteHeader(http.StatusNoContent)
})
```

`ev.ID` se repete só em reentrega: use-o para não processar o mesmo evento duas vezes. A
tolerância padrão de horário da assinatura é de 5 minutos (`DefaultWebhookTolerance`). Na troca
do segredo, a API assina com os dois segredos por 24 h, e qualquer um vale.

## Opções

```go
cli, err := interage.NewClient("SEU-TENANT.wonit.cloud", token, &interage.Options{
	Timeout:    10 * time.Second, // padrão 30 s
	HTTPClient: meuClienteHTTP,   // proxy, transporte próprio, retentativas
})
```

O SDK não refaz requisição sozinho. Leituras (`GET`) podem ser repetidas com segurança; para
criar campanha, use a `IdempotencyKey`.

## Migrar da v1

1. Crie um token novo (todo token novo é da v2) com as operações que a integração usa.
2. Troque o import para `github.com/WonitTecnologia/interage-sdk-go/v2`. Os dois módulos
   convivem no mesmo projeto: dá para migrar um recurso por vez, cada um com o seu token.
3. Ajuste as chamadas:

| v1 (`interage-sdk-go`) | v2 (`interage-sdk-go/v2`) |
|---|---|
| `Telephony.ListExtensions` | `Extensions.List` / `All` (e `Get`, `Create`, `Update`, `SetDND`, `SetForwarding`... novos) |
| `Telephony.ListCallHistory` | `Calls.List` / `All` (cursor, `From`/`To` com fuso) |
| `Telephony.ListActiveCalls` | `Calls.ListActive` / `Calls.AllActive` (paginado) |
| `Telephony.OriginateCall` | `Calls.Originate(ctx, ramal, destino)` |
| `Telephony.CreateRecordingTempLink` | `Calls.RecordingLink` |
| `Contacts.ListContacts` / `GetContact` | `Contacts.List` / `Get` |
| `Contacts.GetContactByPhone` | `Contacts.List(ctx, ListContactsParams{Phone: ...})` |
| `Omni.BatchCreateContacts` | `Contacts.Batch` |
| `Omni.ListQueues` | `Queues.List` (filas de chat e de voz, `Channel`) |
| `Omni.ListAgents` | `Users.List` (por e-mail, com o estado no chat e na voz) |
| `Omni.ListConversations` | `Conversations.ListActive` (em andamento) ou `Conversations.List` (histórico) |
| `Omni.GetConversationHistory` | `Conversations.ListMessages` / `AllMessages` |
| `Omni.CloseConversation` | `Conversations.Close` |
| `Omni.TransferConversation` | `Conversations.Transfer` (fila e usuário pelo id de `/queues` e `/users`) |
| `Omni.CreateMessageFileTempLink` | `Conversations.FileLink(ctx, protocolo, id, validade)` |
| `Messages.ListInstances` / `ListTemplates` | `Instances.List` / `Templates.List` |
| `Messages.SendTemplate` | `Messages.SendTemplate` (avulso) ou `Conversations.Start` (abre conversa) |
| `Messages.SendMessage` | `Conversations.SendMessage(ctx, protocolo, ...)` |
| `Messages.GetMessageStatus` | `Conversations.GetMessage(ctx, protocolo, id)` |
| `Campaigns.*` | `Campaigns.*` (e `CreateFromCSV`, `ListContacts`, `IdempotencyKey`) |
| `*APIError` + `ErrNotFound`... | `*Error` com `Code` (`IsCode(err, CodeNotFound)`) |
| `page`/`page_size` | cursor (`All` ou `Paging.Cursor`) |

O que muda no formato: nomes em inglês por recurso, ids públicos (uuid) em filas, usuários,
contatos e instâncias, horários com fuso, `null` em vez de string vazia e erros RFC 9457.

## Desenvolvimento

Os tipos (`types_gen.go`) são gerados do código dos DTOs da API, com as mesmas tags JSON:

```bash
go run ./internal/gerartipos -api ../../wonit-pabx-app-api
cp ../../wonit-pabx-app-api/docs/publicapi/v2/publicv2_swagger.json testdata/openapi.json
go test ./...
```

Os testes falham quando o SDK e o `openapi.json` se afastam: rota sem método, campo a mais ou a
menos em um tipo e código de erro novo.

## Licença

MIT — veja [LICENSE](../LICENSE).
