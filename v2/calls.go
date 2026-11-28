package interage

import (
	"context"
	"iter"
	"time"
)

// CallsService cobre as ligações (/calls).
type CallsService struct{ c *httpClient }

// ListCallsParams filtra o histórico. From e To são obrigatórios (To é exclusivo).
type ListCallsParams struct {
	From time.Time `query:"from"`
	To   time.Time `query:"to"`
	// Extension: ramal exato; com * no fim, prefixo (ex.: "10*").
	Extension string `query:"extension"`
	// Direction: inbound, outbound ou internal.
	Direction string `query:"direction"`
	// Result: answered, no_answer, busy, failed, abandoned, ivr_ended, voicemail, unavailable ou blocked.
	Result string `query:"result"`
	// Number: trecho do número na origem ou no destino.
	Number string `query:"number"`
	Paging
}

// List devolve uma página do histórico de ligações.
func (s *CallsService) List(ctx context.Context, p ListCallsParams) (*Page[Call], error) {
	return lista[Call](ctx, s.c, rotaLigacoes.caminho(), montarQuery(p))
}

// All percorre todo o histórico do filtro, página a página.
func (s *CallsService) All(ctx context.Context, p ListCallsParams) iter.Seq2[Call, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[Call], error) {
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// Get devolve o detalhe da ligação (com a linha do tempo).
func (s *CallsService) Get(ctx context.Context, callID string) (*CallDetail, error) {
	var out CallDetail
	if err := s.c.get(ctx, rotaLigacao.caminho(callID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListActiveCallsParams pagina as ligações em curso.
type ListActiveCallsParams struct {
	Paging
}

// ListActive devolve uma página das ligações em curso, da mais recente para a mais antiga. É
// uma foto do momento: a que termina entre uma página e outra some, e a que começa depois da
// primeira página só aparece numa nova consulta.
func (s *CallsService) ListActive(ctx context.Context, p ListActiveCallsParams) (*Page[ActiveCall], error) {
	return lista[ActiveCall](ctx, s.c, rotaLigacoesAtivas.caminho(), montarQuery(p))
}

// AllActive percorre todas as ligações em curso.
func (s *CallsService) AllActive(ctx context.Context, p ListActiveCallsParams) iter.Seq2[ActiveCall, error] {
	return todos(ctx, func(ctx context.Context, cursor string) (*Page[ActiveCall], error) {
		p.Cursor = cursor
		return s.ListActive(ctx, p)
	})
}

// Originate faz o click to call: toca o ramal e, ao atender, liga para o destino (202).
func (s *CallsService) Originate(ctx context.Context, extension, destination string) (*CallRequested, error) {
	var out CallRequested
	in := CallRequest{Extension: extension, Destination: destination}
	if err := s.c.json(ctx, rotaOriginar.method, rotaOriginar.caminho(), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// RecordingLink gera o link temporário da gravação. expiresIn zero usa o padrão da API (1 hora);
// o máximo é 7 dias.
func (s *CallsService) RecordingLink(ctx context.Context, callID string, expiresIn time.Duration) (*RecordingLink, error) {
	var out RecordingLink
	var in RecordingLinkRequest
	if expiresIn > 0 {
		seg := int(expiresIn.Seconds())
		in.ExpiresIn = &seg
	}
	if err := s.c.json(ctx, rotaLinkGravacao.method, rotaLinkGravacao.caminho(callID), nil, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}
