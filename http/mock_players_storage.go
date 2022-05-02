package http

import (
	"context"
	"testing"

	"arcadium.dev/arcade"
)

type (
	mockPlayersStorage struct {
		t   *testing.T
		err error

		playerID string
		req      arcade.PlayerRequest

		player  arcade.Player
		players []arcade.Player

		listCalled, getCalled, createCalled, updateCalled, removeCalled bool
	}
)

func (m *mockPlayersStorage) List(context.Context, arcade.PlayersFilter) ([]arcade.Player, error) {
	m.listCalled = true
	if m.err != nil {
		return nil, m.err
	}
	return m.players, nil
}

func (m *mockPlayersStorage) Get(ctx context.Context, playerID string) (arcade.Player, error) {
	m.getCalled = true
	if m.err != nil {
		return arcade.Player{}, m.err
	}
	if m.playerID != playerID {
		m.t.Fatalf("get: expected playerID %s, actual playerID %s", m.playerID, playerID)
	}
	return m.player, nil
}

func (m *mockPlayersStorage) Create(ctx context.Context, req arcade.PlayerRequest) (arcade.Player, error) {
	m.createCalled = true
	if m.err != nil {
		return arcade.Player{}, m.err
	}
	if m.req != req {
		m.t.Fatalf("create: expected player request %+v, actual player requset %+v", m.req, req)
	}
	return m.player, nil
}

func (m *mockPlayersStorage) Update(ctx context.Context, playerID string, req arcade.PlayerRequest) (arcade.Player, error) {
	m.updateCalled = true
	if m.err != nil {
		return arcade.Player{}, m.err
	}
	if m.playerID != playerID {
		m.t.Fatalf("get: expected playerID %s, actual playerID %s", m.playerID, playerID)
	}
	if m.req != req {
		m.t.Fatalf("update: expected player request %+v, actual player requset %+v", m.req, req)
	}
	return m.player, nil
}

func (m *mockPlayersStorage) Remove(ctx context.Context, playerID string) error {
	m.removeCalled = true
	if m.err != nil {
		return m.err
	}
	if m.playerID != playerID {
		m.t.Fatalf("remove: expected playerID %s, actual playerID %s", m.playerID, playerID)
	}
	return nil
}
