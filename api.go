package interage

// Constantes de path da API pública (Client API).
// Todos os paths são relativos à baseURL do tenant e prefixados por /api.
const (
	// ── Whatsapp.Campanhas ───────────────────────────────────────────────────
	pathCampaigns      = "/api/public/whatsapp/campanhas"
	pathCampaignByID   = "/api/public/whatsapp/campanhas/%s"
	pathCampaignStart  = "/api/public/whatsapp/campanhas/%s/start"
	pathCampaignPause  = "/api/public/whatsapp/campanhas/%s/pause"
	pathCampaignCancel = "/api/public/whatsapp/campanhas/%s/cancel"

	// ── Whatsapp.Mensagens ───────────────────────────────────────────────────
	pathInstances     = "/api/public/whatsapp/messages/instances"
	pathTemplates     = "/api/public/whatsapp/messages/templates"
	pathTemplateSend  = "/api/public/whatsapp/messages/templates/send"
	pathMessageSend   = "/api/public/whatsapp/messages/send"
	pathMessageStatus = "/api/public/whatsapp/messages/status"

	// ── Omni.Administrativo ──────────────────────────────────────────────────
	pathOmniQueues               = "/api/public/omni/administrativo/queues"
	pathOmniAgents               = "/api/public/omni/administrativo/agents"
	pathOmniConversations        = "/api/public/omni/administrativo/conversations"
	pathOmniConversationHistory  = "/api/public/omni/administrativo/conversations/%s/history"
	pathOmniConversationClose    = "/api/public/omni/administrativo/conversations/%s/close"
	pathOmniConversationTransfer = "/api/public/omni/administrativo/conversations/%s/transfer"
	pathOmniMessageFileTempLink  = "/api/public/omni/administrativo/conversations/messages/%s/templink"
	pathOmniContactsBatch        = "/api/public/omni/administrativo/contacts/batch"

	// ── Omni.Administrativo.Contatos ───────────────────────────────────────────
	pathContacts       = "/api/public/omni/administrativo/contacts"
	pathContactByID    = "/api/public/omni/administrativo/contacts/%s"
	pathContactByPhone = "/api/public/omni/administrativo/contacts/by-phone"

	// ── Pabx.Telefonia ───────────────────────────────────────────────────────
	pathExtensions        = "/api/public/pabx/telefonia/ramais"
	pathCallHistory       = "/api/public/pabx/telefonia/historico"
	pathOriginate         = "/api/public/pabx/telefonia/originate"
	pathActiveCalls       = "/api/public/pabx/telefonia/chamadas-ativas"
	pathRecordingTempLink = "/api/public/pabx/telefonia/historico/%s/recording/templink"
)
