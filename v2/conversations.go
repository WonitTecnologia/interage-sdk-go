package interage

import (
	"context"
	"iter"
	"strconv"
	"time"
)

// ConversationsService cobre as conversas (/conversations) e as mensagens delas.
type ConversationsService struct{ c *httpClient }

// ConversationFilter são os filtros comuns às duas listas de conversas.
type ConversationFilter struct {
	// ChannelType: whatsapp, instagram, facebook, telegram, webchat ou email.
	ChannelType string `query:"channel_type"`
	InstanceID  string `query:"instance_id"`
	// QueueID: fila de chat (GET /queues).
	QueueID string `query:"queue_id"`
	// AgentID: usuário atendente (GET /users).
	AgentID   string `query:"agent_id"`
	ContactID string `query:"contact_id"`
}

// ListConversationsParams filtra o histórico (encerradas). From e To são pelo encerramento,
// obrigatórios (To exclusivo).
type ListConversationsParams struct {
	From time.Time `query:"from"`
	To   time.Time `query:"to"`
	ConversationFilter
	Paging
}

// ListActiveConversationsParams filtra as conversas em andamento.
type ListActiveConversationsParams struct {
	// Status: bot, ai_agent, queue, attending ou coex.
	Status string `query:"status"`
	ConversationFilter
	Paging
}

// List devolve uma página do histórico de conversas encerradas.
func (s *ConversationsService) List(ctx context.Context, p ListConversationsParams) (*Page[Conversation], error) {
	return lista[Conversation](ctx, s.c, rotaConversas.caminho(), montarQuery(p))
}

// All percorre todo o histórico do filtro, página a página.
func (s *ConversationsService) All(ctx context.Context, p ListConversationsParams) iter.Seq2[Conversation, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[Conversation], error) {
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// ListActive devolve uma página das conversas em andamento.
func (s *ConversationsService) ListActive(ctx context.Context, p ListActiveConversationsParams) (*Page[Conversation], error) {
	return lista[Conversation](ctx, s.c, rotaConversasAtivas.caminho(), montarQuery(p))
}

// AllActive percorre todas as conversas em andamento do filtro.
func (s *ConversationsService) AllActive(ctx context.Context, p ListActiveConversationsParams) iter.Seq2[Conversation, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[Conversation], error) {
		p.Cursor = cursor
		return s.ListActive(ctx, p)
	})
}

// Get devolve o detalhe da conversa pelo protocolo.
func (s *ConversationsService) Get(ctx context.Context, protocol string) (*ConversationDetail, error) {
	var out ConversationDetail
	if err := s.c.get(ctx, rotaConversa.caminho(protocol), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Start inicia uma conversa com um template (WhatsApp). Conversa já ativa com o contato:
// CONVERSATION_ALREADY_ACTIVE, com o protocolo dela em Error.Extra("protocol", ...).
func (s *ConversationsService) Start(ctx context.Context, in ConversationStartRequest) (*ConversationStarted, error) {
	var out ConversationStarted
	if err := s.c.json(ctx, rotaIniciarConversa.method, rotaIniciarConversa.caminho(), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// Close encerra a conversa. classification: uma das opções da fila (vazia se a fila não exige).
func (s *ConversationsService) Close(ctx context.Context, protocol, classification string) (*ConversationDetail, error) {
	var out ConversationDetail
	var in ConversationCloseRequest
	if classification != "" {
		in.Classification = &classification
	}
	if err := s.c.json(ctx, rotaEncerrar.method, rotaEncerrar.caminho(protocol), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// Transfer transfere para uma fila (queueID), para um usuário (userID) ou para os dois.
func (s *ConversationsService) Transfer(ctx context.Context, protocol string, in ConversationTransferRequest) (*ConversationDetail, error) {
	var out ConversationDetail
	if err := s.c.json(ctx, rotaTransferir.method, rotaTransferir.caminho(protocol), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListMessages devolve uma página do histórico de mensagens, em ordem cronológica.
func (s *ConversationsService) ListMessages(ctx context.Context, protocol string, p Paging) (*Page[Message], error) {
	return lista[Message](ctx, s.c, rotaMensagens.caminho(protocol), montarQuery(p))
}

// AllMessages percorre todas as mensagens da conversa.
func (s *ConversationsService) AllMessages(ctx context.Context, protocol string, p Paging) iter.Seq2[Message, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[Message], error) {
		p.Cursor = cursor
		return s.ListMessages(ctx, protocol, p)
	})
}

// GetMessage devolve a mensagem, com status e erro de entrega.
func (s *ConversationsService) GetMessage(ctx context.Context, protocol string, id int64) (*MessageDetail, error) {
	var out MessageDetail
	if err := s.c.get(ctx, rotaMensagem.caminho(protocol, strconv.FormatInt(id, 10)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendMessage envia uma mensagem livre na conversa (WhatsApp, dentro da janela de 24 h).
func (s *ConversationsService) SendMessage(ctx context.Context, protocol string, in MessageSendRequest) (*MessageDetail, error) {
	var out MessageDetail
	if err := s.c.json(ctx, rotaEnviarMensagem.method, rotaEnviarMensagem.caminho(protocol), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// FileLink gera o link temporário do arquivo da mensagem (Media.Downloadable). expiresIn
// zero usa o padrão da API.
func (s *ConversationsService) FileLink(ctx context.Context, protocol string, id int64, expiresIn time.Duration) (*MessageFileLink, error) {
	var out MessageFileLink
	var in MessageFileLinkRequest
	if expiresIn > 0 {
		seg := int(expiresIn.Seconds())
		in.ExpiresIn = &seg
	}
	if err := s.c.json(ctx, rotaLinkArquivo.method, rotaLinkArquivo.caminho(protocol, strconv.FormatInt(id, 10)), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// MessagesService cobre o template avulso (POST /messages).
type MessagesService struct{ c *httpClient }

// SendTemplate envia um template fora de conversa. A resposta do contato em até 24 h reabre a
// conversa com o mesmo protocolo.
func (s *MessagesService) SendTemplate(ctx context.Context, in MessageTemplateRequest) (*MessageTemplateSent, error) {
	var out MessageTemplateSent
	if err := s.c.json(ctx, rotaTemplateAvulso.method, rotaTemplateAvulso.caminho(), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}
