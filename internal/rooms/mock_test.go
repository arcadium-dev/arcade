package rooms

import (
	"context"
	"testing"

	"arcadium.dev/arcade/internal/arcade"
)

type (
	mockService struct {
		t   *testing.T
		err error

		roomID string
		req    roomRequest

		room  arcade.Room
		rooms []arcade.Room

		listCalled, getCalled, createCalled, updateCalled, removeCalled bool
	}
)

func (m *mockService) list(ctx context.Context) ([]arcade.Room, error) {
	m.listCalled = true
	if m.err != nil {
		return nil, m.err
	}
	return m.rooms, nil
}

func (m *mockService) get(ctx context.Context, roomID string) (arcade.Room, error) {
	m.getCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.roomID != roomID {
		m.t.Fatalf("get: expected roomID %s, actual roomID %s", m.roomID, roomID)
	}
	return m.room, nil
}

func (m *mockService) create(ctx context.Context, req roomRequest) (arcade.Room, error) {
	m.createCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.req != req {
		m.t.Fatalf("create: expected room request %+v, actual room requset %+v", m.req, req)
	}
	return m.room, nil
}

func (m *mockService) update(ctx context.Context, roomID string, req roomRequest) (arcade.Room, error) {
	m.updateCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.roomID != roomID {
		m.t.Fatalf("get: expected roomID %s, actual roomID %s", m.roomID, roomID)
	}
	if m.req != req {
		m.t.Fatalf("update: expected room request %+v, actual room requset %+v", m.req, req)
	}
	return m.room, nil
}

func (m *mockService) remove(ctx context.Context, roomID string) error {
	m.removeCalled = true
	if m.err != nil {
		return m.err
	}
	if m.roomID != roomID {
		m.t.Fatalf("remove: expected roomID %s, actual roomID %s", m.roomID, roomID)
	}
	return nil
}
