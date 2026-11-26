package interage

import (
	"net/http"
	"strings"
)

// rota é uma operação da API v2: método e caminho relativo a /api/public/v2, com {param}.
// rotas é a tabela que os métodos usam; o teste confere que ela casa com o openapi.json da v2
// (nenhuma rota da API sem método no SDK, nenhuma rota no SDK que a API não tem).
type rota struct {
	method, path string
}

// caminho troca os {param} do caminho pelos valores, na ordem, já escapados.
func (r rota) caminho(valores ...string) string {
	p := r.path
	for _, v := range valores {
		i := strings.IndexByte(p, '{')
		j := strings.IndexByte(p, '}')
		if i < 0 || j < i {
			break
		}
		p = p[:i] + seg(v) + p[j+1:]
	}
	return p
}

var (
	rotaIdentificacao = rota{http.MethodGet, ""}

	rotaRamais       = rota{http.MethodGet, "/extensions"}
	rotaCriarRamal   = rota{http.MethodPost, "/extensions"}
	rotaRamal        = rota{http.MethodGet, "/extensions/{number}"}
	rotaAlterarRamal = rota{http.MethodPatch, "/extensions/{number}"}
	rotaExcluirRamal = rota{http.MethodDelete, "/extensions/{number}"}
	rotaDndRamal     = rota{http.MethodPut, "/extensions/{number}/dnd"}
	rotaDesviosRamal = rota{http.MethodPut, "/extensions/{number}/forwarding"}
	rotaSenhaRamal   = rota{http.MethodPost, "/extensions/{number}/password"}

	rotaLigacoes       = rota{http.MethodGet, "/calls"}
	rotaOriginar       = rota{http.MethodPost, "/calls"}
	rotaLigacoesAtivas = rota{http.MethodGet, "/calls/active"}
	rotaLigacao        = rota{http.MethodGet, "/calls/{call_id}"}
	rotaLinkGravacao   = rota{http.MethodPost, "/calls/{call_id}/recording-link"}

	rotaContatos       = rota{http.MethodGet, "/contacts"}
	rotaCriarContato   = rota{http.MethodPost, "/contacts"}
	rotaLoteContatos   = rota{http.MethodPost, "/contacts/batch"}
	rotaContato        = rota{http.MethodGet, "/contacts/{id}"}
	rotaAlterarContato = rota{http.MethodPatch, "/contacts/{id}"}
	rotaExcluirContato = rota{http.MethodDelete, "/contacts/{id}"}

	rotaInstancias = rota{http.MethodGet, "/instances"}
	rotaTemplates  = rota{http.MethodGet, "/templates"}
	rotaFilas      = rota{http.MethodGet, "/queues"}
	rotaUsuarios   = rota{http.MethodGet, "/users"}
	rotaUsuario    = rota{http.MethodGet, "/users/{id}"}

	rotaConversas       = rota{http.MethodGet, "/conversations"}
	rotaIniciarConversa = rota{http.MethodPost, "/conversations"}
	rotaConversasAtivas = rota{http.MethodGet, "/conversations/active"}
	rotaConversa        = rota{http.MethodGet, "/conversations/{protocol}"}
	rotaEncerrar        = rota{http.MethodPost, "/conversations/{protocol}/close"}
	rotaTransferir      = rota{http.MethodPost, "/conversations/{protocol}/transfer"}
	rotaMensagens       = rota{http.MethodGet, "/conversations/{protocol}/messages"}
	rotaEnviarMensagem  = rota{http.MethodPost, "/conversations/{protocol}/messages"}
	rotaMensagem        = rota{http.MethodGet, "/conversations/{protocol}/messages/{id}"}
	rotaLinkArquivo     = rota{http.MethodPost, "/conversations/{protocol}/messages/{id}/file-link"}

	rotaTemplateAvulso = rota{http.MethodPost, "/messages"}

	rotaCampanhas        = rota{http.MethodGet, "/campaigns"}
	rotaCriarCampanha    = rota{http.MethodPost, "/campaigns"}
	rotaCampanha         = rota{http.MethodGet, "/campaigns/{id}"}
	rotaExcluirCampanha  = rota{http.MethodDelete, "/campaigns/{id}"}
	rotaIniciarCampanha  = rota{http.MethodPost, "/campaigns/{id}/start"}
	rotaPausarCampanha   = rota{http.MethodPost, "/campaigns/{id}/pause"}
	rotaCancelarCampanha = rota{http.MethodPost, "/campaigns/{id}/cancel"}
	rotaContatosCampanha = rota{http.MethodGet, "/campaigns/{id}/contacts"}
)

// rotas é a tabela completa (usada pelo teste de cobertura).
var rotas = []rota{
	rotaIdentificacao,
	rotaRamais, rotaCriarRamal, rotaRamal, rotaAlterarRamal, rotaExcluirRamal, rotaDndRamal, rotaDesviosRamal, rotaSenhaRamal,
	rotaLigacoes, rotaOriginar, rotaLigacoesAtivas, rotaLigacao, rotaLinkGravacao,
	rotaContatos, rotaCriarContato, rotaLoteContatos, rotaContato, rotaAlterarContato, rotaExcluirContato,
	rotaInstancias, rotaTemplates, rotaFilas, rotaUsuarios, rotaUsuario,
	rotaConversas, rotaIniciarConversa, rotaConversasAtivas, rotaConversa, rotaEncerrar, rotaTransferir,
	rotaMensagens, rotaEnviarMensagem, rotaMensagem, rotaLinkArquivo,
	rotaTemplateAvulso,
	rotaCampanhas, rotaCriarCampanha, rotaCampanha, rotaExcluirCampanha, rotaIniciarCampanha, rotaPausarCampanha,
	rotaCancelarCampanha, rotaContatosCampanha,
}
