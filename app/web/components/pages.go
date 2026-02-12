package components

import (
	"github.com/assaidy/h"
)

func RegisterPage() h.Node {
	return rootLayout(
		h.Div(h.KV{"class": "min-h-screen flex justify-center items-center bg-bg-primary sm:px-6 lg:px-8"},
			h.Div(h.KV{"class": "w-full min-h-screen sm:min-h-0 sm:max-w-md sm:p-8 bg-bg-secondary sm:rounded-lg sm:shadow-lg flex flex-col justify-center"},
				h.Div(h.KV{"class": "p-6 sm:p-0"},
					h.H2(h.KV{"class": "text-fg-primary text-2xl font-bold text-center mb-8"}, "Create Account"),
					RegisterForm(),
					h.P(h.KV{"class": "text-center text-fg-secondary text-sm mt-6"},
						"Already have an account? ", h.A(h.KV{"href": "/login", "class": "text-blue hover:underline font-medium"}, "Sign in"),
					),
				),
			),
		),
	)
}

func LoginPage() h.Node {
	return rootLayout(
		h.Div(h.KV{"class": "min-h-screen flex justify-center items-center bg-bg-primary sm:px-6 lg:px-8"},
			h.Div(h.KV{"class": "w-full min-h-screen sm:min-h-0 sm:max-w-md sm:p-8 bg-bg-secondary sm:rounded-lg sm:shadow-lg flex flex-col justify-center"},
				h.Div(h.KV{"class": "p-6 sm:p-0"},
					h.H2(h.KV{"class": "text-fg-primary text-2xl font-bold text-center mb-8"}, "Sign In"),
					LoginForm(),
					h.P(h.KV{"class": "text-center text-fg-secondary text-sm mt-6"},
						"Don't have an account? ", h.A(h.KV{"href": "/register", "class": "text-blue hover:underline font-medium"}, "Create one"),
					),
				),
			),
		),
	)
}

type UserBlockParams struct {
	Name            string
	Username        string
	ProfileImageUrl string
}

type ChatPageParams struct {
	User UserBlockParams
}

func ChatPage(params ChatPageParams) h.Node {
	return rootLayout(
		h.Div(h.KV{"class": "h-screen flex bg-bg-primary"},
			// Sidebar
			h.Div(h.KV{"class": "w-80 bg-bg-secondary border-r border-bg-tertiary flex flex-col"},
				// Sticky top row
				h.Div(h.KV{"class": "sticky top-0 z-10 bg-aqua/10 border-b border-bg-primary flex items-center"},
					h.Div(h.KV{
						"hx-get":     "/profile_modal",
						"hx-trigger": "click",
						"hx-target":  "body",
						"hx-swap":    "beforeend",
						"class":      "flex-1 flex items-center gap-3 cursor-pointer hover:bg-aqua/20 transition-colors px-4 py-3",
					},
						h.Div(h.KV{"class": "w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold flex-shrink-0"},
							getInitials(params.User.Name),
						),
						h.Div(h.KV{"class": "flex-1 min-w-0"},
							h.P(h.KV{"class": "text-fg-primary font-medium truncate"}, params.User.Name),
							h.P(h.KV{"class": "text-fg-secondary text-sm truncate"}, "@"+params.User.Username),
						),
					),
					h.Button(h.KV{
						"hx-get":     "/search_modal",
						"hx-trigger": "click",
						"hx-target":  "body",
						"hx-swap":    "beforeend",
						"class":      "flex items-center justify-center aspect-square h-full hover:bg-aqua/30 transition-colors cursor-pointer",
					},
						h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>`),
					),
				),
				// Partners list container
				h.Div(h.KV{
					"id":    "partners-container",
					"class": "flex-1 overflow-y-auto px-2 py-2",
				},
					h.Div(h.KV{
						"hx-get":              "/partners",
						"hx-trigger":          "load",
						"hx-swap":             "afterend",
						"hx-indicator":        "#partners-indicator",
						"hx-on::after-settle": "document.getElementById('partners-container').scrollTop = 0; this.remove()",
					}),
					h.Div(h.KV{"class": "flex justify-center"},
						spinner("partners-indicator"),
					),
				),
			),
			// Current chat area (placeholder)
			h.Div(h.KV{
				"id":    "chat-container",
				"class": "flex-1 flex flex-col bg-bg-primary",
			},
				h.Div(h.KV{"class": "flex-1 flex items-center justify-center"},
					h.P(h.KV{"class": "text-fg-secondary text-lg"}, "Select a chat to start messaging"),
				),
			),
		),
	)
}
