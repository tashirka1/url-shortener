package service

import (
	"context"
	"errors"
	"testing"

	"url_shortener/internal/link/model"
	"url_shortener/internal/link/storage"

	"github.com/stretchr/testify/assert"
)

type mockLinkStorage struct {
	createLinkFunc func(ctx context.Context, url string, userId int) (model.Link, error)
	listLinkFunc   func(ctx context.Context, userId, cursor int) ([]model.Link, error)
	removeLinkFunc func(ctx context.Context, userId int, code string) error
	getLinkFunc    func(ctx context.Context, code string) (string, error)
	clickLinkFunc  func(ctx context.Context, code string) error
}

func (m *mockLinkStorage) CreateLink(ctx context.Context, url string, userId int) (model.Link, error) {
	if m.createLinkFunc != nil {
		return m.createLinkFunc(ctx, url, userId)
	}
	return model.Link{}, errors.New("CreateLink not implemented")
}

func (m *mockLinkStorage) ListLink(ctx context.Context, userId, cursor int) ([]model.Link, error) {
	if m.listLinkFunc != nil {
		return m.listLinkFunc(ctx, userId, cursor)
	}
	return nil, errors.New("ListLink not implemented")
}

func (m *mockLinkStorage) RemoveLink(ctx context.Context, userId int, code string) error {
	if m.removeLinkFunc != nil {
		return m.removeLinkFunc(ctx, userId, code)
	}
	return errors.New("RemoveLink not implemented")
}

func (m *mockLinkStorage) GetLink(ctx context.Context, code string) (string, error) {
	if m.getLinkFunc != nil {
		return m.getLinkFunc(ctx, code)
	}
	return "", errors.New("GetLink not implemented")
}

func (m *mockLinkStorage) ClickLink(ctx context.Context, code string) error {
	if m.clickLinkFunc != nil {
		return m.clickLinkFunc(ctx, code)
	}
	return errors.New("ClickLink not implemented")
}

var _ storage.LinkStorage = (*mockLinkStorage)(nil)

func TestCreateLink_Success(t *testing.T) {
	mock := &mockLinkStorage{
		createLinkFunc: func(ctx context.Context, url string, userId int) (model.Link, error) {
			return model.Link{Id: 1, Code: "abc123", Url: url}, nil
		},
	}
	s := NewLink(mock)

	link, err := s.CreateLink(context.Background(), "https://example.com", 1)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), link.Id)
	assert.Equal(t, "abc123", link.Code)
	assert.Equal(t, "https://example.com", link.Url)
}

func TestCreateLink_StorageError(t *testing.T) {
	mock := &mockLinkStorage{
		createLinkFunc: func(ctx context.Context, url string, userId int) (model.Link, error) {
			return model.Link{}, errors.New("db error")
		},
	}
	s := NewLink(mock)

	_, err := s.CreateLink(context.Background(), "https://example.com", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestListLink_Success(t *testing.T) {
	expected := []model.Link{
		{Id: 2, Code: "def", Url: "https://a.com"},
		{Id: 1, Code: "abc", Url: "https://b.com"},
	}
	mock := &mockLinkStorage{
		listLinkFunc: func(ctx context.Context, userId, cursor int) ([]model.Link, error) {
			assert.Equal(t, 1, userId)
			assert.Equal(t, 999, cursor)
			return expected, nil
		},
	}
	s := NewLink(mock)

	links, err := s.ListLink(context.Background(), 1, 999)

	assert.NoError(t, err)
	assert.Equal(t, expected, links)
}

func TestListLink_Empty(t *testing.T) {
	mock := &mockLinkStorage{
		listLinkFunc: func(ctx context.Context, userId, cursor int) ([]model.Link, error) {
			return []model.Link{}, nil
		},
	}
	s := NewLink(mock)

	links, err := s.ListLink(context.Background(), 1, 0)

	assert.NoError(t, err)
	assert.Empty(t, links)
}

func TestRemoveLink_Success(t *testing.T) {
	mock := &mockLinkStorage{
		removeLinkFunc: func(ctx context.Context, userId int, code string) error {
			assert.Equal(t, 1, userId)
			assert.Equal(t, "abc", code)
			return nil
		},
	}
	s := NewLink(mock)

	err := s.RemoveLink(context.Background(), 1, "abc")

	assert.NoError(t, err)
}

func TestRemoveLink_NotFound(t *testing.T) {
	mock := &mockLinkStorage{
		removeLinkFunc: func(ctx context.Context, userId int, code string) error {
			return errors.New("no rows in result set")
		},
	}
	s := NewLink(mock)

	err := s.RemoveLink(context.Background(), 1, "missing")

	assert.Error(t, err)
}

func TestGetLink_Success(t *testing.T) {
	mock := &mockLinkStorage{
		getLinkFunc: func(ctx context.Context, code string) (string, error) {
			assert.Equal(t, "abc", code)
			return "https://example.com", nil
		},
	}
	s := NewLink(mock)

	url, err := s.GetLink(context.Background(), "abc")

	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", url)
}

func TestGetLink_NotFound(t *testing.T) {
	mock := &mockLinkStorage{
		getLinkFunc: func(ctx context.Context, code string) (string, error) {
			return "", errors.New("no rows in result set")
		},
	}
	s := NewLink(mock)

	_, err := s.GetLink(context.Background(), "missing")

	assert.Error(t, err)
}

func TestClickLink_Success(t *testing.T) {
	called := false
	mock := &mockLinkStorage{
		clickLinkFunc: func(ctx context.Context, code string) error {
			assert.Equal(t, "abc", code)
			called = true
			return nil
		},
	}
	s := NewLink(mock)

	err := s.ClickLink(context.Background(), "abc")

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestClickLink_StorageError(t *testing.T) {
	mock := &mockLinkStorage{
		clickLinkFunc: func(ctx context.Context, code string) error {
			return errors.New("db error")
		},
	}
	s := NewLink(mock)

	err := s.ClickLink(context.Background(), "abc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
