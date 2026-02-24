package components

import (
	"fmt"
	"slices"
	"strings"
	"time"

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
				h.Script(h.KV{"src": "/public/js/lib/htmx@2.0.8.js"}),
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

type RegisterFormParams struct {
	Name        string
	NameErr     any
	Username    string
	UsernameErr any
	Email       string
	EmailErr    any
	Bio         string
	BioErr      any
}

func RegisterForm(params ...RegisterFormParams) h.Node {
	var p RegisterFormParams
	if len(params) != 0 {
		p = params[0]
	}

	return h.Form(h.KV{"hx-post": "/register", "hx-swap": "outerHTML", "class": "space-y-5"},
		h.Div(h.KV{"class": "space-y-1"},
			h.Label(h.KV{"for": "name", "class": "block text-sm font-medium text-fg-secondary"}, "Full Name"),
			h.Input(h.KV{"type": "text", "id": "name", "name": "name", "required": true, "value": p.Name, "placeholder": "Enter your full name", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(p.NameErr != nil, h.P(h.KV{"class": "text-red-500 text-sm mt-1"}, p.NameErr)),
		),
		h.Div(h.KV{"class": "space-y-1"},
			h.Label(h.KV{"for": "username", "class": "block text-sm font-medium text-fg-secondary"}, "Username"),
			h.Input(h.KV{"type": "text", "id": "username", "name": "username", "value": p.Username, "placeholder": "Choose a username", "required": true, "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(p.UsernameErr != nil, h.P(h.KV{"class": "text-red-500 text-sm mt-1"}, p.UsernameErr)),
		),
		h.Div(h.KV{"class": "space-y-1"},
			h.Label(h.KV{"for": "email", "class": "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			h.Input(h.KV{"type": "email", "id": "email", "name": "email", "value": p.Email, "required": true, "placeholder": "you@example.com", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(p.EmailErr != nil, h.P(h.KV{"class": "text-red-500 text-sm mt-1"}, p.EmailErr)),
		),
		h.Div(h.KV{"class": "space-y-1"},
			h.Label(h.KV{"for": "bio", "class": "block text-sm font-medium text-fg-secondary"}, "Bio"),
			h.Textarea(h.KV{"id": "bio", "name": "bio", "placeholder": "Tell us about yourself...", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors min-h-24 resize-none"},
				p.Bio,
			),
			h.If(p.BioErr != nil, h.P(h.KV{"class": "text-red-500 text-sm mt-1"}, p.BioErr)),
		),
		h.Button(h.KV{"class": "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
			"Create Account",
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

type LoginFormParams struct {
	Email    string
	EmailErr any
}

func LoginForm(params ...LoginFormParams) h.Node {
	var p LoginFormParams
	if len(params) != 0 {
		p = params[0]
	}

	return h.Form(h.KV{"hx-post": "/login", "hx-swap": "outerHTML", "class": "space-y-5"},
		h.Div(h.KV{"class": "space-y-1"},
			h.Label(h.KV{"for": "email", "class": "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			h.Input(h.KV{"type": "email", "id": "email", "name": "email", "value": p.Email, "required": true, "placeholder": "you@example.com", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(p.EmailErr != nil, h.P(h.KV{"class": "text-red-500 text-sm mt-1"}, p.EmailErr)),
		),
		h.Button(h.KV{"class": "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
			"Sign In",
		),
	)
}

type OtpFormParams struct {
	OtpID  string
	Otp    string
	OtpErr any
}

func OtpForm(params OtpFormParams) h.Node {
	return h.Form(h.KV{"hx-post": "/verify_otp", "hx-swap": "outerHTML", "class": "space-y-5"},
		h.Input(h.KV{"type": "hidden", "name": "otpID", "value": params.OtpID}),
		h.Div(h.KV{"class": "space-y-1"},
			h.Label(h.KV{"for": "otp", "class": "block text-sm font-medium text-fg-secondary"}, "Verification Code"),
			h.P(h.KV{"class": "text-fg-secondary text-sm mb-2"}, "We've sent a 6-digit code to your email address. Please enter it below to verify your identity."),
			h.Input(h.KV{
				"type":         "text",
				"id":           "otp",
				"name":         "otp",
				"value":        params.Otp,
				"required":     true,
				"maxlength":    "6",
				"pattern":      "[0-9]{6}",
				"inputmode":    "numeric",
				"autocomplete": "one-time-code",
				"placeholder":  "000000",
				"class":        "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors text-center text-2xl tracking-widest",
			}),
			h.If(params.OtpErr != nil, h.P(h.KV{"class": "text-red-500 text-sm mt-1"}, params.OtpErr)),
		),
		h.Button(h.KV{"class": "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
			"Verify",
		),
	)
}

type ChatPageParams struct {
	User UserBlockParams
}

func ChatPage(params ChatPageParams) h.Node {
	return rootLayout(
		h.Div(h.KV{"id": "unread-manager-anchor"}),
		h.Script(h.RawText(`
				window.unreadManager = new class {
            constructor() {
							this.counts = {};
            }
            set(partnerID, delta) {
							this.counts[partnerID] = delta;
							setTimeout(() => {
								this.updateBadge(partnerID);
							}, 100);
            }
            add(partnerID, delta) {
							this.counts[partnerID] = (this.counts[partnerID] || 0) + delta;
							// this.updateBadge(partnerID);
            }
            sub(partnerID, delta) {
							this.counts[partnerID] = Math.max(0, (this.counts[partnerID] || 0) - delta);
							this.updateBadge(partnerID);
            }
						updateBadge(partnerID) {
							const badge = document.getElementById('unread-count-badge-' + partnerID);
							if (!badge) return;
							const count = this.counts[partnerID] || 0;
							if (count > 0) {
								badge.textContent = count > 99 ? '99+' : count;
								badge.classList.remove('hidden');
							} else {
								badge.textContent = '';
								badge.classList.add('hidden');
							}
						}
        };
		`)),
		h.Script(h.KV{"src": "/public/js/lib/htmx_ext_ws@2.0.4.js"}),
		h.Div(h.KV{
			"class":      "h-screen flex bg-bg-primary",
			"hx-ext":     "ws",
			"ws-connect": "/ws",
		},
			// Sidebar
			h.Div(h.KV{"class": "w-80 bg-bg-secondary border-r border-bg-tertiary flex flex-col"},
				// Sticky top row
				h.Div(h.KV{"class": "sticky top-0 z-10 h-16 bg-aqua/10 border-b border-bg-primary flex items-center"},
					UserBlock(params.User),
					h.Button(h.KV{
						"hx-get":     "/search_modal",
						"hx-trigger": "click",
						"hx-target":  "body",
						"hx-swap":    "beforeend",
						"class":      "flex items-center justify-center w-16 h-16 flex-shrink-0 hover:bg-aqua/30 transition-colors cursor-pointer",
					},
						h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>`),
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

func spinner(id string) h.Node {
	return h.RawText(fmt.Sprintf(`<svg
	id="%s"
	class="htmx-indicator animate-spin opacity-0 mx-auto hidden [&.htmx-request]:inline-block"
	xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-loader-icon lucide-loader"><path d="M12 2v4"/><path d="m16.2 7.8 2.9-2.9"/><path d="M18 12h4"/><path d="m16.2 16.2 2.9 2.9"/><path d="M12 18v4"/><path d="m4.9 19.1 2.9-2.9"/><path d="M2 12h4"/><path d="m4.9 4.9 2.9 2.9"/></svg>`,
		id,
	))
}

type UserBlockParams struct {
	Name     string
	Username string
}

func UserBlock(params UserBlockParams) h.Node {
	return h.Div(h.KV{
		"id":         "user-block",
		"hx-get":     "/profile_modal",
		"hx-trigger": "click",
		"hx-target":  "body",
		"hx-swap":    "beforeend",
		"class":      "w-64 h-16 flex-shrink-0 flex items-center gap-3 cursor-pointer hover:bg-aqua/20 transition-colors px-4",
	},
		h.Div(h.KV{"class": "w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold flex-shrink-0"},
			getInitials(params.Name),
		),
		h.Div(h.KV{"class": "flex-1 min-w-0"},
			h.P(h.KV{"class": "text-fg-primary font-medium truncate"}, params.Name),
			h.P(h.KV{"class": "text-fg-secondary text-sm truncate"}, "@"+params.Username),
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

func ProfileModal(params ProfileModalParams) h.Node {
	if params.ActiveTab == tabDefault {
		params.ActiveTab = TabProfile
	}

	profileIcon := h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>`)
	sessionsIcon := h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>`)

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
					h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`),
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

type ProfileTabParams struct {
	Name            string
	Username        string
	Email           string
	EmailIsVerified bool
	Bio             string
	JoinedAt        time.Time
}

func ProfileTab(params ProfileTabParams) h.Node {
	return h.Div(
		ProfileForm(ProfileFormParams{
			Name:     params.Name,
			Username: params.Username,
			Email:    params.Email,
			Bio:      params.Bio,
		}),
		h.P(h.KV{"class": "text-fg-secondary text-sm text-center my-4"},
			"Joined ", params.JoinedAt.Format("January 2, 2006"),
		),
		h.Div(h.KV{"class": "flex items-center justify-center gap-2 text-sm"},
			h.IfElse(params.EmailIsVerified,
				h.Empty(
					h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-green-500"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>`),
					h.P(h.KV{"class": "text-green-600"}, "Email verified"),
				),
				h.Empty(
					h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-yellow-500"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>`),
					h.P(h.KV{"class": "text-yellow-600"}, "Email not verified"),
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

func ProfileForm(params ProfileFormParams) h.Node {
	return h.Form(h.KV{
		"hx-put":          "/profile",
		"hx-swap":         "outerHTML",
		"hx-disabled-elt": "find button",
		"hx-indicator":    "#spinner",
		"class":           "space-y-5",
	},
		h.Div(h.KV{"class": "space-y-1"},
			h.Label(h.KV{"for": "name", "class": "block text-sm font-medium text-fg-secondary"}, "Full Name"),
			h.Input(h.KV{"type": "text", "id": "name", "name": "name", "required": true, "value": params.Name, "placeholder": "Enter your full name", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(params.NameErr != nil, h.P(h.KV{"class": "text-red-500 text-sm mt-1"}, params.NameErr)),
		),
		h.Div(h.KV{"class": "space-y-1"},
			h.Label(h.KV{"for": "username", "class": "block text-sm font-medium text-fg-secondary"}, "Username"),
			h.Input(h.KV{"type": "text", "id": "username", "name": "username", "value": params.Username, "placeholder": "Choose a username", "required": true, "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(params.UsernameErr != nil, h.P(h.KV{"class": "text-red-500 text-sm mt-1"}, params.UsernameErr)),
		),
		h.Div(h.KV{"class": "space-y-1"},
			h.Label(h.KV{"for": "email", "class": "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			h.Input(h.KV{"type": "email", "id": "email", "name": "email", "value": params.Email, "required": true, "placeholder": "you@example.com", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(params.EmailErr != nil, h.P(h.KV{"class": "text-red-500 text-sm mt-1"}, params.EmailErr)),
		),
		h.Div(h.KV{"class": "space-y-1"},
			h.Label(h.KV{"for": "bio", "class": "block text-sm font-medium text-fg-secondary"}, "Bio"),
			h.Textarea(h.KV{"id": "bio", "name": "bio", "placeholder": "Tell us about yourself...", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors min-h-24 resize-none"},
				params.Bio,
			),
			h.If(params.BioErr != nil, h.P(h.KV{"class": "text-red-500 text-sm mt-1"}, params.BioErr)),
		),
		// TODO: Add disabled & spinner to other forms
		h.Button(h.KV{"class": "w-full rounded-lg bg-blue hover:bg-blue/80 disabled:hover:bg-blue disabled:opacity-50 disabled:cursor-not-allowed text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2 flex items-center justify-center gap-2"},
			"Save", spinner("spinner"),
		),
	)
}

type SessionsTabParams struct {
	Sessions         []Session
	CurrentSessionID string
}

type Session struct {
	ID       string
	Platform string
	Os       string
}

func SessionsTab(params SessionsTabParams) h.Node {
	// Put the current session first
	if index := slices.IndexFunc(params.Sessions, func(s Session) bool {
		return s.ID == params.CurrentSessionID
	}); index != -1 {
		params.Sessions[0], params.Sessions[index] = params.Sessions[index], params.Sessions[0]
	}

	return h.Div(h.KV{"id": "sessions-tab", "class": "space-y-3"},
		h.MapSlice(params.Sessions, func(session Session) h.Node {
			elementID := fmt.Sprintf("session-%s", session.ID)
			isCurrent := params.CurrentSessionID == session.ID
			spinnerID := fmt.Sprintf("spinner-%s", session.ID)

			return h.Div(h.KV{
				"id":    elementID,
				"class": "flex items-center justify-between p-4 rounded-xl border-2 " + h.IfElse(isCurrent, "border-blue bg-blue/5", "border-bg-tertiary bg-bg-tertiary/30"),
			},
				// Left side: Session info
				h.Div(h.KV{"class": "flex flex-col"},
					h.Div(h.KV{"class": "flex items-center gap-2"},
						h.P(h.KV{"class": "font-semibold text-fg-primary"}, session.Platform),
						h.If(isCurrent,
							h.Span(h.KV{"class": "px-2 py-0.5 text-xs font-medium bg-blue text-bg-primary rounded-full"}, "Current"),
						),
					),
					h.P(h.KV{"class": "text-sm text-fg-secondary"}, session.Os),
				),
				// Right side: Action button
				h.IfElse(isCurrent,
					h.Button(h.KV{
						"hx-post":         "/logout",
						"hx-disabled-elt": "this",
						"hx-indicator":    "#" + spinnerID,
						"class":           "px-4 py-2 rounded-lg bg-red-500/10 hover:bg-red-500/20 text-red-500 font-medium transition-colors flex items-center gap-2 cursor-pointer",
					},
						h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>`),
						"Logout", spinner(spinnerID),
					),
					h.Button(h.KV{
						"hx-delete":       fmt.Sprintf("/sessions/%s", session.ID),
						"hx-target":       "#" + elementID,
						"hx-swap":         "delete",
						"hx-disabled-elt": "this",
						"hx-indicator":    "#" + spinnerID,
						"class":           "px-4 py-2 rounded-lg hover:bg-red-500/10 text-fg-secondary hover:text-red-500 font-medium transition-colors flex items-center gap-2 cursor-pointer",
					},
						h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>`),
						"Remove", spinner(spinnerID),
					),
				),
			)
		}),
	)
}

func SearchModal() h.Node {
	return h.Div(h.KV{
		"id":          "search-modal",
		"class":       "fixed inset-0 z-50 flex items-start justify-center bg-black/50 backdrop-blur-sm p-4 pt-20 outline-none",
		"hx-on:click": "if (event.target === this) this.remove()",
	},
		h.Div(h.KV{"class": "bg-bg-secondary rounded-2xl shadow-2xl w-full max-w-xl flex flex-col overflow-hidden"},
			// Header with search input
			h.Div(h.KV{"class": "flex items-center gap-3 p-4 border-b border-bg-tertiary"},
				h.Div(h.KV{"class": "flex-1 relative"},
					h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="absolute left-3 top-1/2 -translate-y-1/2 text-fg-secondary"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>`),
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
					h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`),
				),
			),
			// Content area with results
			h.Div(h.KV{"id": "search-results", "class": "flex-1 p-4 overflow-y-auto max-h-96"},
				h.P(h.KV{"class": "text-fg-secondary text-center py-3"}, "Search for users by name or username"),
			),
		),
	)
}

type SearchResultParams struct {
	Query   string
	HasMore bool
	Items   []SearchResultItemParams
}

func SearchResult(params SearchResultParams) h.Node {
	if params.Query == "" {
		return h.P(h.KV{"class": "text-fg-secondary text-center py-3"}, "Search for users by name or username")
	}

	if len(params.Items) == 0 {
		return h.Empty()
	}

	lastID := params.Items[len(params.Items)-1].ID

	return h.Div(h.KV{"class": "space-y-1"},
		h.MapSlice(params.Items, func(profile SearchResultItemParams) h.Node {
			return h.Div(h.KV{
				"hx-get":     "/search/users/select/" + profile.ID,
				"hx-trigger": "click",
				"hx-target":  "#chat-container",
				"hx-swap":    "innerHTML",
			},
				h.IfElse(profile.ID == lastID,
					h.Div(h.KV{
						"hx-get":     "/search/users?query=" + params.Query + "&cursor=" + lastID,
						"hx-trigger": "intersect once",
						"hx-swap":    "afterend",
					},
						searchResultItem(profile),
					),
					searchResultItem(profile),
				),
			)
		}),
	)
}

type SearchResultItemParams struct {
	ID       string
	Name     string
	Username string
}

func searchResultItem(params SearchResultItemParams) h.Node {
	return h.Div(h.KV{"class": "flex items-center gap-3 p-3 hover:bg-bg-tertiary/50 hover:rounded-lg cursor-pointer transition-colors"},
		h.Div(h.KV{"class": "relative w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(params.Name),
		),
		h.Div(h.KV{"class": "flex-1 min-w-0"},
			h.P(h.KV{"class": "text-fg-primary font-medium truncate"}, params.Name),
			h.P(h.KV{"class": "text-fg-secondary text-sm truncate"}, "@"+params.Username),
		),
	)
}

type PartnersListParams struct {
	Partners []PartnerBlockParams
	// This is the cursor. Empty when no more partners
	LastMessageWithLastPartnerID string
}

func PartnersList(params PartnersListParams) h.Node {
	if len(params.Partners) == 0 {
		return h.Div(h.KV{
			"id":    "partners-list",
			"class": "space-y-1",
		})
	}

	lastID := params.Partners[len(params.Partners)-1].ID

	return h.Div(h.KV{
		"id":    "partners-list",
		"class": "space-y-1",
	},
		h.MapSlice(params.Partners, func(partner PartnerBlockParams) h.Node {
			attrs := h.IfElse(partner.ID == lastID && params.LastMessageWithLastPartnerID != "", h.KV{
				"hx-get":       "/partners?cursor=" + params.LastMessageWithLastPartnerID,
				"hx-trigger":   "intersect once",
				"hx-swap":      "afterend",
				"hx-indicator": "#partners-indicator",
				// Disable the sidebar indicator for requests in blocks
				"hx-disinherit": "hx-indicator",
			}, nil)

			return h.Div(attrs, PartnersListItem(partner))
		}),
	)
}

func PartnersListItem(params PartnerBlockParams) h.Node {
	return h.Div(h.KV{
		"id":         "partner-" + params.ID,
		"class":      "cursor-pointer transition-colors",
		"hx-get":     "/chat/" + params.ID,
		"hx-trigger": "click",
		"hx-target":  "#chat-container",
		"hx-swap":    "innerHTML",
		// Cancel the request if clicking the active partner
		"hx-on::config-request": `
			if (window.currentActivePartnerId === "` + params.ID + `") {
				event.preventDefault();
				return;
			}
		`,
		"hx-on::load": `
			// Re-applies data-active attribute.
			// This prevents losing the active styles for newly instered element if it was the active
			// as the old one which has the attribute is deleted by the oob-swap response.
			document.getElementById("partner-"+window.currentActivePartnerId)?.setAttribute("data-active", "");

			window.unreadManager.updateBadge("` + params.ID + `");
		`,
	},
		partnerBlock(params),
	)
}

type PartnerBlockParams struct {
	ID       string
	Name     string
	Username string
	IsOnline bool
}

func partnerBlock(params PartnerBlockParams) h.Node {
	return h.Div(h.KV{
		"class": "flex items-center gap-3 p-3 hover:bg-bg-tertiary/50 hover:rounded-lg cursor-pointer transition-colors",
	},
		h.Div(h.KV{"class": "relative w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(params.Name),
			PartnerBlockPresenceIndicator(params.ID, params.IsOnline),
		),
		h.Div(h.KV{"class": "flex-1 min-w-0"},
			h.P(h.KV{"class": "text-fg-primary font-medium truncate"}, params.Name),
			h.P(h.KV{"class": "text-fg-secondary text-sm truncate"}, "@"+params.Username),
		),
		h.Span(h.KV{
			"id":    "unread-count-badge-" + params.ID,
			"class": "flex-shrink-0 min-w-[20px] h-5 px-1.5 bg-blue text-bg-primary text-xs font-bold rounded-full flex items-center justify-center hidden",
		}),
	)
}

func PartnerBlockPresenceIndicator(partnerID string, isOnline bool) h.Node {
	return h.Div(h.KV{
		"id": "profile-block-presence-indicator-" + partnerID,
		"class": "absolute bottom-0 right-0 w-3.5 h-3.5 bg-green rounded-full border-2 border-bg-primary " +
			h.IfElse(!isOnline, "hidden", ""),
	})
}

type ChatContainerParams struct {
	Partner PartnerBlockParams
}

func ChatContainer(params ChatContainerParams) h.Node {
	return h.Div(h.KV{
		"class": "flex-1 flex flex-col bg-bg-primary h-full",
		"hx-on::load": `
			window.currentActivePartnerId = "` + params.Partner.ID + `";
		  document.querySelector("#partners-list [data-active]")?.removeAttribute("data-active");
		  document.getElementById("partner-" + "` + params.Partner.ID + `")?.setAttribute("data-active", "");
		`,
	},
		ChatContainerHeader(params.Partner),
		h.Div(h.KV{
			"id":    "messages-container",
			"class": "flex-1 overflow-y-auto px-6 sm:px-10 py-4 flex flex-col-reverse gap-3",
		},
			h.Div(h.KV{
				"id": "new-message-inserter",
				"hx-on::oob-before-swap": `
					// don't insert the new message if it doesn't come from the active partner
					if (window.currentActivePartnerId !== "` + params.Partner.ID + `") {
						event.preventDefault();
						return;
					}
				`,
			}),
			h.Div(h.KV{
				"hx-get":              fmt.Sprintf("/chat/%s/messages", params.Partner.ID),
				"hx-trigger":          "load",
				"hx-swap":             "afterend",
				"hx-indicator":        "#messages-indicator",
				"hx-on::after-settle": "this.remove()",
			}),
			h.Div(h.KV{"class": "flex justify-center"},
				spinner("messages-indicator"),
			),
		),
		ChatInputForm(ChatInputFormParams{PartnerID: params.Partner.ID}),
	)
}

func ChatContainerHeader(partner PartnerBlockParams) h.Node {
	return h.Div(h.KV{
		"id":    "chat-container-header-" + partner.ID,
		"class": "h-16 px-4 bg-bg-secondary border-b border-bg-tertiary flex items-center gap-3",
	},
		h.Div(h.KV{"class": "relative w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(partner.Name),
			ChatContainerPresenceIndicator(partner.ID, partner.IsOnline),
		),
		h.Div(h.KV{"class": "flex-1 min-w-0"},
			h.P(h.KV{"class": "text-fg-primary font-medium truncate"}, partner.Name),
			h.P(h.KV{"class": "text-fg-secondary text-sm truncate"}, "@"+partner.Username),
		),
	)
}

func ChatContainerPresenceIndicator(partnerID string, isOnline bool) h.Node {
	return h.Div(h.KV{
		"id": "chat-container-presence-indicator-" + partnerID,
		"class": "absolute bottom-0 right-0 w-3.5 h-3.5 bg-green rounded-full border-2 border-bg-primary " +
			h.IfElse(!isOnline, "hidden", ""),
	})
}

type ChatMessagesListParams struct {
	PartnerID string
	Messages  []ChatMessageParams
	HasMore   bool
}

func ChatMessagesList(params ChatMessagesListParams) h.Node {
	if len(params.Messages) == 0 {
		return h.Empty()
	}

	cursorMessageID := params.Messages[len(params.Messages)-1].ID

	return h.MapSlice(params.Messages, func(msg ChatMessageParams) h.Node {
		msg.PartnerID = params.PartnerID
		return h.Div(
			h.IfElse(params.HasMore && msg.ID == cursorMessageID,
				h.KV{
					"hx-get":       fmt.Sprintf("/chat/%s/messages?cursor=%s", params.PartnerID, cursorMessageID),
					"hx-trigger":   "intersect once",
					"hx-swap":      "afterend",
					"hx-indicator": "#messages-indicator",
					// For internal requests not to use this indicator
					"hx-disinherit": "hx-indicator",
				},
				nil,
			),
			ChatMessage(msg),
		)
	})
}

type ChatMessageParams struct {
	ID        string
	PartnerID string
	Content   string
	SentAt    time.Time
	IsRead    bool
	FromMe    bool
}

func ChatMessage(params ChatMessageParams) h.Node {
	return h.Div(
		h.IfElse(!params.FromMe && !params.IsRead,
			h.KV{
				"hx-post":    "/api/v1/chats/" + params.PartnerID + "/mark_as_read?upto_message_id=" + params.ID,
				"hx-trigger": "intersect once",
				"hx-swap":    "none",
			},
			nil,
		),
		h.Div(h.KV{"class": "flex w-full " + h.IfElse(params.FromMe, "justify-end", "justify-start")},
			h.Div(h.KV{"class": "flex flex-col " + h.IfElse(params.FromMe, "items-end", "items-start") + " max-w-[70%]"},
				h.Div(h.KV{"class": "px-4 py-2 " + h.IfElse(params.FromMe, "bg-blue text-bg-primary rounded-l-2xl rounded-tr-2xl", "bg-bg-tertiary text-fg-primary rounded-r-2xl rounded-tl-2xl")},
					h.P(h.KV{"class": "whitespace-pre-wrap"}, params.Content),
					h.Div(h.KV{"class": "flex items-center gap-1 mt-1 justify-end"},
						h.P(h.KV{"class": "text-xs " + h.IfElse(params.FromMe, "text-bg-primary/70", "text-fg-secondary")}, params.SentAt.Format("Jan 2, 3:04 PM")),
						h.If(params.FromMe,
							h.IfElse(params.IsRead, ReadMessageIndicator(), unreadMessageIndicator(params.ID)),
						),
					),
				),
			),
		),
	)
}

func ReadMessageIndicator() h.Node {
	return h.Span(h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-bg-primary"><polyline points="16 6 8 14 4 10"></polyline><polyline points="22 6 14 14 10 10"></polyline></svg>`))
}

func unreadMessageIndicator(messageID string) h.Node {
	return h.Span(h.RawText(fmt.Sprintf(`
		<svg id="unread-message-indicator-%s" xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-bg-primary"><polyline points="20 6 9 17 4 12"></polyline></svg>`,
		messageID,
	)))
}

type ChatInputFormParams struct {
	PartnerID string
}

// TODO: Handle message retry when failing,
// and add an indicator when sending message until
// a confirmation message is recived `MessageWasSent`.
func ChatInputForm(params ChatInputFormParams) h.Node {
	return h.Form(h.KV{
		"ws-send": true,
		"hx-on::ws-config-send": `
			const content = event.detail.parameters.content.trim();
			if (content == "") {
				event.preventDefault();
				return;
			}
			event.detail.parameters.content = content;
		`,
		"hx-on::ws-after-send": `
			const textarea = event.detail.elt.querySelector('textarea[name="content"]');
			textarea.value = '';
			textarea.style.height = 'auto';
		`,
		"class": "flex items-center gap-2 px-3 py-2",
	},
		h.Input(h.KV{"type": "hidden", "name": "kind", "value": "SendMessage"}),
		h.Input(h.KV{"type": "hidden", "name": "partnerID", "value": params.PartnerID}),
		h.Textarea(h.KV{
			"class":       "flex-1 bg-bg-tertiary text-fg-primary resize-none outline-none max-h-[120px] min-h-[44px] py-2.5 px-4 rounded-2xl leading-6 self-center overflow-y-auto",
			"name":        "content",
			"placeholder": "Write a message...",
			"rows":        "1",
			"autofocus":   true,
			"oninput": `
				this.style.height = "auto";
				this.style.height = Math.min(this.scrollHeight, 120)+"px";
			`,
			"onkeydown": `
				if (event.key === "Enter" && !event.shiftKey) {
					event.preventDefault();
					document.getElementById("send-btn").click();
				}
			`,
		}),
		h.Button(h.KV{
			"id":    "send-btn",
			"type":  "submit",
			"class": "w-12 h-12 rounded-full bg-bg-primary hover:bg-bg-tertiary flex items-center justify-center flex-shrink-0 cursor-pointer transition-colors",
		},
			h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-send-horizontal-icon lucide-send-horizontal text-blue"><path d="M3.714 3.048a.498.498 0 0 0-.683.627l2.843 7.627a2 2 0 0 1 0 1.396l-2.842 7.627a.498.498 0 0 0 .682.627l18-8.5a.5.5 0 0 0 0-.904z"/><path d="M6 12h16"/></svg>`),
		),
	)
}
