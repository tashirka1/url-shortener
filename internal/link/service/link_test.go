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
	createLinkFunc    func(ctx context.Context, url string, userId int) (model.Link, error)
	listLinkFunc      func(ctx context.Context, userId, cursor int) ([]model.Link, error)
	removeLinkFunc    func(ctx context.Context, userId int, code string) error
	searchLinkFunc    func(ctx context.Context, userId int, query string) ([]model.Link, error)
	getAndClickFunc   func(ctx context.Context, code string) (string, error)
	getLinkByCodeFunc func(ctx context.Context, code string, userId int) (model.Link, error)
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

func (m *mockLinkStorage) SearchLink(ctx context.Context, userId int, query string) ([]model.Link, error) {
	if m.searchLinkFunc != nil {
		return m.searchLinkFunc(ctx, userId, query)
	}
	return nil, errors.New("SearchLink not implemented")
}

func (m *mockLinkStorage) GetAndClick(ctx context.Context, code string) (string, error) {
	if m.getAndClickFunc != nil {
		return m.getAndClickFunc(ctx, code)
	}
	return "", errors.New("GetAndClick not implemented")
}

func (m *mockLinkStorage) GetLinkByCode(ctx context.Context, code string, userId int) (model.Link, error) {
	if m.getLinkByCodeFunc != nil {
		return m.getLinkByCodeFunc(ctx, code, userId)
	}
	return model.Link{}, errors.New("GetLinkByCode not implemented")
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

func TestSearchLink_Success(t *testing.T) {
	expected := []model.Link{
		{Id: 1, Code: "abc", Url: "https://example.com"},
	}
	mock := &mockLinkStorage{
		searchLinkFunc: func(ctx context.Context, userId int, query string) ([]model.Link, error) {
			assert.Equal(t, 1, userId)
			assert.Equal(t, "example", query)
			return expected, nil
		},
	}
	s := NewLink(mock)

	links, err := s.SearchLink(context.Background(), 1, "example")

	assert.NoError(t, err)
	assert.Equal(t, expected, links)
}

func TestSearchLink_NoResults(t *testing.T) {
	mock := &mockLinkStorage{
		searchLinkFunc: func(ctx context.Context, userId int, query string) ([]model.Link, error) {
			return []model.Link{}, nil
		},
	}
	s := NewLink(mock)

	links, err := s.SearchLink(context.Background(), 1, "nonexistent")

	assert.NoError(t, err)
	assert.Empty(t, links)
}

func TestSearchLink_StorageError(t *testing.T) {
	mock := &mockLinkStorage{
		searchLinkFunc: func(ctx context.Context, userId int, query string) ([]model.Link, error) {
			return nil, errors.New("db error")
		},
	}
	s := NewLink(mock)

	_, err := s.SearchLink(context.Background(), 1, "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
