package session

import (
	"context"
	"net/http"
	"slices"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
)

type CookieReader func(ctx context.Context) ([]*http.Cookie, error)

// reads the session cookie from every browser on this machine, newest first
func ReadBrowserCookies(ctx context.Context) ([]*http.Cookie, error) {
	found, err := kooky.ReadCookies(ctx,
		kooky.Valid,
		kooky.DomainHasSuffix(Domain),
		kooky.Name(CookieName),
	)

	// some browsers' stores fail to read, that's fine as long as another one had the cookie
	if len(found) == 0 && err != nil {
		return nil, err
	}

	slices.SortFunc(found, func(a, b *kooky.Cookie) int {
		return b.Expires.Compare(a.Expires)
	})

	var cookies []*http.Cookie

	for _, cookie := range found {
		cookies = append(cookies, &http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	return cookies, nil
}
