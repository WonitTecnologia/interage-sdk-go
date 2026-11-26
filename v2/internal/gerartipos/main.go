// Command gerartipos gera types_gen.go (e o registro do teste, tipos_gen_test.go) a partir do
// código-fonte dos DTOs da API pública v2 — os tipos com //@name, as mesmas tags JSON e os
// ponteiros de onde a API devolve null. Rodar a cada mudança no contrato da v2:
//
//	go run ./internal/gerartipos -api ../../wonit-pabx-app-api
//
// e copiar o openapi da API para testdata/openapi.json (os testes conferem os dois).
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

var raiz string

const modulo = "github.com/WonitTecnologia/wonit-pabx-app-api"

type tipo struct {
	pkgDir string
	nome   string // nome Go na API
	sdk    string // nome no SDK (@name)
	spec   *ast.TypeSpec
	doc    string
	file   *ast.File
}

var porChave = map[string]*tipo{} // pkgDir.Nome
var gerados []string              // structs escritos (registro do teste)
var faltando = map[string]bool{}

func chave(dir, nome string) string { return dir + "." + nome }

func main() {
	flag.StringVar(&raiz, "api", "", "caminho do repositório wonit-pabx-app-api")
	saida := flag.String("out", ".", "pasta do módulo v2 do SDK")
	flag.Parse()
	if raiz == "" {
		fmt.Fprintln(os.Stderr, "informe -api <caminho do wonit-pabx-app-api>")
		os.Exit(2)
	}
	raiz = strings.TrimRight(raiz, "/")
	dirs := []string{}
	filepath.Walk(raiz+"/internal/publicapi/v2", func(p string, i os.FileInfo, _ error) error {
		if i.IsDir() {
			dirs = append(dirs, p)
		}
		return nil
	})
	dirs = append(dirs, raiz+"/internal/service/webhook")
	fset := token.NewFileSet()
	for _, d := range dirs {
		pkgs, err := parser.ParseDir(fset, d, func(fi os.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		for _, p := range pkgs {
			for fn, f := range p.Files {
				src, _ := os.ReadFile(fn)
				for _, decl := range f.Decls {
					g, ok := decl.(*ast.GenDecl)
					if !ok || g.Tok != token.TYPE {
						continue
					}
					for _, s := range g.Specs {
						ts := s.(*ast.TypeSpec)
						t := &tipo{pkgDir: d, nome: ts.Name.Name, spec: ts, file: f}
						if g.Doc != nil {
							t.doc = g.Doc.Text()
						} else if ts.Doc != nil {
							t.doc = ts.Doc.Text()
						}
						// @name fica na linha do "}" de fechamento
						fim := fset.Position(ts.End()).Offset
						resto := string(src[fim:])
						if nl := strings.IndexByte(resto, '\n'); nl >= 0 {
							resto = resto[:nl]
						}
						if i := strings.Index(resto, "@name "); i >= 0 {
							t.sdk = strings.TrimSpace(resto[i+6:])
						}
						porChave[chave(d, t.nome)] = t
					}
				}
			}
		}
	}
	// segue aliases (type A = pkg.B) e monta a saída a partir dos tipos com @name
	var nomes []string
	porSdk := map[string]*tipo{}
	for _, t := range porChave {
		if t.sdk != "" && !strings.HasPrefix(t.sdk, "Webhook") {
			porSdk[t.sdk] = t
			nomes = append(nomes, t.sdk)
		}
	}
	sort.Strings(nomes)
	var b bytes.Buffer
	b.WriteString("// Code generated from the Interage API v2 DTOs. DO NOT EDIT.\n// Tipos da API pública v2, gerados a partir do código da API (mesmas tags JSON).\n\npackage interage\n\nimport (\n\t\"encoding/json\"\n\t\"time\"\n)\n\nvar _ = json.RawMessage(nil)\nvar _ = time.Time{}\n\n")
	for _, n := range nomes {
		t := porSdk[n]
		if envelope(t) {
			continue // o SDK devolve o conteúdo de data, não o envelope
		}
		escreverTipo(&b, t, n)
		if _, ok := resolverAlias(t).spec.Type.(*ast.StructType); ok {
			gerados = append(gerados, n)
		}
	}
	out, err := format.Source(b.Bytes())
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(*saida, "types_gen.go"), out, 0o644); err != nil {
		panic(err)
	}
	var r bytes.Buffer
	r.WriteString("// Code generated junto com types_gen.go. DO NOT EDIT.\n\npackage interage\n\nimport \"reflect\"\n\nfunc tiposGeradosParaTeste() map[string]reflect.Type {\n\treturn map[string]reflect.Type{\n")
	for _, n := range gerados {
		fmt.Fprintf(&r, "\t\t%q: reflect.TypeOf(%s{}),\n", n, n)
	}
	r.WriteString("\t}\n}\n")
	reg, err := format.Source(r.Bytes())
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(*saida, "tipos_gen_test.go"), reg, 0o644); err != nil {
		panic(err)
	}
	if len(faltando) > 0 {
		var l []string
		for k := range faltando {
			l = append(l, k)
		}
		sort.Strings(l)
		fmt.Fprintln(os.Stderr, "SEM @name (resolvidos como inline/nome gerado):", strings.Join(l, ", "))
	}
}

// envelope: struct só com data (e pagination).
func envelope(t *tipo) bool {
	st, ok := resolverAlias(t).spec.Type.(*ast.StructType)
	if !ok {
		return false
	}
	temData := false
	for _, f := range st.Fields.List {
		if f.Tag == nil {
			return false
		}
		j := strings.Split(reflect.StructTag(strings.Trim(f.Tag.Value, "`")).Get("json"), ",")[0]
		switch j {
		case "data":
			temData = true
		case "pagination":
		default:
			return false
		}
	}
	return temData
}

func comentario(b *bytes.Buffer, doc, indent string) {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return
	}
	for _, l := range strings.Split(doc, "\n") {
		b.WriteString(indent + "// " + l + "\n")
	}
}

// Corpos em que a API distingue campo ausente (não muda), null (limpa/desliga) e valor.
var triEstado = map[string]bool{
	"ContactUpdate": true, "ExtensionUpdate": true, "ExtensionCapabilitiesUpdate": true,
	"ExtensionDialingUpdate": true, "ExtensionVoicemailUpdate": true, "ExtensionSipUpdate": true,
	"ExtensionSessionTimersUpdate": true, "ExtensionForwardingRequest": true,
}

func escreverTipo(b *bytes.Buffer, t *tipo, nome string) {
	t = resolverAlias(t)
	doc := t.doc
	if strings.HasPrefix(doc, t.nome+" ") {
		doc = nome + doc[len(t.nome):]
	}
	comentario(b, doc, "")
	switch st := t.spec.Type.(type) {
	case *ast.StructType:
		fmt.Fprintf(b, "type %s struct {\n", nome)
		for _, f := range st.Fields.List {
			tag := ""
			if f.Tag != nil {
				tag = strings.Trim(f.Tag.Value, "`")
			}
			js := reflect.StructTag(tag).Get("json")
			jnome := strings.Split(js, ",")[0]
			if f.Doc != nil {
				comentario(b, f.Doc.Text(), "\t")
			}
			if en := reflect.StructTag(tag).Get("enums"); en != "" {
				b.WriteString("\t// Valores: " + strings.ReplaceAll(en, ",", ", ") + ".\n")
			}
			tipoGo := traduzir(t, f.Type)
			linha := ""
			if f.Comment != nil {
				linha = " // " + strings.TrimSpace(f.Comment.Text())
			}
			if len(f.Names) == 0 { // embutido
				fmt.Fprintf(b, "\t%s%s\n", tipoGo, linha)
				continue
			}
			if jnome == "-" {
				continue
			}
			campo := camel(jnome)
			if jnome == "" {
				campo = f.Names[0].Name
			}
			if triEstado[nome] {
				fmt.Fprintf(b, "\t%s Opt[%s] `json:\"%s,omitzero\"`%s\n", campo, strings.TrimPrefix(tipoGo, "*"), jnome, linha)
				continue
			}
			fmt.Fprintf(b, "\t%s %s `json:\"%s\"`%s\n", campo, tipoGo, js, linha)
		}
		b.WriteString("}\n\n")
	default:
		fmt.Fprintf(b, "type %s %s\n\n", nome, traduzir(t, t.spec.Type))
	}
}

func resolverAlias(t *tipo) *tipo {
	for t.spec.Assign.IsValid() {
		alvo := buscar(t, t.spec.Type)
		if alvo == nil {
			break
		}
		t = alvo
	}
	return t
}

func importDir(t *tipo, alias string) string {
	for _, im := range t.file.Imports {
		path := strings.Trim(im.Path.Value, `"`)
		nome := filepath.Base(path)
		if im.Name != nil {
			nome = im.Name.Name
		}
		if nome == alias && strings.HasPrefix(path, modulo) {
			return raiz + strings.TrimPrefix(path, modulo)
		}
	}
	return ""
}

func buscar(t *tipo, e ast.Expr) *tipo {
	switch x := e.(type) {
	case *ast.Ident:
		return porChave[chave(t.pkgDir, x.Name)]
	case *ast.SelectorExpr:
		d := importDir(t, x.X.(*ast.Ident).Name)
		if d == "" {
			return nil
		}
		return porChave[chave(d, x.Sel.Name)]
	}
	return nil
}

func traduzir(t *tipo, e ast.Expr) string {
	switch x := e.(type) {
	case *ast.StarExpr:
		return "*" + traduzir(t, x.X)
	case *ast.ArrayType:
		return "[]" + traduzir(t, x.Elt)
	case *ast.MapType:
		return "map[" + traduzir(t, x.Key) + "]" + traduzir(t, x.Value)
	case *ast.InterfaceType:
		return "any"
	case *ast.Ident:
		switch x.Name {
		case "string", "int", "int64", "int32", "int16", "int8", "uint", "uint64", "float64", "float32", "bool", "any", "byte":
			return x.Name
		}
	case *ast.SelectorExpr:
		pk := x.X.(*ast.Ident).Name
		switch pk + "." + x.Sel.Name {
		case "time.Time":
			return "time.Time"
		case "json.RawMessage":
			return "json.RawMessage"
		case "uuid.UUID":
			return "string"
		}
	}
	alvo := buscar(t, e)
	if alvo == nil {
		faltando[fmt.Sprintf("%T %v em %s.%s", e, e, filepath.Base(t.pkgDir), t.nome)] = true
		return "any"
	}
	real := resolverAlias(alvo)
	if real.sdk != "" {
		return real.sdk
	}
	if alvo.sdk != "" {
		return alvo.sdk
	}
	// tipo sem @name: básico por baixo (ex.: type Codigo string)
	if id, ok := real.spec.Type.(*ast.Ident); ok {
		return traduzir(real, id)
	}
	faltando[filepath.Base(real.pkgDir)+"."+real.nome] = true
	return "any"
}

var iniciais = map[string]string{"id": "ID", "url": "URL", "sip": "SIP", "dnd": "DND", "did": "DID", "ip": "IP", "api": "API", "uuid": "UUID", "csv": "CSV", "ivr": "IVR", "uri": "URI", "cpf": "CPF", "ids": "IDs", "dtmf": "DTMF", "sms": "SMS", "http": "HTTP"}

func camel(s string) string {
	var out strings.Builder
	for _, p := range strings.Split(s, "_") {
		if p == "" {
			continue
		}
		if v, ok := iniciais[p]; ok {
			out.WriteString(v)
			continue
		}
		out.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return out.String()
}
