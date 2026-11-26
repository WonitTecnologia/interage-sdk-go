package interage

import "encoding/json"

// Opt é um campo de alteração parcial (PATCH de contato e de ramal, PUT de desvios), em que a
// API distingue três estados: ausente (o campo fica como está), null (limpa ou desliga) e um
// valor. O valor zero de Opt é "ausente" e não vai no corpo.
//
//	interage.ContactUpdate{
//		Email:  interage.Set("joao@empresa.com"), // troca
//		Company: interage.Null[string](),          // limpa
//		// Name ausente: não muda
//	}
type Opt[T any] struct {
	set, null bool
	v         T
}

// Set devolve um Opt com o valor.
func Set[T any](v T) Opt[T] { return Opt[T]{set: true, v: v} }

// Null devolve um Opt que vai como null: limpa o campo (ou desliga, no desvio).
func Null[T any]() Opt[T] { return Opt[T]{set: true, null: true} }

// IsZero informa se o campo está ausente (usado pelo encoding/json com omitzero).
func (o Opt[T]) IsZero() bool { return !o.set }

// IsNull informa se o campo vai como null.
func (o Opt[T]) IsNull() bool { return o.set && o.null }

// Value devolve o valor e se há um (false quando ausente ou null).
func (o Opt[T]) Value() (T, bool) { return o.v, o.set && !o.null }

func (o Opt[T]) MarshalJSON() ([]byte, error) {
	if o.null || !o.set {
		return []byte("null"), nil
	}
	return json.Marshal(o.v)
}

func (o *Opt[T]) UnmarshalJSON(b []byte) error {
	o.set = true
	if string(b) == "null" {
		o.null = true
		var zero T
		o.v = zero
		return nil
	}
	o.null = false
	return json.Unmarshal(b, &o.v)
}
