package components

import (
	"github.com/assaidy/hyper"
)

func rootLayout(children ...any) h.Node {
	return h.Empty(
		h.DoctypeHTML(),
		h.Html(
			h.Head(
				h.Title("blink"),
				h.Meta(h.KV{"charset": "UTF-8"}),
				h.Meta(h.KV{"name": "viewport", "content": "width=device-width, initial-scale=1.0"}),
				h.Script(h.KV{"src": "/public/js/lib/htmx_2.0.7.js"}),
				h.Script(h.KV{"src": "/public/js/script.js", "defer": true}),
				h.Link(h.KV{"rel": "stylesheet", "href": "/public/css/style.css"}),
			),
			h.Body(h.KV{
				"hx-on::config-request": `
					event.detail.headers['X-CSRF-Token'] = document.cookie
						.split('; ')
						.find(row => row.startsWith('csrf_token='))
						?.split('=')[1]
						?.trim() || '';
				`,
				"class": "bg-bg-primary text-fg-primary",
			},
				h.Div(children...),
			),
		),
	)
}
