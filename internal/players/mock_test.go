package players

import (
	"context"
	"testing"

	"arcadium.dev/arcade/internal/arcade"
)

type (
	mockService struct {
		t   *testing.T
		err error

		playerID string
		req      playerRequest

		player  arcade.Player
		players []arcade.Player

		listCalled, getCalled, createCalled, updateCalled, removeCalled bool
	}
)

func (m *mockService) list(ctx context.Context) ([]arcade.Player, error) {
	m.listCalled = true
	if m.err != nil {
		return nil, m.err
	}
	return m.players, nil
}

func (m *mockService) get(ctx context.Context, playerID string) (arcade.Player, error) {
	m.getCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.playerID != playerID {
		m.t.Fatalf("get: expected playerID %s, actual playerID %s", m.playerID, playerID)
	}
	return m.player, nil
}

func (m *mockService) create(ctx context.Context, req playerRequest) (arcade.Player, error) {
	m.createCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.req != req {
		m.t.Fatalf("create: expected player request %+v, actual player requset %+v", m.req, req)
	}
	return m.player, nil
}

func (m *mockService) update(ctx context.Context, playerID string, req playerRequest) (arcade.Player, error) {
	m.updateCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.playerID != playerID {
		m.t.Fatalf("get: expected playerID %s, actual playerID %s", m.playerID, playerID)
	}
	if m.req != req {
		m.t.Fatalf("update: expected player request %+v, actual player requset %+v", m.req, req)
	}
	return m.player, nil
}

func (m *mockService) remove(ctx context.Context, playerID string) error {
	m.removeCalled = true
	if m.err != nil {
		return m.err
	}
	if m.playerID != playerID {
		m.t.Fatalf("remove: expected playerID %s, actual playerID %s", m.playerID, playerID)
	}
	return nil
}
