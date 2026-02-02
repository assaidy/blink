package components

import (
	"strings"
	"time"

	"github.com/assaidy/gg"
)

type ChatPageParams struct {
	User         UserBlockParams
	ChatPartners []UserBlockParams
}

func ChatPage(params ChatPageParams) gg.Node {
	return rootLayout(
		gg.Div(gg.KV{"class": "h-screen flex bg-bg-primary"},
			// Sidebar
			gg.Div(gg.KV{"class": "w-80 bg-bg-secondary border-r border-bg-tertiary flex flex-col"},
				// Sticky top row
				gg.Div(gg.KV{"class": "sticky top-0 bg-bg-secondary border-b border-bg-tertiary py-2 px-3 flex items-center gap-3 z-10"},
					gg.Div(gg.KV{
						"hx-get":     "/profile_modal",
						"hx-trigger": "click",
						"hx-target":  "body",
						"hx-swap":    "beforeend",
						"class":      "flex-1 min-w-0",
					},
						userBlock(params.User),
					),
					gg.Button(gg.KV{"class": "p-3 hover:bg-bg-tertiary rounded-lg transition-colors flex-shrink-0 cursor-pointer"},
						gg.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>`),
					),
				),
				// Partners list
				gg.Div(gg.KV{"class": "flex-1 overflow-y-auto"},
					gg.MapSlice(params.ChatPartners, func(partner UserBlockParams) gg.Node {
						return userBlock(partner)
					}),
				),
			),
			// Current chat area (placeholder)
			gg.Div(gg.KV{"class": "flex-1 flex flex-col bg-bg-primary"},
				gg.Div(gg.KV{"class": "flex-1 flex items-center justify-center"},
					gg.P(gg.KV{"class": "text-fg-secondary text-lg"}, "Select a chat to start messaging"),
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

func userBlock(params UserBlockParams) gg.Node {
	return gg.Div(gg.KV{"class": "flex items-center gap-3 p-3 hover:bg-bg-tertiary/50 hover:rounded-lg cursor-pointer transition-colors"},
		gg.IfElse(params.ProfileImageUrl != "",
			gg.Img(gg.KV{
				"src":   params.ProfileImageUrl,
				"alt":   params.Name,
				"class": "w-10 h-10 rounded-full object-cover",
			}),
			gg.Div(gg.KV{"class": "w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
				getInitials(params.Name),
			),
		),
		gg.Div(gg.KV{"class": "flex-1 min-w-0"},
			gg.P(gg.KV{"class": "text-fg-primary font-medium truncate"}, params.Name),
			gg.P(gg.KV{"class": "text-fg-secondary text-sm truncate"}, "@"+params.Username),
		),
	)
}

func getInitials(name string) string {
	words := strings.Fields(name)
	if len(words) == 0 {
		return "?"
	}
	if len(words) == 1 {
		return strings.ToUpper(string(words[0][0]))
	}
	return strings.ToUpper(string(words[0][0]) + string(words[1][0]))
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

func ProfileModal(params ProfileModalParams) gg.Node {
	if params.ActiveTab == tabDefault {
		params.ActiveTab = TabProfile
	}

	profileIcon := gg.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>`)
	sessionsIcon := gg.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>`)

	return gg.Div(gg.KV{
		"hx-on:click": "if (event.target === this) this.remove()",
		"id":          "profile-modal",
		"class":       "fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4",
	},
		gg.Div(gg.KV{"class": "bg-bg-secondary rounded-2xl shadow-2xl max-w-3xl w-full flex overflow-hidden", "style": "height: 80vh; max-height: 700px;"},
			// Left sidebar with tabs
			gg.Div(gg.KV{"class": "w-56 bg-bg-tertiary flex flex-col p-2 flex-shrink-0"},
				// Close button at top
				gg.Button(gg.KV{
					"hx-on:click": "this.closest('#profile-modal').remove()",
					"class":       "self-start p-2 hover:bg-bg-secondary rounded-lg transition-colors cursor-pointer mb-6",
				},
					gg.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`),
				),
				// Tabs
				gg.Div(gg.KV{"class": "flex flex-col gap-1"},
					gg.Button(gg.KV{
						"hx-get":     "/profile_modal?tab=profile",
						"hx-target":  "#profile-modal",
						"hx-swap":    "outerHTML",
						"hx-trigger": gg.IfElse(params.ActiveTab == TabProfile, "none", "click"),
						"class":      "flex items-center gap-3 px-2 py-2.5 rounded-lg text-left font-medium text-fg-primary " + gg.IfElse(params.ActiveTab == TabProfile, "bg-bg-secondary", "hover:bg-bg-secondary/50"),
					},
						profileIcon, "Profile",
					),
					gg.Button(gg.KV{
						"hx-get":     "/profile_modal?tab=sessions",
						"hx-target":  "#profile-modal",
						"hx-swap":    "outerHTML",
						"hx-trigger": gg.IfElse(params.ActiveTab == TabSessions, "none", "click"),
						"class":      "flex items-center gap-3 px-2 py-2.5 rounded-lg text-left font-medium text-fg-primary " + gg.IfElse(params.ActiveTab == TabSessions, "bg-bg-secondary", "hover:bg-bg-secondary/50"),
					},
						sessionsIcon, "Sessions",
					),
				),
			),
			// Right content area
			gg.Div(gg.KV{"class": "flex-1 flex flex-col min-w-0"},
				// Header
				gg.Div(gg.KV{"class": "px-8 py-6 border-b border-bg-tertiary"},
					gg.H2(gg.KV{"class": "text-xl font-semibold text-fg-primary"},
						gg.IfElse(params.ActiveTab == TabProfile, "Profile", "Sessions"),
					),
				),
				// Content
				gg.Div(gg.KV{"id": "tab-content", "class": "flex-1 p-8 overflow-y-auto"},
					gg.IfElse(params.ActiveTab == TabProfile,
						ProfileTab(params.ProfileTabParams),
						SessionsTab(params.SessionsTabParams),
					),
				),
			),
		),
	)
}

type ProfileTabParams struct {
	Name            string
	Username        string
	Email           string
	EmailIsVerified bool
	Bio             string
	JoinedAt        time.Time
}

func ProfileTab(params ProfileTabParams) gg.Node {
	return gg.Div(
		ProfileForm(ProfileFormParams{
			Name:     params.Name,
			Username: params.Username,
			Email:    params.Email,
			Bio:      params.Bio,
		}),
		gg.P(gg.KV{"class": "text-fg-secondary text-sm text-center my-4"},
			"Joined ", params.JoinedAt.Format("January 2, 2006"),
		),
		gg.Div(gg.KV{"class": "flex items-center justify-center gap-2 text-sm"},
			gg.IfElse(params.EmailIsVerified,
				gg.Empty(
					gg.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-green-500"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>`),
					gg.P(gg.KV{"class": "text-green-600"}, "Email verified"),
				),
				gg.Empty(
					gg.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-yellow-500"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>`),
					gg.P(gg.KV{"class": "text-yellow-600"}, "Email not verified"),
				),
			),
		),
	)
}

type ProfileFormParams struct {
	Name        string
	NameErr     any
	Username    string
	UsernameErr any
	Email       string
	EmailErr    any
	Bio         string
	BioErr      any
}

func ProfileForm(params ProfileFormParams) gg.Node {
	return gg.Form(gg.KV{
		"hx-put":          "/profile",
		"hx-swap":         "outerHTML",
		"hx-disabled-elt": "find button",
		"hx-indicator":    "#spinner",
		"class":           "space-y-5",
	},
		gg.Div(gg.KV{"class": "space-y-1"},
			gg.Label(gg.KV{"for": "name", "class": "block text-sm font-medium text-fg-secondary"}, "Full Name"),
			gg.Input(gg.KV{"type": "text", "id": "name", "name": "name", "required": true, "value": params.Name, "placeholder": "Enter your full name", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			gg.If(params.NameErr != nil, gg.P(gg.KV{"class": "text-red-500 text-sm mt-1"}, params.NameErr)),
		),
		gg.Div(gg.KV{"class": "space-y-1"},
			gg.Label(gg.KV{"for": "username", "class": "block text-sm font-medium text-fg-secondary"}, "Username"),
			gg.Input(gg.KV{"type": "text", "id": "username", "name": "username", "value": params.Username, "placeholder": "Choose a username", "required": true, "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			gg.If(params.UsernameErr != nil, gg.P(gg.KV{"class": "text-red-500 text-sm mt-1"}, params.UsernameErr)),
		),
		gg.Div(gg.KV{"class": "space-y-1"},
			gg.Label(gg.KV{"for": "email", "class": "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			gg.Input(gg.KV{"type": "email", "id": "email", "name": "email", "value": params.Email, "required": true, "placeholder": "you@example.com", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			gg.If(params.EmailErr != nil, gg.P(gg.KV{"class": "text-red-500 text-sm mt-1"}, params.EmailErr)),
		),
		gg.Div(gg.KV{"class": "space-y-1"},
			gg.Label(gg.KV{"for": "bio", "class": "block text-sm font-medium text-fg-secondary"}, "Bio"),
			gg.Textarea(gg.KV{"id": "bio", "name": "bio", "placeholder": "Tell us about yourself...", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors min-h-24 resize-none"},
				params.Bio,
			),
			gg.If(params.BioErr != nil, gg.P(gg.KV{"class": "text-red-500 text-sm mt-1"}, params.BioErr)),
		),
		// TODO: add disabled & spinner to other forms
		gg.Button(gg.KV{"class": "w-full rounded-lg bg-blue hover:bg-blue/80 disabled:hover:bg-blue disabled:opacity-50 disabled:cursor-not-allowed text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2 flex items-center justify-center gap-2"},
			"Save", spinner(),
		),
	)
}

func spinner() gg.Node {
	return gg.Span(gg.KV{"id": "spinner", "class": "htmx-indicator animate-spin hidden [&.htmx-request]:block"},
		gg.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-loader-icon lucide-loader"><path d="M12 2v4"/><path d="m16.2 7.8 2.9-2.9"/><path d="M18 12h4"/><path d="m16.2 16.2 2.9 2.9"/><path d="M12 18v4"/><path d="m4.9 19.1 2.9-2.9"/><path d="M2 12h4"/><path d="m4.9 4.9 2.9 2.9"/></svg>`),
	)
}

type SessionsTabParams struct {
}

func SessionsTab(parms SessionsTabParams) gg.Node {
	return gg.P("sessions tab")
}
