package links

import (
	"context"
	"testing"

	"arcadium.dev/arcade/internal/arcade"
)

type (
	mockService struct {
		t   *testing.T
		err error

		linkID string
		req    linkRequest

		link  arcade.Link
		links []arcade.Link

		listCalled, getCalled, createCalled, updateCalled, removeCalled bool
	}
)

func (m *mockService) list(ctx context.Context) ([]arcade.Link, error) {
	m.listCalled = true
	if m.err != nil {
		return nil, m.err
	}
	return m.links, nil
}

func (m *mockService) get(ctx context.Context, linkID string) (arcade.Link, error) {
	m.getCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.linkID != linkID {
		m.t.Fatalf("get: expected linkID %s, actual linkID %s", m.linkID, linkID)
	}
	return m.link, nil
}

func (m *mockService) create(ctx context.Context, req linkRequest) (arcade.Link, error) {
	m.createCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.req != req {
		m.t.Fatalf("create: expected link request %+v, actual link requset %+v", m.req, req)
	}
	return m.link, nil
}

func (m *mockService) update(ctx context.Context, linkID string, req linkRequest) (arcade.Link, error) {
	m.updateCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.linkID != linkID {
		m.t.Fatalf("get: expected linkID %s, actual linkID %s", m.linkID, linkID)
	}
	if m.req != req {
		m.t.Fatalf("update: expected link request %+v, actual link requset %+v", m.req, req)
	}
	return m.link, nil
}

func (m *mockService) remove(ctx context.Context, linkID string) error {
	m.removeCalled = true
	if m.err != nil {
		return m.err
	}
	if m.linkID != linkID {
		m.t.Fatalf("remove: expected linkID %s, actual linkID %s", m.linkID, linkID)
	}
	return nil
}
