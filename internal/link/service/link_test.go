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
	getAndClickFunc   func(ctx context.Context, code, referrer, userAgent string) (model.Link, error)
	getLinkByCodeFunc func(ctx context.Context, code string, userId int) (model.Link, error)
	updateAliasFunc   func(ctx context.Context, userID int, currentCode, newCode string) error
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

func (m *mockLinkStorage) GetAndClick(ctx context.Context, code, referrer, userAgent string) (model.Link, error) {
	if m.getAndClickFunc != nil {
		return m.getAndClickFunc(ctx, code, referrer, userAgent)
	}
	return model.Link{}, errors.New("GetAndClick not implemented")
}

func (m *mockLinkStorage) GetLinkByCode(ctx context.Context, code string, userId int) (model.Link, error) {
	if m.getLinkByCodeFunc != nil {
		return m.getLinkByCodeFunc(ctx, code, userId)
	}
	return model.Link{}, errors.New("GetLinkByCode not implemented")
}

func (m *mockLinkStorage) UpdateAlias(ctx context.Context, userID int, currentCode, newCode string) error {
	if m.updateAliasFunc != nil {
		return m.updateAliasFunc(ctx, userID, currentCode, newCode)
	}
	return errors.New("UpdateAlias not implemented")
}

type mockClickStorage struct {
	getDailyClicksFunc  func(ctx context.Context, linkID int64, days int) ([]model.DailyClick, error)
	getTopReferrersFunc func(ctx context.Context, linkID int64, limit int) ([]model.ReferrerStat, error)
}

func (m *mockClickStorage) GetDailyClicks(ctx context.Context, linkID int64, days int) ([]model.DailyClick, error) {
	if m.getDailyClicksFunc != nil {
		return m.getDailyClicksFunc(ctx, linkID, days)
	}
	return nil, errors.New("GetDailyClicks not implemented")
}

func (m *mockClickStorage) GetTopReferrers(ctx context.Context, linkID int64, limit int) ([]model.ReferrerStat, error) {
	if m.getTopReferrersFunc != nil {
		return m.getTopReferrersFunc(ctx, linkID, limit)
	}
	return nil, errors.New("GetTopReferrers not implemented")
}

var _ storage.LinkStorage = (*mockLinkStorage)(nil)
var _ storage.ClickStorage = (*mockClickStorage)(nil)

func TestCreateLink_Success(t *testing.T) {
	mock := &mockLinkStorage{
		createLinkFunc: func(ctx context.Context, url string, userId int) (model.Link, error) {
			return model.Link{Id: 1, Code: "abc123", Url: url}, nil
		},
	}
	s := NewLink(mock, &mockClickStorage{})

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
	s := NewLink(mock, &mockClickStorage{})

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
	s := NewLink(mock, &mockClickStorage{})

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
	s := NewLink(mock, &mockClickStorage{})

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
	s := NewLink(mock, &mockClickStorage{})

	err := s.RemoveLink(context.Background(), 1, "abc")

	assert.NoError(t, err)
}

func TestRemoveLink_NotFound(t *testing.T) {
	mock := &mockLinkStorage{
		removeLinkFunc: func(ctx context.Context, userId int, code string) error {
			return errors.New("no rows in result set")
		},
	}
	s := NewLink(mock, &mockClickStorage{})

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
	s := NewLink(mock, &mockClickStorage{})

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
	s := NewLink(mock, &mockClickStorage{})

	links, err := s.SearchLink(context.Background(), 1, "nonexistent")

	assert.NoError(t, err)
	assert.Empty(t, links)
}

func TestUpdateAlias_SameCode(t *testing.T) {
	mock := &mockLinkStorage{
		updateAliasFunc: func(ctx context.Context, userID int, currentCode, newCode string) error {
			return nil
		},
	}
	s := NewLink(mock, &mockClickStorage{})

	err := s.UpdateAlias(context.Background(), 1, "abc", "abc")

	assert.NoError(t, err)
}

func TestUpdateAlias_Success(t *testing.T) {
	mock := &mockLinkStorage{
		updateAliasFunc: func(ctx context.Context, userID int, currentCode, newCode string) error {
			assert.Equal(t, "my-link", newCode)
			return nil
		},
	}
	s := NewLink(mock, &mockClickStorage{})

	err := s.UpdateAlias(context.Background(), 1, "abc123", "my-link")

	assert.NoError(t, err)
}

func TestUpdateAlias_TooShort(t *testing.T) {
	s := NewLink(nil, nil)

	err := s.UpdateAlias(context.Background(), 1, "abc", "ab")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "от 3 до 32")
}

func TestUpdateAlias_TooLong(t *testing.T) {
	s := NewLink(nil, nil)

	long := string(make([]byte, 33))
	for i := range long {
		long = long[:i] + "a" + long[i+1:]
	}
	err := s.UpdateAlias(context.Background(), 1, "abc", long)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "от 3 до 32")
}

func TestUpdateAlias_InvalidChars(t *testing.T) {
	s := NewLink(nil, nil)

	err := s.UpdateAlias(context.Background(), 1, "abc", "my link!")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "только латинские буквы")
}

func TestUpdateAlias_ReservedWord(t *testing.T) {
	s := NewLink(nil, nil)

	err := s.UpdateAlias(context.Background(), 1, "abc", "health")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "зарезервирован")
}

func TestUpdateAlias_AliasTaken(t *testing.T) {
	mock := &mockLinkStorage{
		updateAliasFunc: func(ctx context.Context, userID int, currentCode, newCode string) error {
			return model.ErrAliasTaken
		},
	}
	s := NewLink(mock, &mockClickStorage{})

	err := s.UpdateAlias(context.Background(), 1, "abc", "taken")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alias already taken")
}

func TestUpdateAlias_NotFound(t *testing.T) {
	mock := &mockLinkStorage{
		updateAliasFunc: func(ctx context.Context, userID int, currentCode, newCode string) error {
			return errors.New("not found")
		},
	}
	s := NewLink(mock, &mockClickStorage{})

	err := s.UpdateAlias(context.Background(), 1, "missing", "new-code")

	assert.Error(t, err)
}

func TestSearchLink_StorageError(t *testing.T) {
	mock := &mockLinkStorage{
		searchLinkFunc: func(ctx context.Context, userId int, query string) ([]model.Link, error) {
			return nil, errors.New("db error")
		},
	}
	s := NewLink(mock, &mockClickStorage{})

	_, err := s.SearchLink(context.Background(), 1, "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestGetAndClick_Success(t *testing.T) {
	mock := &mockLinkStorage{
		getAndClickFunc: func(ctx context.Context, code, referrer, userAgent string) (model.Link, error) {
			assert.Equal(t, "abc", code)
			assert.Equal(t, "https://google.com", referrer)
			assert.Equal(t, "Mozilla/5.0", userAgent)
			return model.Link{Id: 1, Url: "https://example.com", Code: "abc"}, nil
		},
	}
	s := NewLink(mock, &mockClickStorage{})

	url, err := s.GetAndClick(context.Background(), "abc", "https://google.com", "Mozilla/5.0")

	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", url)
}

func TestGetAndClick_NotFound(t *testing.T) {
	mock := &mockLinkStorage{
		getAndClickFunc: func(ctx context.Context, code, referrer, userAgent string) (model.Link, error) {
			return model.Link{}, errors.New("not found")
		},
	}
	s := NewLink(mock, &mockClickStorage{})

	_, err := s.GetAndClick(context.Background(), "missing", "", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
