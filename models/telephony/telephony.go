// Package telephony contém os modelos do domínio de telefonia PABX
// (ramais, histórico de ligações, click-to-call e gravações).
package telephony

// ExtensionResponse é um ramal da central telefônica.
type ExtensionResponse struct {
	Number     string `json:"number"`
	Name       string `json:"name"`
	CallerID   string `json:"caller_id,omitempty"`
	IsOnline   bool   `json:"is_online"`
	DNDEnabled bool   `json:"dnd_enabled"`
}

// ListExtensionsResponse é o envelope paginado da listagem de ramais.
type ListExtensionsResponse struct {
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Items    []ExtensionResponse `json:"items"`
}

// CallHistoryResponse é um registro do histórico de ligações.
type CallHistoryResponse struct {
	CallID           string `json:"call_id"`
	UniqueID         string `json:"unique_id"`
	LinkedID         string `json:"linked_id,omitempty"`
	ProtocolNumber   string `json:"protocol_number,omitempty"`
	CallType         string `json:"call_type"`
	CallResult       string `json:"call_result"`
	Status           string `json:"status"`
	CallerIDNum      string `json:"caller_id_num"`
	CallerIDName     string `json:"caller_id_name,omitempty"`
	CalledNumber     string `json:"called_number"`
	DestinationNum   string `json:"destination_num,omitempty"`
	DIDNumber        string `json:"did_number,omitempty"`
	AgentExtension   string `json:"agent_extension,omitempty"`
	AgentName        string `json:"agent_name,omitempty"`
	AgentAnsweredAt  string `json:"agent_answered_at,omitempty"`
	QueueName        string `json:"queue_name,omitempty"`
	QueueWaitTime    int    `json:"queue_wait_time,omitempty"`
	IsAbandoned      bool   `json:"is_abandoned"`
	AbandonReason    string `json:"abandon_reason,omitempty"`
	TrunkName        string `json:"trunk_name,omitempty"`
	RouteType        string `json:"route_type,omitempty"`
	HangupCause      string `json:"hangup_cause,omitempty"`
	StartTime        string `json:"start_time"`
	AnswerTime       string `json:"answer_time,omitempty"`
	EndTime          string `json:"end_time,omitempty"`
	Duration         int    `json:"duration"`
	BillableSec      int    `json:"billable_sec"`
	RecordingEnabled bool   `json:"recording_enabled"`
}

// ListCallHistoryResponse é o envelope paginado do histórico de ligações.
type ListCallHistoryResponse struct {
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Items    []CallHistoryResponse `json:"items"`
}

// OriginateCallRequest são os dados do click-to-call.
type OriginateCallRequest struct {
	// FromExtension é o ramal de origem (toca primeiro). Obrigatório.
	FromExtension string `json:"from_extension"`
	// ToNumber é o número de destino. Obrigatório.
	ToNumber string `json:"to_number"`
}

// OriginateCallResponse é o retorno da originação de chamada.
type OriginateCallResponse struct {
	ActionID  string `json:"action_id"`
	Channel   string `json:"channel,omitempty"`
	Extension string `json:"extension"`
	ToNumber  string `json:"to_number"`
}

// TempLinkResponse é um link temporário de download de gravação.
type TempLinkResponse struct {
	URL       string `json:"url"`
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	ExpiresAt string `json:"expires_at"`
}
