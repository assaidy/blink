package components

import (
	"github.com/assaidy/gg"
)

func rootLayout(children ...any) gg.Node {
	return gg.Empty(
		gg.DoctypeHTML(),
		gg.Html(
			gg.Head(
				gg.Title("blink"),
				gg.Meta(gg.KV{"charset": "UTF-8"}),
				gg.Meta(gg.KV{"name": "viewport", "content": "width=device-width, initial-scale=1.0"}),
				gg.Script(gg.KV{"src": "/public/js/lib/htmx_2.0.7.js"}),
				gg.Script(gg.KV{"src": "/public/js/script.js", "defer": true}),
				gg.Link(gg.KV{"rel": "stylesheet", "href": "/public/css/style.css"}),
			),
			gg.Body(gg.KV{
				"hx-on::config-request": `
					event.detail.headers['X-CSRF-Token'] = document.cookie
						.split('; ')
						.find(row => row.startsWith('csrf_token='))?
						.split('=')[1]?
						.trim() || '';
				`,
				"class": "bg-bg-primary",
			},
				gg.Div(children...),
			),
		),
	)
}
