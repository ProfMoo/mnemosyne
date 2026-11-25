package ui_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/maxence-charriere/go-app/v11/pkg/app"

	"github.com/profmoo/mnemosyne/internal/ui"
)

var errUnavailable = errors.New("service unavailable")

func TestListensRender(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		want string
		page ui.Listens
	}{
		{name: "heading", want: "Response from GET /listens"},
		{name: "loading", page: ui.Listens{Loading: true}, want: "Loading listens"},
		{
			name: "empty",
			page: ui.Listens{Result: ui.ListenResult{JSON: emptyListens, Empty: true}},
			want: "No listens yet.",
		},
		{name: "failure", page: ui.Listens{Failure: errUnavailable}, want: "Try again"},
		{
			name: "escaped response",
			page: ui.Listens{Result: ui.ListenResult{JSON: "<script>alert(1)</script>"}},
			want: "&lt;script&gt;",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			html := app.HTMLString(tc.page.Render())
			if !strings.Contains(html, tc.want) {
				t.Errorf("rendered HTML %q does not contain %q", html, tc.want)
			}
		})
	}
}
