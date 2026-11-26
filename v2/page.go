package interage

import (
	"context"
	"iter"
)

// Page é uma página de uma lista. NextCursor nil = não há próxima.
type Page[T any] struct {
	Items      []T
	NextCursor *string
	PageSize   int
	Total      int64 // total de itens do filtro
}

// HasNext informa se há próxima página.
func (p *Page[T]) HasNext() bool { return p.NextCursor != nil && *p.NextCursor != "" }

// todos percorre todas as páginas, pedindo a próxima só quando a anterior acaba. Para no
// primeiro erro (devolvido como o último par) ou quando quem itera para.
func todos[T any](ctx context.Context, pagina func(ctx context.Context, cursor string) (*Page[T], error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		cursor := ""
		for {
			p, err := pagina(ctx, cursor)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range p.Items {
				if !yield(item, nil) {
					return
				}
			}
			if !p.HasNext() {
				return
			}
			cursor = *p.NextCursor
		}
	}
}
