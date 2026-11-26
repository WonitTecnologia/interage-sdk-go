package interage

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// testdata/openapi.json é o /api/public/v2/docs/openapi.json da API, copiado a cada versão do
// SDK (docs/publicapi/v2/publicv2_swagger.json no repositório da API). Estes testes falham
// quando o SDK e a API se afastam.

type especificacao struct {
	BasePath    string                                `json:"basePath"`
	Paths       map[string]map[string]json.RawMessage `json:"paths"`
	Definitions map[string]struct {
		Properties map[string]json.RawMessage `json:"properties"`
		Enum       []string                   `json:"enum"`
	} `json:"definitions"`
}

func lerEspecificacao(t *testing.T) especificacao {
	t.Helper()
	b, err := os.ReadFile("testdata/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var e especificacao
	if err := json.Unmarshal(b, &e); err != nil {
		t.Fatal(err)
	}
	return e
}

// Toda rota da API tem método no SDK, e o SDK não chama rota que a API não tem.
func TestRotasCasamComAAPI(t *testing.T) {
	e := lerEspecificacao(t)
	daAPI := map[string]bool{}
	for caminho, ops := range e.Paths {
		for metodo := range ops {
			daAPI[strings.ToUpper(metodo)+" "+strings.TrimPrefix(caminho, "/public/v2")] = true
		}
	}
	doSDK := map[string]bool{}
	for _, r := range rotas {
		doSDK[r.method+" "+r.path] = true
	}
	for k := range daAPI {
		if !doSDK[k] {
			t.Errorf("rota da API sem método no SDK: %s", k)
		}
	}
	for k := range doSDK {
		if !daAPI[k] {
			t.Errorf("rota do SDK que a API não tem: %s", k)
		}
	}
}

func camposJSON(t reflect.Type) []string {
	var out []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		nome, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if f.Anonymous && nome == "" {
			out = append(out, camposJSON(f.Type)...)
			continue
		}
		if nome != "" && nome != "-" {
			out = append(out, nome)
		}
	}
	sort.Strings(out)
	return out
}

func propriedades(e especificacao, nome string) []string {
	var out []string
	for k := range e.Definitions[nome].Properties {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Os tipos gerados têm os mesmos campos JSON das definições do Swagger (e o webhook, os dos
// modelos Webhook* da API). Problem tem membros extras fora da definição, lidos à parte.
func TestTiposCasamComAsDefinicoes(t *testing.T) {
	e := lerEspecificacao(t)
	tipos := map[string]reflect.Type{}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "types_gen.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Decls {
		if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.TYPE {
			for _, s := range g.Specs {
				if ts := s.(*ast.TypeSpec); ts != nil {
					if _, ok := ts.Type.(*ast.StructType); ok {
						tipos[ts.Name.Name] = nil
					}
				}
			}
		}
	}
	registro := tiposGeradosParaTeste()
	for nome := range tipos {
		tp, ok := registro[nome]
		if !ok {
			t.Errorf("%s gerado mas fora de tiposGeradosParaTeste (regere o arquivo de teste)", nome)
			continue
		}
		if _, ok := e.Definitions[nome]; !ok {
			t.Errorf("%s não existe no openapi.json", nome)
			continue
		}
		if got, want := camposJSON(tp), propriedades(e, nome); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: SDK %v, API %v", nome, got, want)
		}
	}
	webhook := map[string]reflect.Type{
		"WebhookLote":          reflect.TypeOf(WebhookBatch{}),
		"WebhookEvento":        reflect.TypeOf(WebhookEvent{}),
		"WebhookMensagem":      reflect.TypeOf(WebhookMessage{}),
		"WebhookMensagemMidia": reflect.TypeOf(WebhookMessageMedia{}),
		"WebhookRamal":         reflect.TypeOf(WebhookExtension{}),
	}
	for nome, tp := range webhook {
		if got, want := camposJSON(tp), propriedades(e, nome); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: SDK %v, API %v", nome, got, want)
		}
	}
}

// As constantes Code* são exatamente os códigos do enum ErrorCode da API.
func TestCodigosDeErroCasamComAAPI(t *testing.T) {
	e := lerEspecificacao(t)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "errors.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var doSDK []string
	ast.Inspect(f, func(n ast.Node) bool {
		if vs, ok := n.(*ast.ValueSpec); ok && len(vs.Names) == 1 && strings.HasPrefix(vs.Names[0].Name, "Code") {
			if lit, ok := vs.Values[0].(*ast.BasicLit); ok {
				v, _ := strconv.Unquote(lit.Value)
				doSDK = append(doSDK, v)
			}
		}
		return true
	})
	daAPI := append([]string(nil), e.Definitions["ErrorCode"].Enum...)
	sort.Strings(doSDK)
	sort.Strings(daAPI)
	if !reflect.DeepEqual(doSDK, daAPI) {
		t.Errorf("códigos: SDK %v\nAPI %v", doSDK, daAPI)
	}
}
