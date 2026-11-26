package interage

import (
	"net/url"
	"reflect"
	"strconv"
	"time"
)

// Paging são os campos de paginação comuns às listas.
type Paging struct {
	// PageSize: itens por página (padrão 50, máximo 100).
	PageSize int `query:"page_size"`
	// Cursor: o NextCursor da página anterior; vazio = primeira página.
	Cursor string `query:"cursor"`
}

// montarQuery converte um struct de filtro (campos com a tag query) na query string. Valor
// zero não vai: string vazia, 0, ponteiro nil e time.Time zero.
func montarQuery(filtro any) url.Values {
	q := url.Values{}
	if filtro == nil {
		return q
	}
	v := reflect.ValueOf(filtro)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return q
		}
		v = v.Elem()
	}
	acrescentar(q, v)
	return q
}

func acrescentar(q url.Values, v reflect.Value) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f, valor := t.Field(i), v.Field(i)
		if f.Anonymous && valor.Kind() == reflect.Struct {
			acrescentar(q, valor)
			continue
		}
		nome := f.Tag.Get("query")
		if nome == "" {
			continue
		}
		if s, ok := textoDaQuery(valor); ok {
			q.Set(nome, s)
		}
	}
}

func textoDaQuery(v reflect.Value) (string, bool) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", false
		}
		v = v.Elem()
	}
	if t, ok := v.Interface().(time.Time); ok {
		if t.IsZero() {
			return "", false
		}
		return t.Format(time.RFC3339), true
	}
	switch v.Kind() {
	case reflect.String:
		return v.String(), v.String() != ""
	case reflect.Int, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), v.Int() != 0
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), true
	}
	return "", false
}
