//go:build unit

package service

import (
	"context"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type proxyRepoValidationStub struct {
	nextID int64
	items  map[int64]*Proxy
}

func newProxyRepoValidationStub(items ...*Proxy) *proxyRepoValidationStub {
	stub := &proxyRepoValidationStub{nextID: 1000, items: make(map[int64]*Proxy, len(items))}
	for _, item := range items {
		cloned := *item
		stub.items[item.ID] = &cloned
		if item.ID >= stub.nextID {
			stub.nextID = item.ID + 1
		}
	}
	return stub
}

func (s *proxyRepoValidationStub) Create(ctx context.Context, proxy *Proxy) error {
	if proxy.ID == 0 {
		proxy.ID = s.nextID
		s.nextID++
	}
	cloned := *proxy
	s.items[proxy.ID] = &cloned
	return nil
}

func (s *proxyRepoValidationStub) GetByID(ctx context.Context, id int64) (*Proxy, error) {
	item, ok := s.items[id]
	if !ok {
		return nil, ErrProxyNotFound
	}
	cloned := *item
	return &cloned, nil
}

func (s *proxyRepoValidationStub) ListByIDs(ctx context.Context, ids []int64) ([]Proxy, error) {
	out := make([]Proxy, 0, len(ids))
	for _, id := range ids {
		if item, ok := s.items[id]; ok {
			out = append(out, *item)
		}
	}
	return out, nil
}

func (s *proxyRepoValidationStub) Update(ctx context.Context, proxy *Proxy) error {
	if _, ok := s.items[proxy.ID]; !ok {
		return ErrProxyNotFound
	}
	cloned := *proxy
	s.items[proxy.ID] = &cloned
	return nil
}

func (s *proxyRepoValidationStub) Delete(ctx context.Context, id int64) error {
	delete(s.items, id)
	return nil
}

func (s *proxyRepoValidationStub) List(ctx context.Context, params pagination.PaginationParams) ([]Proxy, *pagination.PaginationResult, error) {
	all := make([]Proxy, 0, len(s.items))
	for _, item := range s.items {
		all = append(all, *item)
	}

	offset := params.Offset()
	if offset > len(all) {
		offset = len(all)
	}
	limit := params.Limit()
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}

	paged := append([]Proxy(nil), all[offset:end]...)
	return paged, &pagination.PaginationResult{Total: int64(len(all))}, nil
}

func (s *proxyRepoValidationStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]Proxy, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *proxyRepoValidationStub) ListWithFiltersAndAccountCount(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]ProxyWithAccountCount, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFiltersAndAccountCount call")
}

func (s *proxyRepoValidationStub) ListActive(ctx context.Context) ([]Proxy, error) {
	panic("unexpected ListActive call")
}

func (s *proxyRepoValidationStub) ListActiveWithAccountCount(ctx context.Context) ([]ProxyWithAccountCount, error) {
	panic("unexpected ListActiveWithAccountCount call")
}

func (s *proxyRepoValidationStub) ExistsByHostPortAuth(ctx context.Context, host string, port int, username, password string) (bool, error) {
	panic("unexpected ExistsByHostPortAuth call")
}

func (s *proxyRepoValidationStub) CountAccountsByProxyID(ctx context.Context, proxyID int64) (int64, error) {
	panic("unexpected CountAccountsByProxyID call")
}

func (s *proxyRepoValidationStub) ListAccountSummariesByProxyID(ctx context.Context, proxyID int64) ([]ProxyAccountSummary, error) {
	panic("unexpected ListAccountSummariesByProxyID call")
}

func TestAdminServiceProxyCreateNormalizesAndChecksDuplicate(t *testing.T) {
	repo := newProxyRepoValidationStub(&Proxy{ID: 1, Name: "existing", Protocol: "http", Host: "proxy.example.com", Port: 8080, Username: "u", Password: "p", Status: StatusActive})
	svc := &adminServiceImpl{proxyRepo: repo}

	_, err := svc.CreateProxy(context.Background(), &CreateProxyInput{
		Name:     "dup",
		Protocol: " HTTP ",
		Host:     " Proxy.Example.com ",
		Port:     8080,
		Username: " u ",
		Password: " p ",
	})
	require.ErrorIs(t, err, ErrProxyDuplicate)

	created, err := svc.CreateProxy(context.Background(), &CreateProxyInput{
		Name:     "new",
		Protocol: "HTTPS",
		Host:     " Example.COM ",
		Port:     443,
		Username: " user ",
		Password: " pass ",
	})
	require.NoError(t, err)
	require.Equal(t, "https", created.Protocol)
	require.Equal(t, "example.com", created.Host)
	require.Equal(t, "user", created.Username)
	require.Equal(t, "pass", created.Password)
}

func TestAdminServiceProxyUpdateValidatesAndChecksDuplicate(t *testing.T) {
	repo := newProxyRepoValidationStub(
		&Proxy{ID: 1, Name: "p1", Protocol: "http", Host: "proxy-a.local", Port: 8080, Username: "u1", Password: "p1", Status: StatusActive},
		&Proxy{ID: 2, Name: "p2", Protocol: "https", Host: "proxy-b.local", Port: 443, Username: "u2", Password: "p2", Status: StatusActive},
	)
	svc := &adminServiceImpl{proxyRepo: repo}

	_, err := svc.UpdateProxy(context.Background(), 2, &UpdateProxyInput{Port: 70000})
	require.Error(t, err)
	require.True(t, infraerrors.IsBadRequest(err))
	require.Equal(t, "PROXY_INVALID_PORT", infraerrors.Reason(err))

	_, err = svc.UpdateProxy(context.Background(), 2, &UpdateProxyInput{
		Protocol: "http",
		Host:     "proxy-a.local",
		Port:     8080,
		Username: "u1",
		Password: "p1",
	})
	require.ErrorIs(t, err, ErrProxyDuplicate)
}
