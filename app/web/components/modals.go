package components

import (
	"github.com/assaidy/h"
)

func SearchModal() h.Node {
	return h.Div(h.KV{
		"hx-on:click": "if (event.target === this) this.remove()",
		"id":          "search-modal",
		"class":       "fixed inset-0 z-50 flex items-start justify-center bg-black/50 backdrop-blur-sm p-4 pt-20 outline-none",
	},
		h.Div(h.KV{"class": "bg-bg-secondary rounded-2xl shadow-2xl w-full max-w-xl flex flex-col overflow-hidden"},
			// Header with search input
			h.Div(h.KV{"class": "flex items-center gap-3 p-4 border-b border-bg-tertiary"},
				h.Div(h.KV{"class": "flex-1 relative"},
					h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="absolute left-3 top-1/2 -translate-y-1/2 text-fg-secondary"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>`),
					h.Input(h.KV{
						"type":        "text",
						"placeholder": "Search users...",
						"autofocus":   "true",
						"name":        "query",
						"hx-get":      "/search/users",
						"hx-trigger":  "input changed delay:300ms",
						"hx-target":   "#search-results",
						"hx-swap":     "innerHTML",
						"class":       "w-full bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary pl-10 pr-4 py-3 outline-none transition-colors",
					}),
				),
				h.Button(h.KV{
					"hx-on:click": "this.closest('#search-modal').remove()",
					"class":       "p-2 hover:bg-bg-tertiary rounded-lg transition-colors cursor-pointer",
				},
					h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`),
				),
			),
			// Content area with results
			h.Div(h.KV{"id": "search-results", "class": "flex-1 p-4 overflow-y-auto max-h-96"},
				h.P(h.KV{"class": "text-fg-secondary text-center py-3"}, "Search for users by name or username"),
			),
		),
	)
}

type ProfileModalTab int

const (
	tabDefault ProfileModalTab = iota
	TabProfile
	TabSessions
)

type ProfileModalParams struct {
	ActiveTab         ProfileModalTab
	ProfileTabParams  ProfileTabParams
	SessionsTabParams SessionsTabParams
}

func ProfileModal(params ProfileModalParams) h.Node {
	if params.ActiveTab == tabDefault {
		params.ActiveTab = TabProfile
	}

	profileIcon := h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>`)
	sessionsIcon := h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>`)

	return h.Div(h.KV{
		"hx-on:click": "if (event.target === this) this.remove()",
		"id":          "profile-modal",
		"class":       "fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 outline-none",
	},
		h.Div(h.KV{"class": "bg-bg-secondary rounded-2xl shadow-2xl max-w-3xl w-full flex overflow-hidden", "style": "height: 80vh; max-height: 700px;"},
			// Left sidebar with tabs
			h.Div(h.KV{"class": "w-56 bg-bg-tertiary flex flex-col p-2 flex-shrink-0"},
				// Close button at top
				h.Button(h.KV{
					"hx-on:click": "this.closest('#profile-modal').remove()",
					"class":       "self-start p-2 hover:bg-bg-secondary rounded-lg transition-colors cursor-pointer mb-6",
				},
					h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`),
				),
				// Tabs
				h.Div(h.KV{"class": "flex flex-col gap-1"},
					h.Button(h.KV{
						"hx-get":     "/profile_modal?tab=profile",
						"hx-target":  "#profile-modal",
						"hx-swap":    "outerHTML",
						"hx-trigger": h.IfElse(params.ActiveTab == TabProfile, "none", "click"),
						"class":      "flex items-center gap-3 px-2 py-2.5 rounded-lg text-left font-medium text-fg-primary " + h.IfElse(params.ActiveTab == TabProfile, "bg-bg-secondary", "hover:bg-bg-secondary/50"),
					},
						profileIcon, "Profile",
					),
					h.Button(h.KV{
						"hx-get":     "/profile_modal?tab=sessions",
						"hx-target":  "#profile-modal",
						"hx-swap":    "outerHTML",
						"hx-trigger": h.IfElse(params.ActiveTab == TabSessions, "none", "click"),
						"class":      "flex items-center gap-3 px-2 py-2.5 rounded-lg text-left font-medium text-fg-primary " + h.IfElse(params.ActiveTab == TabSessions, "bg-bg-secondary", "hover:bg-bg-secondary/50"),
					},
						sessionsIcon, "Sessions",
					),
				),
			),
			// Right content area
			h.Div(h.KV{"class": "flex-1 flex flex-col min-w-0"},
				// Header
				h.Div(h.KV{"class": "px-8 py-6 border-b border-bg-tertiary"},
					h.H2(h.KV{"class": "text-xl font-semibold text-fg-primary"},
						h.IfElse(params.ActiveTab == TabProfile, "Profile", "Sessions"),
					),
				),
				// Content
				h.Div(h.KV{"id": "tab-content", "class": "flex-1 p-8 overflow-y-auto"},
					h.IfElse(params.ActiveTab == TabProfile,
						ProfileTab(params.ProfileTabParams),
						SessionsTab(params.SessionsTabParams),
					),
				),
			),
		),
	)
}
