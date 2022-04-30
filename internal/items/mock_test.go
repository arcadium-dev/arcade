package items

import (
	"context"
	"testing"

	"arcadium.dev/arcade/internal/arcade"
)

type (
	mockService struct {
		t   *testing.T
		err error

		itemID string
		req    itemRequest

		item  arcade.Item
		items []arcade.Item

		listCalled, getCalled, createCalled, updateCalled, removeCalled bool
	}
)

func (m *mockService) list(ctx context.Context) ([]arcade.Item, error) {
	m.listCalled = true
	if m.err != nil {
		return nil, m.err
	}
	return m.items, nil
}

func (m *mockService) get(ctx context.Context, itemID string) (arcade.Item, error) {
	m.getCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.itemID != itemID {
		m.t.Fatalf("get: expected itemID %s, actual itemID %s", m.itemID, itemID)
	}
	return m.item, nil
}

func (m *mockService) create(ctx context.Context, req itemRequest) (arcade.Item, error) {
	m.createCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.req != req {
		m.t.Fatalf("create: expected item request %+v, actual item requset %+v", m.req, req)
	}
	return m.item, nil
}

func (m *mockService) update(ctx context.Context, itemID string, req itemRequest) (arcade.Item, error) {
	m.updateCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.itemID != itemID {
		m.t.Fatalf("get: expected itemID %s, actual itemID %s", m.itemID, itemID)
	}
	if m.req != req {
		m.t.Fatalf("update: expected item request %+v, actual item requset %+v", m.req, req)
	}
	return m.item, nil
}

func (m *mockService) remove(ctx context.Context, itemID string) error {
	m.removeCalled = true
	if m.err != nil {
		return m.err
	}
	if m.itemID != itemID {
		m.t.Fatalf("remove: expected itemID %s, actual itemID %s", m.itemID, itemID)
	}
	return nil
}
