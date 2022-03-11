package players

import (
	"context"
	"net/http"
)

type (
	handler struct {
		s service
	}

	service interface {
		list(ctx context.Context) ([]player, error)
		get(ctx context.Context, playerID string) (player, error)
		create(ctx context.Context, p player) error
		update(ctx context.Context, p player) error
		remove(ctx context.Context, playerID string) error
	}
)

func (h handler) list(w http.ResponseWriter, r *http.Request) {
}

func (h handler) get(w http.ResponseWriter, r *http.Request) {
}

func (h handler) create(w http.ResponseWriter, r *http.Request) {
}

func (h handler) update(w http.ResponseWriter, r *http.Request) {
}

func (h handler) remove(w http.ResponseWriter, r *http.Request) {
}
