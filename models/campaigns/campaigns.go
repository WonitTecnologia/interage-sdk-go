// Package campaigns contém os modelos do domínio de campanhas WhatsApp.
package campaigns

// CampaignStatus é o status de uma campanha.
type CampaignStatus string

const (
	CampaignStatusPending    CampaignStatus = "pending"    // criada; CSV em importação
	CampaignStatusProcessing CampaignStatus = "processing" // importando contatos
	CampaignStatusReady      CampaignStatus = "ready"      // pronta para iniciar
	CampaignStatusScheduled  CampaignStatus = "scheduled"  // agendada (auto_start futuro)
	CampaignStatusRunning    CampaignStatus = "running"    // em execução
	CampaignStatusPaused     CampaignStatus = "paused"     // pausada (retomável)
	CampaignStatusCompleted  CampaignStatus = "completed"  // concluída
	CampaignStatusCanceled   CampaignStatus = "canceled"   // cancelada (não reinicia)
	CampaignStatusFailed     CampaignStatus = "failed"     // falhou
)

// CollisionPolicy define o que fazer quando o contato do CSV já existe na base.
type CollisionPolicy string

const (
	// CollisionIgnore mantém os dados atuais do contato; não sobrescreve nada.
	CollisionIgnore CollisionPolicy = "ignore"
	// CollisionOverwrite sobrescreve nome/email/empresa com os valores do CSV.
	CollisionOverwrite CollisionPolicy = "overwrite"
	// CollisionUpdateEmpty preenche apenas os campos vazios do contato.
	CollisionUpdateEmpty CollisionPolicy = "update_empty"
)

// CreateCampaignRequest são os dados para criar uma campanha com CSV anexo.
//
// O CSV deve ter delimitador `,` ou `;` (detectado automaticamente) e conter
// uma coluna de telefone (phone, telefone, numero, celular ou whatsapp).
// Colunas opcionais reconhecidas: name, email, company.
type CreateCampaignRequest struct {
	// Name é o nome da campanha. Obrigatório.
	Name string
	// InstanceID é o ID da instância WhatsApp de envio (Messages.ListInstances). Obrigatório.
	InstanceID string
	// TemplateID é o gupshup_id do template HSM aprovado (Messages.ListTemplates). Obrigatório.
	TemplateID string
	// CollisionPolicy define o tratamento de contatos já existentes. Obrigatório.
	CollisionPolicy CollisionPolicy
	// FileName é o nome do arquivo CSV (ex.: contatos.csv). Obrigatório.
	FileName string
	// FileContent é o conteúdo do arquivo CSV. Obrigatório.
	FileContent []byte

	// TemplateParams são os valores das variáveis do template, na ordem em que aparecem.
	TemplateParams []string
	// Description é a descrição da campanha.
	Description string
	// StartAt é a data/hora de início do disparo (RFC3339, ex.: 2026-01-20T09:00:00-03:00).
	StartAt string
	// EndAt é a data/hora limite do disparo (RFC3339). Vazio = dispara até concluir.
	EndAt string
	// AutoStart, quando true, inicia a campanha automaticamente em StartAt.
	AutoStart *bool
	// Settings são configurações extras (ex.: {"delay_ms": 1000}).
	Settings map[string]any
}

// CreateCampaignResponse é o retorno da criação de campanha.
// A importação do CSV é assíncrona; a campanha nasce com status pending.
type CreateCampaignResponse struct {
	CampaignID string `json:"campaign_id"`
	Status     string `json:"status"`
	MailingID  string `json:"mailing_id"`
	FileName   string `json:"file_name"`
}

// CampaignResponse é a visão de uma campanha (listagem e detalhe).
type CampaignResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Channel        string `json:"channel"`
	Status         string `json:"status"`
	TemplateID     string `json:"template_id,omitempty"`
	ScheduledAt    string `json:"scheduled_at,omitempty"`
	StartAt        string `json:"start_at,omitempty"`
	EndAt          string `json:"end_at,omitempty"`
	TotalContacts  int    `json:"total_contacts"`
	SentCount      int    `json:"sent_count"`
	DeliveredCount int    `json:"delivered_count"`
	ReadCount      int    `json:"read_count"`
	RepliedCount   int    `json:"replied_count"`
	FailedCount    int    `json:"failed_count"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ListCampaignsResponse é o envelope paginado da listagem de campanhas.
type ListCampaignsResponse struct {
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Items    []CampaignResponse `json:"items"`
}
