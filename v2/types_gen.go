// Code generated from the Interage API v2 DTOs. DO NOT EDIT.
// Tipos da API pública v2, gerados a partir do código da API (mesmas tags JSON).

package interage

import (
	"encoding/json"
	"time"
)

var _ = json.RawMessage(nil)
var _ = time.Time{}

// Action é a permissão do token que a rota exige. Os valores são contrato (aparecem em
// permissions[].actions da identificação).
type Action string

// ActiveCall é a ligação em curso: o modelo sem ended_at, result, hangup_cause e abandono,
// com a duração até agora.
type ActiveCall struct {
	ID       string  `json:"id"`
	Protocol *string `json:"protocol"`
	// Valores: inbound, outbound, internal.
	Direction *string  `json:"direction"`
	From      CallFrom `json:"from"`
	To        CallTo   `json:"to"`
	DID       *string  `json:"did"`
	Trunk     *string  `json:"trunk"`
	// Valores: extension, queue, ivr, trunk, passthrough, whatsapp.
	Route           *string       `json:"route"`
	Extension       *string       `json:"extension"`
	Queue           *CallQueue    `json:"queue"`
	Agent           *CallAgent    `json:"agent"`
	StartedAt       time.Time     `json:"started_at"`
	AnsweredAt      *time.Time    `json:"answered_at"`
	DurationSeconds int           `json:"duration_seconds"` // até agora
	TalkSeconds     int           `json:"talk_seconds"`     // desde o atendimento, até agora
	Recording       CallRecording `json:"recording"`
}

// Call é a ligação encerrada (lista, detalhe e webhooks).
type Call struct {
	ID       string  `json:"id"`
	Protocol *string `json:"protocol"`
	// Valores: inbound, outbound, internal.
	Direction *string  `json:"direction"`
	From      CallFrom `json:"from"`
	To        CallTo   `json:"to"`
	DID       *string  `json:"did"`   // só na entrada
	Trunk     *string  `json:"trunk"` // nome cadastrado; null na interna
	// Valores: extension, queue, ivr, trunk, passthrough, whatsapp.
	Route     *string    `json:"route"`
	Extension *string    `json:"extension"` // ramal que atendeu ou tocou (entrada e interna); null na saída
	Queue     *CallQueue `json:"queue"`
	Agent     *CallAgent `json:"agent"`
	// Valores: answered, no_answer, busy, failed, abandoned, ivr_ended, voicemail, unavailable, blocked.
	Result          *string       `json:"result"`
	HangupCause     *string       `json:"hangup_cause"`
	Abandoned       bool          `json:"abandoned"`
	AbandonReason   *string       `json:"abandon_reason"`
	StartedAt       time.Time     `json:"started_at"`
	AnsweredAt      *time.Time    `json:"answered_at"`
	EndedAt         *time.Time    `json:"ended_at"`
	DurationSeconds int           `json:"duration_seconds"`
	TalkSeconds     int           `json:"talk_seconds"`
	Recording       CallRecording `json:"recording"`
}

// CallAgent é o ramal que atendeu.
type CallAgent struct {
	Extension  string     `json:"extension"`
	Name       *string    `json:"name"`
	AnsweredAt *time.Time `json:"answered_at"`
}

// CallDetail é a ligação com as ofertas da fila, a linha do tempo e o estado da gravação.
// Os campos do modelo (Call) vêm primeiro; recording troca o {enabled} pelo bloco completo.
type CallDetail struct {
	ID       string  `json:"id"`
	Protocol *string `json:"protocol"`
	// Valores: inbound, outbound, internal.
	Direction *string  `json:"direction"`
	From      CallFrom `json:"from"`
	To        CallTo   `json:"to"`
	DID       *string  `json:"did"`
	Trunk     *string  `json:"trunk"`
	// Valores: extension, queue, ivr, trunk, passthrough, whatsapp.
	Route     *string    `json:"route"`
	Extension *string    `json:"extension"`
	Queue     *CallQueue `json:"queue"`
	Agent     *CallAgent `json:"agent"`
	// Valores: answered, no_answer, busy, failed, abandoned, ivr_ended, voicemail, unavailable, blocked.
	Result          *string             `json:"result"`
	HangupCause     *string             `json:"hangup_cause"`
	Abandoned       bool                `json:"abandoned"`
	AbandonReason   *string             `json:"abandon_reason"`
	StartedAt       time.Time           `json:"started_at"`
	AnsweredAt      *time.Time          `json:"answered_at"`
	EndedAt         *time.Time          `json:"ended_at"`
	DurationSeconds int                 `json:"duration_seconds"`
	TalkSeconds     int                 `json:"talk_seconds"`
	Recording       CallRecordingDetail `json:"recording"`
	Offers          []CallOffer         `json:"offers"`
	Timeline        []CallTimelineEvent `json:"timeline"`
}

// CallFrom é quem ligou.
type CallFrom struct {
	Number *string `json:"number"`
	Name   *string `json:"name"`
}

// CallOffer é uma tentativa da fila de entregar a ligação a um ramal (queue_offers).
//
// result: answered (o ramal atendeu — CONNECT no queue_log), no_answer (tocou e não atendeu —
// RINGNOANSWER), canceled (o chamador desligou enquanto este ramal tocava — RINGCANCELED; quando
// outro ramal atende, o Asterisk não grava evento para os demais). offered_at é quando o ramal
// começou a tocar (nas ligações antigas, sem started_at, o fim do toque).
type CallOffer struct {
	Extension string     `json:"extension"`
	Name      *string    `json:"name"`
	OfferedAt *time.Time `json:"offered_at"`
	// Valores: answered, no_answer, canceled.
	Result string `json:"result"`
}

// CallQueue é a fila por onde a ligação passou.
type CallQueue struct {
	Name        string `json:"name"` // nome de exibição
	WaitSeconds *int   `json:"wait_seconds"`
}

// CallRecording diz se a ligação foi gravada.
type CallRecording struct {
	Enabled bool `json:"enabled"`
}

// CallRecordingDetail é o estado da gravação da ligação.
type CallRecordingDetail struct {
	Enabled   bool `json:"enabled"`   // a ligação foi gravada
	Available bool `json:"available"` // dá para gerar o link agora (recording-link)
	Archived  bool `json:"archived"`  // removida pela retenção do armazenamento; não volta
}

// CallRequest é o corpo do click to call (os mesmos campos valem na query string).
type CallRequest struct {
	Extension   string `json:"extension"`
	Destination string `json:"destination"`
}

// CallRequested é a resposta do click to call. O nome no Swagger não pode ser "Call", que é a
// ligação (LigacaoV2): com dois tipos "Call", o swag documentava um no lugar do outro.
type CallRequested struct {
	Extension   string `json:"extension"`
	Destination string `json:"destination"` // os dígitos que serão discados
	// Valores: dialing_extension.
	Status      string    `json:"status"`
	RequestedAt time.Time `json:"requested_at"` // no fuso do cliente
}

// CallTimelineEvent é um passo do caminho da ligação (call_path). details só traz o que o
// cliente reconhece (números de ramal, nomes cadastrados); o que é interno do PABX fica fora.
type CallTimelineEvent struct {
	Event   string            `json:"event"`
	At      *time.Time        `json:"at"`
	Details map[string]string `json:"details,omitempty"`
}

// CallTo é o número chamado (o discado, ou o DID/número de entrada).
type CallTo struct {
	Number *string `json:"number"`
}

// Campaign é a campanha na lista da v2.
type Campaign struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// pending (importando os contatos), processing, ready (pronta para iniciar), scheduled, running, paused, completed, canceled ou failed.
	// Valores: pending, processing, ready, scheduled, running, paused, completed, canceled, failed.
	Status     string         `json:"status"`
	InstanceID *string        `json:"instance_id"`
	TemplateID *string        `json:"template_id"`
	AutoStart  bool           `json:"auto_start"`
	StartAt    *time.Time     `json:"start_at"`
	EndAt      *time.Time     `json:"end_at"`
	Counts     CampaignCounts `json:"counts"`
	Reply      CampaignReply  `json:"reply"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// CampaignContact é o resultado do envio a um contato da campanha.
type CampaignContact struct {
	Contact CampaignContactRef `json:"contact"`
	// pending (ainda não enviado), sent, delivered, read, replied, failed ou skipped (não enviado: ex. número repetido ou sem WhatsApp).
	// Valores: pending, sent, delivered, read, replied, failed, skipped.
	Status      string     `json:"status"`
	SentAt      *time.Time `json:"sent_at"`
	DeliveredAt *time.Time `json:"delivered_at"`
	ReadAt      *time.Time `json:"read_at"`
	RepliedAt   *time.Time `json:"replied_at"`
	FailedAt    *time.Time `json:"failed_at"`
	// Motivo da falha, como o provedor devolveu.
	Error *string `json:"error"`
	// Botão do template que o contato tocou, quando respondeu por botão.
	ReplyButton *string `json:"reply_button"`
	// Conversa aberta pela resposta do contato (GET /conversations/{protocol}). null sem resposta.
	Conversation *CampaignContactConversation `json:"conversation"`
}

// CampaignContactConversation é a conversa aberta pela resposta.
type CampaignContactConversation struct {
	Protocol string `json:"protocol"`
}

// CampaignContactRef é o contato do envio.
type CampaignContactRef struct {
	// id do contato (GET /contacts/{id}).
	ID   *string `json:"id"`
	Name *string `json:"name"`
	// Número para o qual a campanha enviou.
	Address string `json:"address"`
}

// CampaignCounts são os contadores da campanha.
type CampaignCounts struct {
	Total     int `json:"total"`
	Sent      int `json:"sent"`
	Delivered int `json:"delivered"`
	Read      int `json:"read"`
	Replied   int `json:"replied"`
	Failed    int `json:"failed"`
}

// CampaignCreateRequest é a criação: JSON (contatos de grupos) ou a parte `campaign` do multipart
// (contatos do CSV na parte `file`).
type CampaignCreateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// id da instância de WhatsApp (GET /instances).
	InstanceID string                        `json:"instance_id"`
	Template   CampaignCreateRequestTemplate `json:"template"`
	// Início do disparo, ISO 8601 com fuso. Com auto_start, a campanha começa sozinha nessa hora.
	StartAt *string `json:"start_at"`
	// Limite do disparo, depois de start_at. Vazio = até concluir.
	EndAt     *string `json:"end_at"`
	AutoStart *bool   `json:"auto_start"`
	// Grupos de contatos (ids de grupo, os de GET /contacts), quando os contatos não vêm de CSV. Até 20.
	GroupIDs []int32 `json:"group_ids"`
	// Com CSV, obrigatório: o que fazer com o contato que já existe — ignore (mantém), overwrite (sobrescreve nome, e-mail e empresa) ou update_empty (preenche só o vazio).
	// Valores: ignore, overwrite, update_empty.
	CollisionPolicy string            `json:"collision_policy"`
	Reply           *CampaignReply    `json:"reply"`
	Settings        *CampaignSettings `json:"settings"`
}

// CampaignCreateRequestTemplate é o template da campanha.
type CampaignCreateRequestTemplate struct {
	// id do template (GET /templates), aprovado, da instância.
	ID string `json:"id"`
	// Valores das variáveis, iguais para todos os contatos. Sem eles, os valores vêm dos dados do contato (colunas do CSV) pelo mapeamento do template.
	Params []string `json:"params"`
}

// CampaignDetail é o detalhe: os campos da lista, a importação dos contatos e o envio.
type CampaignDetail struct {
	Campaign
	// Importação dos contatos. null antes de começar.
	Import   *CampaignImport  `json:"import"`
	Settings CampaignSettings `json:"settings"`
}

// CampaignDispatchPeriod é um período de disparo.
type CampaignDispatchPeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// CampaignImport é o processamento dos contatos.
type CampaignImport struct {
	// Nome do CSV, ou "Grupos de contatos: …" quando os contatos vieram de grupos.
	FileName      string  `json:"file_name"`
	Status        string  `json:"status"`
	TotalRows     int     `json:"total_rows"`
	ProcessedRows int     `json:"processed_rows"`
	FailedRows    int     `json:"failed_rows"`
	Error         *string `json:"error"`
}

// CampaignReply é como a resposta do contato é contada.
type CampaignReply struct {
	// true: a mensagem do contato sem citar o template também conta como resposta, dentro de window_hours.
	WithoutContext bool `json:"without_context"`
	WindowHours    int  `json:"window_hours"`
}

// CampaignSettings são as configurações de disparo.
type CampaignSettings struct {
	// Intervalo entre envios, em ms (1 a 2000). null = padrão (500).
	DelayMs *int `json:"delay_ms"`
	// Novas tentativas para envio com falha (0 a 10).
	MaxRetries *int `json:"max_retries"`
	// Espera antes de nova tentativa, em segundos (1 a 86400).
	RetryDelaySeconds *int `json:"retry_delay_seconds"`
	// Dias permitidos, de 0 (domingo) a 6 (sábado), no fuso do cliente. Vazio = todos.
	DispatchDays []int `json:"dispatch_days"`
	// Horários permitidos no fuso do cliente (até 3). Vazio = dia todo.
	DispatchPeriods []CampaignDispatchPeriod `json:"dispatch_periods"`
}

// Contact é um contato da central do cliente.
type Contact struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	CPF         *string           `json:"cpf"` // só dígitos (cadastros antigos pela tela podem ter pontuação)
	Email       *string           `json:"email"`
	Company     *string           `json:"company"`
	CustomInfo  map[string]any    `json:"custom_info"`  // campos personalizados, pela key do campo
	ReplaceName bool              `json:"replace_name"` // o nome pode ser trocado pelo nome de perfil do canal
	Identities  []ContactIdentity `json:"identities"`
	Labels      []ContactLabel    `json:"labels"`
	Groups      []ContactGroup    `json:"groups"`
	CreatedAt   time.Time         `json:"created_at"` // no fuso do cliente
	UpdatedAt   time.Time         `json:"updated_at"` // no fuso do cliente
}

// ContactBatchItemResult é o resultado de um contato do lote.
type ContactBatchItemResult struct {
	Index int `json:"index"` // posição no contacts enviado (base 0)
	// Valores: created, existing, updated, error.
	Status    string  `json:"status"`
	ContactID *string `json:"contact_id"` // null quando o item falhou
	Message   *string `json:"message"`    // motivo do erro ou aviso
	Field     *string `json:"field"`      // campo recusado, no item com erro de validação
}

// ContactBatchRequest documenta o corpo do POST /contacts/batch.
type ContactBatchRequest struct {
	// CollisionPolicy: o que fazer quando uma identidade do item já está em outro contato.
	// ignore (padrão) mantém os dados do contato (etiquetas e grupos são acrescentados);
	// overwrite sobrescreve os campos enviados e mescla custom_info por chave; update_empty
	// preenche só os campos vazios. As duas últimas também acrescentam as identidades novas.
	// Valores: ignore, overwrite, update_empty.
	CollisionPolicy string          `json:"collision_policy,omitempty"`
	Contacts        []ContactCreate `json:"contacts"` // 1 a 500
}

// ContactBatchResult é o resultado do lote.
type ContactBatchResult struct {
	Total    int                      `json:"total"`
	Created  int                      `json:"created"`
	Updated  int                      `json:"updated"`
	Existing int                      `json:"existing"`
	Errors   int                      `json:"errors"`
	Items    []ContactBatchItemResult `json:"items"`
}

// ContactCreate documenta o corpo do POST (e de cada item do lote).
type ContactCreate struct {
	Name        string                 `json:"name"`                   // obrigatório, até 255 caracteres
	CPF         *string                `json:"cpf,omitempty"`          // com ou sem pontuação; gravado só com os dígitos
	Email       *string                `json:"email,omitempty"`        // até 255 caracteres
	Company     *string                `json:"company,omitempty"`      // até 255 caracteres
	CustomInfo  map[string]string      `json:"custom_info,omitempty"`  // pela key do campo personalizado cadastrado (até 10 KB)
	ReplaceName *bool                  `json:"replace_name,omitempty"` // o nome pode ser trocado pelo nome de perfil do canal
	Identities  []ContactIdentityInput `json:"identities"`             // ao menos uma, até 10
	Labels      []int64                `json:"labels,omitempty"`       // números das etiquetas (até 20)
	Groups      []int32                `json:"groups,omitempty"`       // ids dos grupos de contatos (até 20)
}

// ContactGroup é um grupo de contatos (restringe quem vê o contato no painel).
type ContactGroup struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

// ContactIdentity é uma forma de contato (telefone, WhatsApp, username...).
type ContactIdentity struct {
	// Valores: phone, whatsapp, telegram, instagram, facebook, sip.
	Channel string `json:"channel"` // contatos antigos podem ter sms
	// Valores: phone, username, chat_id, handle, extension, profile_id.
	IDType    string `json:"id_type"`
	IDValue   string `json:"id_value"`
	IsPrimary bool   `json:"is_primary"`
}

// ContactIdentityInput é uma identidade do corpo. Canal e tipo aceitos: phone→phone;
// whatsapp→phone|username; telegram→chat_id|username; instagram e facebook→handle|profile_id;
// sip→extension.
type ContactIdentityInput struct {
	// Valores: phone, whatsapp, telegram, instagram, facebook, sip.
	Channel string `json:"channel"`
	// Valores: phone, username, chat_id, handle, extension, profile_id.
	IDType  string `json:"id_type"`
	IDValue string `json:"id_value"`
	// Principal do canal: uma por canal. O canal sem nenhuma marcada fica com a primeira identidade dele.
	IsPrimary bool `json:"is_primary,omitempty"`
}

// ContactLabel é uma etiqueta do contato, pelo número (o mesmo da gestão de etiquetas).
type ContactLabel struct {
	Number int64  `json:"number"`
	Name   string `json:"name"`
}

// ContactUpdate documenta o corpo do PATCH: campo ausente fica como está.
type ContactUpdate struct {
	Name        Opt[string]                 `json:"name,omitzero"`
	CPF         Opt[string]                 `json:"cpf,omitzero"`         // null ou "" limpa
	Email       Opt[string]                 `json:"email,omitzero"`       // null ou "" limpa
	Company     Opt[string]                 `json:"company,omitzero"`     // null ou "" limpa
	CustomInfo  Opt[map[string]string]      `json:"custom_info,omitzero"` // substitui o objeto inteiro; null limpa
	ReplaceName Opt[bool]                   `json:"replace_name,omitzero"`
	Identities  Opt[[]ContactIdentityInput] `json:"identities,omitzero"` // substitui a lista (ao menos uma)
	Labels      Opt[[]int64]                `json:"labels,omitzero"`     // substitui a lista; [] ou null tira todas
	Groups      Opt[[]int32]                `json:"groups,omitzero"`     // substitui a lista; [] ou null tira de todos
}

// Conversation é uma conversa na lista da v2.
type Conversation struct {
	Protocol string `json:"protocol"`
	// Valores: bot, ai_agent, queue, attending, coex, finished.
	Status      string `json:"status"`
	ChannelType string `json:"channel_type"`
	// Endereço do canal da empresa por onde a conversa passou: o número do WhatsApp, o id da página ou do perfil. Vem mesmo com o canal excluído.
	ChannelAddress *string `json:"channel_address"`
	// id da instância ativa com esse endereço (GET /instances). null no webchat e quando o canal foi excluído e não recadastrado; com o número recadastrado, é o id da instância nova.
	InstanceID *string `json:"instance_id"`
	// Valores: inbound, outbound.
	Direction *string             `json:"direction"`
	Contact   ConversationContact `json:"contact"`
	// Fila da conversa (id de GET /queues). null quando não passou por fila.
	Queue *ConversationReference `json:"queue"`
	// Atendente da conversa (id de GET /users). null quando não foi atendida por pessoa.
	Agent         *ConversationReference `json:"agent"`
	LastMessageAt *time.Time             `json:"last_message_at"`
	CreatedAt     time.Time              `json:"created_at"`
	QueuedAt      *time.Time             `json:"queued_at"`
	AssignedAt    *time.Time             `json:"assigned_at"`
	FinishedAt    *time.Time             `json:"finished_at"`
}

// ConversationCloseRequest é o corpo opcional do close.
type ConversationCloseRequest struct {
	// Classificação do atendimento, uma das opções da fila. Obrigatória se a fila exigir.
	Classification *string `json:"classification"`
}

// ConversationContact é o contato da conversa.
type ConversationContact struct {
	// id do contato (GET /contacts/{id}). null nas raras conversas sem contato.
	ID   *string `json:"id"`
	Name *string `json:"name"`
	// Endereço do contato no canal da conversa: o número no WhatsApp, o id do perfil para a página no Instagram e no Facebook, o chat_id no Telegram, a sessão do visitante no webchat.
	Address *string `json:"address"`
}

// ConversationDetail é o detalhe da conversa (GET .../conversations/{protocol}): os campos da
// lista e os do atendimento.
type ConversationDetail struct {
	Conversation
	FirstReplyAt *time.Time `json:"first_reply_at"`
	// Quem ou o que encerrou: agent, bot, timeout, monitor, api_token, contact, error, coex_inactivity, coex_history_import, coex_phrase. null em andamento.
	FinishReason *string                  `json:"finish_reason"`
	ClosedBy     *ConversationReference   `json:"closed_by"`
	LastMessage  *ConversationLastMessage `json:"last_message"`
	Tags         []string                 `json:"tags"` // etiquetas de atendimento, pelo nome
}

// ConversationLastMessage é a última mensagem da conversa.
type ConversationLastMessage struct {
	Content   *string `json:"content"`
	Type      *string `json:"type"`
	Direction *string `json:"direction"`
}

// ConversationReference é uma fila ou um usuário da conversa.
type ConversationReference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ConversationStartRequest é o corpo do POST /conversations.
type ConversationStartRequest struct {
	// id da instância de WhatsApp (GET /instances).
	InstanceID string `json:"instance_id"`
	// Número de destino, com DDI.
	To string `json:"to"`
	// Nome do contato, usado se o número ainda não tem contato cadastrado.
	ContactName string `json:"contact_name"`
	// Usuário que fica com a conversa (GET /users), ativo e com acesso ao canal.
	UserID string `json:"user_id"`
	// Fila de chat da conversa (GET /queues), opcional.
	QueueID  string                           `json:"queue_id"`
	Template ConversationStartRequestTemplate `json:"template"`
}

// ConversationStartRequestMedia é a mídia do header do template.
type ConversationStartRequestMedia struct {
	// Link http(s) do arquivo. Arquivo do armazenamento da Wonit só das conversas ou de mídia de template (outros arquivos, como gravações e exportações, são recusados) e vai ao provedor por um link que vence em 24 horas.
	URL string `json:"url"`
	// Valores: image, video, document.
	Type     string `json:"type"`
	Filename string `json:"filename"`
}

// ConversationStartRequestTemplate é o template da primeira mensagem.
type ConversationStartRequestTemplate struct {
	// id do template (GET /templates).
	ID string `json:"id"`
	// Valores das variáveis {{1}}, {{2}}…; sem valor, valem os mapeamentos do template, quando houver.
	Params []string `json:"params"`
	// Mídia do header, só em template de imagem, vídeo ou documento.
	Media *ConversationStartRequestMedia `json:"media"`
}

// ConversationStarted é o resultado do início de conversa.
type ConversationStarted struct {
	Protocol     string             `json:"protocol"`
	Conversation ConversationDetail `json:"conversation"`
	Message      MessageDetail      `json:"message"`
}

// ConversationTransferRequest é o corpo do transfer: queue_id, user_id ou os dois.
type ConversationTransferRequest struct {
	// Fila de chat de destino (GET /queues). Sozinha, a conversa vai para a fila; com user_id, troca a fila e atribui ao usuário.
	QueueID *string `json:"queue_id"`
	// Usuário de destino (GET /users). Sem queue_id, a conversa fica na fila atual.
	UserID *string `json:"user_id"`
}

// ErrorCode é o code estável do erro na v2, em inglês SCREAMING_SNAKE. É contrato: o cliente
// decide por ele, não pela mensagem. Código novo entra aqui, em Codigos e no doc.go de
// internal/publicapi/v2 (TestCodigosDocumentados confere).
type ErrorCode string

// Extension é o item de ramal da v2 (lista, detalhe e respostas dos PUT). Nunca leva senha SIP,
// contexto, id interno do endpoint nem domínio.
type Extension struct {
	Number   string  `json:"number"`
	Name     string  `json:"name"`
	CallerID *string `json:"caller_id"` // só o número do callerid
	// Valores: offline, in_call, dnd, online.
	Status     string              `json:"status"`
	Registered bool                `json:"registered"` // algum aparelho com registro SIP válido
	InCall     bool                `json:"in_call"`
	DND        bool                `json:"dnd"`
	Forwarding ExtensionForwarding `json:"forwarding"`
	User       *ExtensionUser      `json:"user"` // usuário vinculado ao ramal; null quando não há
}

// ExtensionCapabilities liberados no ramal. dnd, forward_internal e forward_external são os mesmos que os
// PUT de dnd e forwarding conferem (PSEndpoint.PermiteDnd / PermiteDesvioInterno / Externo).
type ExtensionCapabilities struct {
	DND             bool  `json:"dnd"`              // pode ligar o Não perturbe
	ForwardInternal bool  `json:"forward_internal"` // pode desviar para ramal do cliente
	ForwardExternal bool  `json:"forward_external"` // pode desviar para número externo
	SpyMonitor      *bool `json:"spy_monitor"`      // pode monitorar (espionar) outros ramais
	SpyMonitored    *bool `json:"spy_monitored"`    // pode ser monitorado
	Subscribe       *bool `json:"subscribe"`        // aceita SUBSCRIBE (BLF/presença)
}

// PedidoRecursos: capabilities do PATCH/POST.
type ExtensionCapabilitiesUpdate struct {
	DND             Opt[bool] `json:"dnd,omitzero"`
	ForwardInternal Opt[bool] `json:"forward_internal,omitzero"`
	ForwardExternal Opt[bool] `json:"forward_external,omitzero"`
	SpyMonitor      Opt[bool] `json:"spy_monitor,omitzero"`
	SpyMonitored    Opt[bool] `json:"spy_monitored,omitzero"`
	Subscribe       Opt[bool] `json:"subscribe,omitzero"`
}

// ExtensionCreate documenta o corpo do POST: number e os blocos do PATCH.
type ExtensionCreate struct {
	Number string `json:"number"` // 2 a 4 dígitos, único no cliente
	ExtensionUpdate
}

// ExtensionCreated é a resposta de POST /extensions.
type ExtensionCreated struct {
	Extension      ExtensionDetail `json:"extension"`
	SIPCredentials SipCredentials  `json:"sip_credentials"`
}

// ExtensionDetail é o detalhe do ramal (GET .../extensions/{number} e respostas dos PUT): os
// campos da lista e as configurações. Nunca leva senha/auth SIP, outbound_auth, aors, contexto,
// outbound_proxy, id interno, domínio, nem os aparelhos registrados (contacts, user agent).
type ExtensionDetail struct {
	Extension
	Capabilities ExtensionCapabilities `json:"capabilities"`
	Dialing      ExtensionDialing      `json:"dialing"`
	// Valores: always, never.
	Recording    *string                `json:"recording"` // null: não configurado
	Voicemail    ExtensionVoicemail     `json:"voicemail"`
	PickupGroups []ExtensionPickupGroup `json:"pickup_groups"`
	// Filas de voz em que o ramal é membro estático (fixo, sem login), em ordem de nome. As
	// filas dinâmicas são do usuário: GET /users/{id}, voice.queues.
	Queues []ExtensionQueue `json:"queues"`
	SIP    ExtensionSip     `json:"sip"`
}

// ExtensionDialCategory é a categoria de discagem do ramal.
type ExtensionDialCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Discagem: tipos de ligação externa liberados e a categoria de discagem.
type ExtensionDialing struct {
	Unrestricted bool                   `json:"unrestricted"` // true: disca qualquer tipo (allowed vazio)
	Allowed      []string               `json:"allowed"`
	Category     *ExtensionDialCategory `json:"category"`
}

// PedidoDiscagem: dialing do PATCH/POST.
type ExtensionDialingUpdate struct {
	Allowed    Opt[[]string] `json:"allowed,omitzero"`
	CategoryID Opt[int]      `json:"category_id,omitzero"`
}

// ExtensionDndRequest é o corpo de PUT .../dnd.
type ExtensionDndRequest struct {
	Enabled *bool `json:"enabled"`
}

// ExtensionForwarding do ramal; null quando o desvio está desligado.
type ExtensionForwarding struct {
	Always   *string `json:"always"`
	NoAnswer *string `json:"no_answer"`
	Busy     *string `json:"busy"`
}

// ExtensionForwardingRequest é o corpo de PUT .../forwarding: campo ausente fica como está, null desliga.
type ExtensionForwardingRequest struct {
	Always   Opt[string] `json:"always,omitzero"`
	NoAnswer Opt[string] `json:"no_answer,omitzero"`
	Busy     Opt[string] `json:"busy,omitzero"`
}

// ExtensionPasswordResponse é a resposta de POST /extensions/{number}/password.
type ExtensionPasswordResponse struct {
	SIPCredentials SipCredentials `json:"sip_credentials"`
}

// ExtensionPickupGroup é um grupo de captura em que o ramal está.
type ExtensionPickupGroup struct {
	Number    *int64 `json:"number"` // número do grupo na tela de grupos de captura
	Name      string `json:"name"`
	CanCall   bool   `json:"can_call"`   // as ligações dele podem ser capturadas pelo grupo
	CanPickup bool   `json:"can_pickup"` // captura as ligações do grupo
}

// ExtensionPickupGroupRequest é um grupo de captura pedido no corpo (pelo número do grupo na tela).
type ExtensionPickupGroupRequest struct {
	Number    int64 `json:"number"`
	CanCall   *bool `json:"can_call,omitempty"`
	CanPickup *bool `json:"can_pickup,omitempty"`
}

// ExtensionQueue é uma fila de voz em que o ramal é membro estático.
type ExtensionQueue struct {
	ID     string  `json:"id"` // id de GET /queues
	Name   string  `json:"name"`
	Number *string `json:"number"`
}

// ExtensionSessionTimers são os session timers (segundos).
type ExtensionSessionTimers struct {
	Expires *int32 `json:"expires"`
	MinSe   *int32 `json:"min_se"`
}

// PedidoTimersDeSessao: sip.session_timers do PATCH/POST.
type ExtensionSessionTimersUpdate struct {
	Expires Opt[int] `json:"expires,omitzero"`
	MinSe   Opt[int] `json:"min_se,omitzero"`
}

// ExtensionSip são os parâmetros SIP do ramal (sem credenciais).
type ExtensionSip struct {
	Transport      *string                `json:"transport"`
	Codecs         []string               `json:"codecs"`
	DTMFMode       *string                `json:"dtmf_mode"`
	DirectMedia    *bool                  `json:"direct_media"`
	IceSupport     *bool                  `json:"ice_support"`
	ForceRport     *bool                  `json:"force_rport"`
	RewriteContact *bool                  `json:"rewrite_contact"`
	RtpSymmetric   *bool                  `json:"rtp_symmetric"`
	RtpTimeout     *int32                 `json:"rtp_timeout"`
	SessionTimers  ExtensionSessionTimers `json:"session_timers"`
}

// PedidoSip: sip do PATCH/POST.
type ExtensionSipUpdate struct {
	// Valores: udp, wss.
	Transport Opt[string]   `json:"transport,omitzero"`
	Codecs    Opt[[]string] `json:"codecs,omitzero"`
	// Valores: rfc4733, rfc2833, inband, info, auto.
	DTMFMode       Opt[string]                       `json:"dtmf_mode,omitzero"`
	DirectMedia    Opt[bool]                         `json:"direct_media,omitzero"`
	IceSupport     Opt[bool]                         `json:"ice_support,omitzero"`
	ForceRport     Opt[bool]                         `json:"force_rport,omitzero"`
	RewriteContact Opt[bool]                         `json:"rewrite_contact,omitzero"`
	RtpSymmetric   Opt[bool]                         `json:"rtp_symmetric,omitzero"`
	RtpTimeout     Opt[int]                          `json:"rtp_timeout,omitzero"`
	SessionTimers  Opt[ExtensionSessionTimersUpdate] `json:"session_timers,omitzero"`
}

// ExtensionUpdate documenta o corpo do PATCH (os blocos do detalhe).
type ExtensionUpdate struct {
	Name         Opt[string]                      `json:"name,omitzero"`
	CallerID     Opt[string]                      `json:"caller_id,omitzero"`
	Capabilities Opt[ExtensionCapabilitiesUpdate] `json:"capabilities,omitzero"`
	Dialing      Opt[ExtensionDialingUpdate]      `json:"dialing,omitzero"`
	// Valores: always, never.
	Recording    Opt[string]                        `json:"recording,omitzero"`
	Voicemail    Opt[ExtensionVoicemailUpdate]      `json:"voicemail,omitzero"`
	PickupGroups Opt[[]ExtensionPickupGroupRequest] `json:"pickup_groups,omitzero"`
	SIP          Opt[ExtensionSipUpdate]            `json:"sip,omitzero"`
}

// ExtensionUser é o usuário da plataforma vinculado ao ramal.
type ExtensionUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ExtensionVoicemail do ramal.
type ExtensionVoicemail struct {
	Extension        *string  `json:"extension"`
	Emails           []string `json:"emails"`
	MissedCallNotice bool     `json:"missed_call_notice"` // aviso por e-mail de ligação perdida (não atendida ou ocupado)
}

// PedidoCaixaPostal: voicemail do PATCH/POST.
type ExtensionVoicemailUpdate struct {
	Extension        Opt[string]   `json:"extension,omitzero"`
	Emails           Opt[[]string] `json:"emails,omitzero"`
	MissedCallNotice Opt[bool]     `json:"missed_call_notice,omitzero"`
}

// FieldError é um item de errors: o campo (caminho no corpo ou nome do parâmetro) e o
// motivo. Sem field quando a recusa é do corpo inteiro.
type FieldError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// Identification é a resposta de GET /api/public/v2. Nunca leva o token nem o hash.
type Identification struct {
	Name          string                     `json:"name"`
	Version       string                     `json:"version"`
	Documentation string                     `json:"documentation"`
	Tenant        IdentificationTenant       `json:"tenant"`
	Token         IdentificationToken        `json:"token"`
	Modules       []string                   `json:"modules"`
	RateLimit     IdentificationRateLimit    `json:"rate_limit"`
	Permissions   []IdentificationPermission `json:"permissions"`
}

// IdentificationPermission é uma rota da v2 liberada no token e as ações liberadas nela.
type IdentificationPermission struct {
	Route   string   `json:"route"`
	Actions []Action `json:"actions"`
}

// IdentificationRateLimit é a cota em vigor para o cliente: uma só para todas as rotas da v2 (somando
// todos os tokens dele e a cota padrão da v1), recomposta continuamente ao longo da janela.
type IdentificationRateLimit struct {
	Limit         int `json:"limit"`          // tamanho da cota (requisições)
	WindowSeconds int `json:"window_seconds"` // tempo para a cota se recompor inteira
}

// IdentificationTenant é o cliente dono do token, só pelo domínio (tenants.domain).
type IdentificationTenant struct {
	// Endereço do cliente na plataforma (o host do base_url): serve para confirmar de qual cliente é o token.
	Domain string `json:"domain"`
}

// IdentificationToken identifica o token sem o valor: prefix é o mesmo prefixo da tela de tokens.
type IdentificationToken struct {
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	ExpiresAt time.Time `json:"expires_at"` // no fuso do cliente
}

// Instance é uma instância de canal na v2.
type Instance struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Source string `json:"source"`
	// Valores: whatsapp, instagram, facebook, telegram.
	ChannelType string `json:"channel_type"`
	Active      bool   `json:"active"`
	Receptive   bool   `json:"receptive"`
}

// Message é uma mensagem da conversa no histórico da v2.
type Message struct {
	// id da mensagem, sempre dentro da conversa (GET /conversations/{protocol}/messages/{id}).
	ID int64 `json:"id"`
	// Valores: inbound, outbound.
	Direction string `json:"direction"`
	// Quem enviou: client (o contato), agent (atendente), bot (chatbot), ai_agent (agente de IA) ou agent_coex (resposta pelo app WhatsApp Business no celular).
	// Valores: client, agent, bot, ai_agent, agent_coex.
	SentBy string `json:"sent_by"`
	// O usuário, quando foi uma pessoa (GET /users).
	Agent *ConversationReference `json:"agent"`
	// text, image, audio, video, file, sticker, template, interactive, location, contact, reaction (só quando a mensagem reagida foi apagada) ou unsupported.
	Type    string  `json:"type"`
	Content *string `json:"content"`
	Caption *string `json:"caption"`
	// id da mensagem citada (a resposta a uma mensagem anterior). Pode apontar para uma mensagem apagada, que não está no histórico. null sem citação.
	ReplyToID *int64 `json:"reply_to_id"`
	// Mídia da mensagem; null sem mídia.
	Media *MessageMedia `json:"media"`
	// pending, sent, delivered, read ou failed.
	Status string `json:"status"`
	// O remetente editou o conteúdo depois do envio.
	Edited bool `json:"edited"`
	// O remetente apagou a mensagem para todos; o conteúdo original continua em content.
	RemovedAt *time.Time `json:"removed_at"`
	// Reações à mensagem: a última de cada remetente, em ordem de chegada.
	Reactions []MessageReaction `json:"reactions"`
	CreatedAt time.Time         `json:"created_at"`
}

// MessageDetail é o detalhe da mensagem: os campos do histórico e os da entrega.
type MessageDetail struct {
	Message
	SentAt      *time.Time `json:"sent_at"`
	DeliveredAt *time.Time `json:"delivered_at"`
	ReadAt      *time.Time `json:"read_at"`
	FailedAt    *time.Time `json:"failed_at"`
	// Erro do envio, como o provedor devolveu; null sem erro.
	Error *MessageError `json:"error"`
	// id da mensagem no provedor (o wamid no WhatsApp), para cruzar com os webhooks dele.
	ProviderID *string `json:"provider_id"`
}

// MessageError é o erro do envio de uma mensagem.
type MessageError struct {
	Code    *string `json:"code"`
	Details *string `json:"details"`
}

// MessageFileLink é o link temporário do arquivo.
type MessageFileLink struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// MessageFileLinkRequest é o corpo opcional do file-link.
type MessageFileLinkRequest struct {
	ExpiresIn *int `json:"expires_in"`
}

// MessageMedia é a mídia da mensagem.
type MessageMedia struct {
	Mime     *string `json:"mime"`
	FileName *string `json:"file_name"`
	// true: o arquivo está guardado; baixe pelo file-link.
	Downloadable bool `json:"downloadable"`
	// Só quando a mídia foi enviada por link externo: o link original.
	URL *string `json:"url"`
}

// MessageReaction é uma reação a uma mensagem.
type MessageReaction struct {
	Emoji  string `json:"emoji"`
	SentBy string `json:"sent_by"`
}

// MessageSendRequest é o corpo do POST /conversations/{protocol}/messages.
type MessageSendRequest struct {
	// text, image, video, file, audio, sticker ou location.
	// Valores: text, image, video, file, audio, sticker, location.
	Type string `json:"type"`
	// Texto, em type text.
	Content string `json:"content"`
	// Legenda, em image, video e file.
	Caption string `json:"caption"`
	// Mídia, em image, video, file, audio e sticker.
	Media *MessageSendRequestMedia `json:"media"`
	// Localização, em type location.
	Location *MessageSendRequestLocation `json:"location"`
}

// MessageSendRequestLocation é a localização enviada.
type MessageSendRequestLocation struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Name      string   `json:"name"`
	Address   string   `json:"address"`
}

// MessageSendRequestMedia é a mídia enviada.
type MessageSendRequestMedia struct {
	// Link http(s) do arquivo. Arquivo do armazenamento da Wonit só das conversas ou de mídia de template (outros arquivos, como gravações e exportações, são recusados) e vai ao provedor por um link que vence em 24 horas.
	URL string `json:"url"`
	// Nome do arquivo, em type file.
	Filename string `json:"filename"`
}

// MessageTemplateRequest é o corpo do POST /messages.
type MessageTemplateRequest struct {
	// id da instância de WhatsApp (GET /instances).
	InstanceID string `json:"instance_id"`
	// Número de destino, com DDI (o nono dígito é ajustado como no painel).
	To string `json:"to"`
	// Nome do contato, usado se o número ainda não tem contato cadastrado.
	ContactName string                         `json:"contact_name"`
	Template    MessageTemplateRequestTemplate `json:"template"`
}

// MessageTemplateRequestMedia é a mídia do header do template.
type MessageTemplateRequestMedia struct {
	// Link http(s) do arquivo. Arquivo do armazenamento da Wonit só das conversas ou de mídia de template (outros arquivos, como gravações e exportações, são recusados) e vai ao provedor por um link que vence em 24 horas.
	URL string `json:"url"`
	// image, video ou document; se vier, tem de ser o tipo do template.
	// Valores: image, video, document.
	Type string `json:"type"`
	// Nome do arquivo para documento.
	Filename string `json:"filename"`
}

// MessageTemplateRequestTemplate é o template do envio.
type MessageTemplateRequestTemplate struct {
	// id do template (GET /templates).
	ID string `json:"id"`
	// Valores das variáveis {{1}}, {{2}}… do corpo, em ordem. Toda variável precisa de valor.
	Params []string `json:"params"`
	// Mídia do header, só em template de imagem, vídeo ou documento. Sem ela, vale a mídia cadastrada no template.
	Media *MessageTemplateRequestMedia `json:"media"`
}

// MessageTemplateSent é o resultado: onde o template ficou e a mensagem.
type MessageTemplateSent struct {
	// Protocolo da conversa do envio: a conversa ativa do contato, se havia uma, ou a conversa criada para o envio (encerrada até o contato responder). Use para acompanhar o status: GET /conversations/{protocol}/messages/{id}.
	Protocol string        `json:"protocol"`
	Message  MessageDetail `json:"message"`
}

// Pagination é o bloco pagination das listas.
type Pagination struct {
	NextCursor *string `json:"next_cursor"` // null quando não há próxima página
	PageSize   int     `json:"page_size"`
	Total      int64   `json:"total"` // total de itens do filtro
}

// Problem é o corpo de erro da v2: RFC 9457 (Problem Details), Content-Type
// application/problem+json, sem "type" (vale about:blank: não há página por erro). code é o
// que a integração trata; title e detail são para pessoas e podem mudar. Os membros de
// extensão (action, module, retry_after_seconds, limit, current, user, token_version,
// route_version) vêm no primeiro nível; errors só no VALIDATION_ERROR, um item por campo.
type Problem struct {
	Title  string       `json:"title"`
	Status int          `json:"status"`
	Detail string       `json:"detail"`
	Code   ErrorCode    `json:"code"`
	Errors []FieldError `json:"errors,omitempty"`
	// Extras são os membros de extensão do código (fora do swagger: variam por código).
}

// Queue é uma fila de atendimento (chat ou voz) na v2.
type Queue struct {
	ID string `json:"id"`
	// Valores: chat, voice.
	Channel string `json:"channel"`
	Name    string `json:"name"`
	// Número da fila de voz (o que se disca ou se usa na URA). null na fila de chat.
	Number *string `json:"number"`
	// Estratégia de distribuição. Chat: ringall, roundrobin, random, linear, leastchats. Voz:
	// ringall, leastrecent, fewestcalls, random, rrmemory, rrordered, linear, wrandom.
	Strategy string `json:"strategy"`
	// A fila de voz não tem desativação: é sempre true.
	Active bool `json:"active"`
}

// RecordingLink é o link temporário da gravação.
type RecordingLink struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"` // no fuso do cliente
}

// RecordingLinkRequest é o corpo opcional do recording-link.
type RecordingLinkRequest struct {
	ExpiresIn *int `json:"expires_in"`
}

// SipCredentials é o que o aparelho precisa para registrar. A senha só aparece na resposta de
// criar e de gerar senha: não pode ser recuperada depois.
type SipCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain"` // o domínio SIP (registrar)
}

// Template é um template HSM na v2.
type Template struct {
	ID          string    `json:"id"`
	InstanceID  string    `json:"instance_id"`
	Name        string    `json:"name"`
	Language    string    `json:"language"`
	Category    string    `json:"category"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Content     string    `json:"content"`
	Header      *string   `json:"header"`
	Footer      *string   `json:"footer"`
	MediaURL    *string   `json:"media_url"`
	ParamsCount int       `json:"params_count"`
	CreatedAt   time.Time `json:"created_at"` // no fuso do cliente
	UpdatedAt   time.Time `json:"updated_at"` // no fuso do cliente
}

// User é um usuário do cliente na lista da v2: quem é e como está no chat e na voz.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// Número do ramal do usuário. null sem ramal ou sem a licença de telefonia.
	Extension *string `json:"extension"`
	// Estado nas filas de chat. null sem a licença de Multicanais.
	Chat *UserChannelState `json:"chat"`
	// Estado nas filas dinâmicas de voz. null sem a licença de telefonia.
	Voice *UserChannelState `json:"voice"`
}

// UserChannelState é o estado do usuário nas filas de um canal.
type UserChannelState struct {
	// online = logado em alguma fila do canal (no chat, também com a plataforma aberta, no
	// navegador ou no app); paused = online com pausa; offline = o resto.
	// Valores: online, paused, offline.
	Status      string     `json:"status"`
	PauseReason *string    `json:"pause_reason"`
	LoggedInAt  *time.Time `json:"logged_in_at"`
	PausedAt    *time.Time `json:"paused_at"`
}

// UserChannelStateDetail é o estado do usuário num canal com as filas dele.
type UserChannelStateDetail struct {
	UserChannelState
	// Filas dinâmicas do usuário no canal: as vinculadas a ele e as em que está logado, em
	// ordem de nome. As filas estáticas de voz são do ramal (GET /extensions/{number}).
	Queues []UserQueue `json:"queues"`
}

// UserDetail é o detalhe do usuário (GET .../users/{id}): os campos da lista e, em cada
// canal, as filas.
type UserDetail struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Extension *string `json:"extension"`
	// Estado nas filas de chat, com as filas. null sem a licença de Multicanais.
	Chat *UserChannelStateDetail `json:"chat"`
	// Estado nas filas dinâmicas de voz, com as filas. null sem a licença de telefonia.
	Voice *UserChannelStateDetail `json:"voice"`
}

// UserQueue é uma fila dinâmica do usuário.
type UserQueue struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LoggedIn bool   `json:"logged_in"`
}
