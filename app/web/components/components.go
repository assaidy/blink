package components

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/assaidy/hyper"
	"github.com/assaidy/hyper/htmx"
	"github.com/assaidy/hyper/htmx/extensions/ws"
)

func rootLayout(children ...any) h.Node {
	return h.Empty(
		h.DoctypeHtml(),
		h.Html(
			h.Head(
				h.Title("blink"),
				h.Meta(h.KV{h.AttrCharset: "UTF-8"}),
				h.Meta(h.KV{h.AttrName: "viewport", h.AttrContent: "width=device-width, initial-scale=1.0"}),
				h.Link(h.KV{h.AttrRel: "stylesheet", h.AttrHref: "/public/css/style.css"}),
				h.Script(h.KV{h.AttrSrc: "/public/js/lib/htmx@2.0.8.js"}),
				h.Script(h.RawText(`
					const $cookie = (name) => {
						return document.cookie
							.split('; ')
							.find(row => row.startsWith(name+"="))
							?.split('=')[1]
							?.trim() || '';
					}
				`)),
			),
			h.Body(h.KV{
				hx.AttrHxOn(hx.EventConfigRequest): `
					event.detail.headers['X-CSRF-Token'] = $cookie("csrf_token");
				`,
				h.AttrClass: "bg-bg-primary text-fg-primary",
			},
				h.Div(children...),
			),
		),
	)
}

func RegisterPage() h.Node {
	return rootLayout(
		h.Div(h.KV{h.AttrClass: "min-h-screen flex justify-center items-center bg-bg-primary sm:px-6 lg:px-8"},
			h.Div(h.KV{h.AttrClass: "w-full min-h-screen sm:min-h-0 sm:max-w-md sm:p-8 bg-bg-secondary sm:rounded-lg sm:shadow-lg flex flex-col justify-center"},
				h.Div(h.KV{h.AttrClass: "p-6 sm:p-0"},
					h.H2(h.KV{h.AttrClass: "text-fg-primary text-2xl font-bold text-center mb-8"}, "Create Account"),
					RegisterForm(),
					h.P(h.KV{h.AttrClass: "text-center text-fg-secondary text-sm mt-6"},
						"Already have an account? ", h.A(h.KV{h.AttrHref: "/login", h.AttrClass: "text-blue hover:underline font-medium"}, "Sign in"),
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

	return h.Form(h.KV{hx.AttrHxPost: "/register", hx.AttrHxSwap: hx.SwapOuterHtml, h.AttrClass: "space-y-5"},
		h.Div(h.KV{h.AttrClass: "space-y-1"},
			h.Label(h.KV{"for": "name", h.AttrClass: "block text-sm font-medium text-fg-secondary"}, "Full Name"),
			h.Input(h.KV{h.AttrType: "text", h.AttrId: "name", h.AttrName: "name", h.AttrRequired: true, h.AttrValue: p.Name, h.AttrPlaceholder: "Enter your full name", h.AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(p.NameErr != nil, h.P(h.KV{h.AttrClass: "text-red-500 text-sm mt-1"}, p.NameErr)),
		),
		h.Div(h.KV{h.AttrClass: "space-y-1"},
			h.Label(h.KV{"for": "username", h.AttrClass: "block text-sm font-medium text-fg-secondary"}, "Username"),
			h.Input(h.KV{h.AttrType: "text", h.AttrId: "username", h.AttrName: "username", h.AttrValue: p.Username, h.AttrPlaceholder: "Choose a username", h.AttrRequired: true, h.AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(p.UsernameErr != nil, h.P(h.KV{h.AttrClass: "text-red-500 text-sm mt-1"}, p.UsernameErr)),
		),
		h.Div(h.KV{h.AttrClass: "space-y-1"},
			h.Label(h.KV{"for": "email", h.AttrClass: "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			h.Input(h.KV{h.AttrType: "email", h.AttrId: "email", h.AttrName: "email", h.AttrValue: p.Email, h.AttrRequired: true, h.AttrPlaceholder: "you@example.com", h.AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(p.EmailErr != nil, h.P(h.KV{h.AttrClass: "text-red-500 text-sm mt-1"}, p.EmailErr)),
		),
		h.Div(h.KV{h.AttrClass: "space-y-1"},
			h.Label(h.KV{"for": "bio", h.AttrClass: "block text-sm font-medium text-fg-secondary"}, "Bio"),
			h.Textarea(h.KV{h.AttrId: "bio", h.AttrName: "bio", h.AttrPlaceholder: "Tell us about yourself...", h.AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors min-h-24 resize-none"},
				p.Bio,
			),
			h.If(p.BioErr != nil, h.P(h.KV{h.AttrClass: "text-red-500 text-sm mt-1"}, p.BioErr)),
		),
		h.Button(h.KV{h.AttrClass: "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
			"Create Account",
		),
	)
}

func LoginPage() h.Node {
	return rootLayout(
		h.Div(h.KV{h.AttrClass: "min-h-screen flex justify-center items-center bg-bg-primary sm:px-6 lg:px-8"},
			h.Div(h.KV{h.AttrClass: "w-full min-h-screen sm:min-h-0 sm:max-w-md sm:p-8 bg-bg-secondary sm:rounded-lg sm:shadow-lg flex flex-col justify-center"},
				h.Div(h.KV{h.AttrClass: "p-6 sm:p-0"},
					h.H2(h.KV{h.AttrClass: "text-fg-primary text-2xl font-bold text-center mb-8"}, "Sign In"),
					LoginForm(),
					h.P(h.KV{h.AttrClass: "text-center text-fg-secondary text-sm mt-6"},
						"Don't have an account? ", h.A(h.KV{h.AttrHref: "/register", h.AttrClass: "text-blue hover:underline font-medium"}, "Create one"),
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

	return h.Form(h.KV{hx.AttrHxPost: "/login", hx.AttrHxSwap: hx.SwapOuterHtml, h.AttrClass: "space-y-5"},
		h.Div(h.KV{h.AttrClass: "space-y-1"},
			h.Label(h.KV{"for": "email", h.AttrClass: "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			h.Input(h.KV{h.AttrType: "email", h.AttrId: "email", h.AttrName: "email", h.AttrValue: p.Email, h.AttrRequired: true, h.AttrPlaceholder: "you@example.com", h.AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(p.EmailErr != nil, h.P(h.KV{h.AttrClass: "text-red-500 text-sm mt-1"}, p.EmailErr)),
		),
		h.Button(h.KV{h.AttrClass: "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
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
	return h.Form(h.KV{hx.AttrHxPost: "/verify_otp", hx.AttrHxSwap: hx.SwapOuterHtml, h.AttrClass: "space-y-5"},
		h.Input(h.KV{h.AttrType: h.TypeHidden, h.AttrName: "otpID", h.AttrValue: params.OtpID}),
		h.Div(h.KV{h.AttrClass: "space-y-1"},
			h.Label(h.KV{"for": "otp", h.AttrClass: "block text-sm font-medium text-fg-secondary"}, "Verification Code"),
			h.P(h.KV{h.AttrClass: "text-fg-secondary text-sm mb-2"}, "We've sent a 6-digit code to your email address. Please enter it below to verify your identity."),
			h.Input(h.KV{
				h.AttrType:         "text",
				h.AttrId:           "otp",
				h.AttrName:         "otp",
				h.AttrValue:        params.Otp,
				h.AttrRequired:     true,
				h.AttrMaxLength:    "6",
				h.AttrPattern:      "[0-9]{6}",
				h.AttrInputMode:    "numeric",
				h.AttrAutocomplete: "one-time-code",
				h.AttrPlaceholder:  "000000",
				h.AttrClass:        "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors text-center text-2xl tracking-widest",
			}),
			h.If(params.OtpErr != nil, h.P(h.KV{h.AttrClass: "text-red-500 text-sm mt-1"}, params.OtpErr)),
		),
		h.Button(h.KV{h.AttrClass: "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
			"Verify",
		),
	)
}

type ChatPageParams struct {
	User UserBlockParams
}

func ChatPage(params ChatPageParams) h.Node {
	return rootLayout(
		h.Script(h.KV{h.AttrSrc: "/public/js/lib/htmx_ext_ws@2.0.4.js"}),
		h.Script(h.RawText(`
			htmx.createWebSocket = (url) => {
				const csrfToken = $cookie("csrf_token");
				const fullUrl = csrfToken ? url+"?csrf_token="+csrfToken : url;
				return new WebSocket(fullUrl);
			};
		`)),
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

			window.messageIDGenerator = new class {
				constructor() {
					this.id = 1;
				}
				GetID() {
					return this.id++;
				}
			};
		`)),
		h.Div(h.KV{h.AttrId: "unread-manager-anchor"}),
		// Actual page content
		h.Div(h.KV{
			h.AttrClass:        "h-screen flex bg-bg-primary",
			hx.AttrHxExt:         "ws",
			hxws.AttrWsConnect: "/ws",
		},
			// Sidebar
			h.Div(h.KV{h.AttrClass: "w-80 bg-bg-secondary border-r border-bg-tertiary flex flex-col"},
				// Sticky top row
				h.Div(h.KV{h.AttrClass: "sticky top-0 z-10 h-16 bg-aqua/10 border-b border-bg-primary flex items-center"},
					UserBlock(params.User),
					h.Button(h.KV{
						hx.AttrHxGet:     "/search_modal",
						hx.AttrHxTrigger: "click",
						hx.AttrHxTarget:  "body",
						hx.AttrHxSwap:    hx.SwapBeforeEnd,
						h.AttrClass:    "flex items-center justify-center w-16 h-16 flex-shrink-0 hover:bg-aqua/30 transition-colors cursor-pointer",
					},
						h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>`),
					),
				),
				// Partners list container
				h.Div(h.KV{
					h.AttrId:    "partners-container",
					h.AttrClass: "flex-1 overflow-y-auto px-2 py-2",
				},
					h.Div(h.KV{
						hx.AttrHxGet:                     "/partners",
						hx.AttrHxTrigger:                 "load",
						hx.AttrHxSwap:                    hx.SwapAfterEnd,
						hx.AttrHxIndicator:               "#partners-indicator",
						hx.AttrHxOn(hx.EventAfterSettle): "document.getElementById('partners-container').scrollTop = 0; this.remove()",
					}),
					h.Div(h.KV{h.AttrClass: "flex justify-center"},
						spinner("partners-indicator"),
					),
				),
			),
			h.Div(h.KV{
				h.AttrId:    "chat-container",
				h.AttrClass: "flex-1 flex flex-col bg-bg-primary",
			},
				ChatContainerPlaceholder(),
			),
		),
	)
}

func ChatContainerPlaceholder() h.Node {
	return h.Div(h.KV{h.AttrClass: "flex-1 flex items-center justify-center"},
		h.P(h.KV{h.AttrClass: "text-fg-secondary text-lg"}, "Select a chat to start messaging"),
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
		h.AttrId:       "user-block",
		hx.AttrHxGet:     "/profile_modal",
		hx.AttrHxTrigger: "click",
		hx.AttrHxTarget:  "body",
		hx.AttrHxSwap:    hx.SwapBeforeEnd,
		h.AttrClass:    "w-64 h-16 flex-shrink-0 flex items-center gap-3 cursor-pointer hover:bg-aqua/20 transition-colors px-4",
	},
		h.Div(h.KV{h.AttrClass: "w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold flex-shrink-0"},
			getInitials(params.Name),
		),
		h.Div(h.KV{h.AttrClass: "flex-1 min-w-0"},
			h.P(h.KV{h.AttrClass: "text-fg-primary font-medium truncate"}, params.Name),
			h.P(h.KV{h.AttrClass: "text-fg-secondary text-sm truncate"}, "@"+params.Username),
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
		h.AttrId:                "profile-modal",
		h.AttrClass:             "fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 outline-none",
		hx.AttrHxOn(h.EventClick): "if (event.target === this) this.remove()",
	},
		h.Div(h.KV{h.AttrClass: "bg-bg-secondary rounded-2xl shadow-2xl max-w-3xl w-full flex overflow-hidden", h.AttrStyle: "height: 80vh; max-height: 700px;"},
			// Left sidebar with tabs
			h.Div(h.KV{h.AttrClass: "w-56 bg-bg-tertiary flex flex-col p-2 flex-shrink-0"},
				// Close button at top
				h.Button(h.KV{
					h.AttrClass:             "self-start p-2 hover:bg-bg-secondary rounded-lg transition-colors cursor-pointer mb-6",
					hx.AttrHxOn(h.EventClick): "this.closest('#profile-modal').remove()",
				},
					h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`),
				),
				// Tabs
				h.Div(h.KV{h.AttrClass: "flex flex-col gap-1"},
					h.Button(h.KV{
						hx.AttrHxGet:     "/profile_modal?tab=profile",
						hx.AttrHxTarget:  "#profile-modal",
						hx.AttrHxSwap:    hx.SwapOuterHtml,
						hx.AttrHxTrigger: h.IfElse(params.ActiveTab == TabProfile, hx.SwapNone, "click"),
						h.AttrClass:    "flex items-center gap-3 px-2 py-2.5 rounded-lg text-left font-medium text-fg-primary " + h.IfElse(params.ActiveTab == TabProfile, "bg-bg-secondary", "hover:bg-bg-secondary/50"),
					},
						profileIcon, "Profile",
					),
					h.Button(h.KV{
						hx.AttrHxGet:     "/profile_modal?tab=sessions",
						hx.AttrHxTarget:  "#profile-modal",
						hx.AttrHxSwap:    hx.SwapOuterHtml,
						hx.AttrHxTrigger: h.IfElse(params.ActiveTab == TabSessions, hx.SwapNone, "click"),
						h.AttrClass:    "flex items-center gap-3 px-2 py-2.5 rounded-lg text-left font-medium text-fg-primary " + h.IfElse(params.ActiveTab == TabSessions, "bg-bg-secondary", "hover:bg-bg-secondary/50"),
					},
						sessionsIcon, "Sessions",
					),
				),
			),
			// Right content area
			h.Div(h.KV{h.AttrClass: "flex-1 flex flex-col min-w-0"},
				// Header
				h.Div(h.KV{h.AttrClass: "px-8 py-6 border-b border-bg-tertiary"},
					h.H2(h.KV{h.AttrClass: "text-xl font-semibold text-fg-primary"},
						h.IfElse(params.ActiveTab == TabProfile, "Profile", "Sessions"),
					),
				),
				// Content
				h.Div(h.KV{h.AttrId: "tab-content", h.AttrClass: "flex-1 p-8 overflow-y-auto"},
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
		h.P(h.KV{h.AttrClass: "text-fg-secondary text-sm text-center my-4"},
			"Joined ", params.JoinedAt.Format("January 2, 2006"),
		),
		h.Div(h.KV{h.AttrClass: "flex items-center justify-center gap-2 text-sm"},
			h.IfElse(params.EmailIsVerified,
				h.Empty(
					h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-green-500"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>`),
					h.P(h.KV{h.AttrClass: "text-green-600"}, "Email verified"),
				),
				h.Empty(
					h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-yellow-500"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>`),
					h.P(h.KV{h.AttrClass: "text-yellow-600"}, "Email not verified"),
				),
			),
		),
		h.Div(h.KV{h.AttrClass: "border-t border-bg-tertiary my-6"}),
		h.Button(h.KV{
			hx.AttrHxDelete:  "/profile",
			hx.AttrHxSwap:    hx.SwapNone,
			hx.AttrHxConfirm: "Are you sure you wish to delete your account? This action cannot be undone.",
			h.AttrClass:    "w-full px-4 py-2 rounded-lg !bg-red-600 hover:!bg-red-700 text-white font-medium transition-colors flex items-center justify-center gap-2 cursor-pointer",
		},
			h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>`),
			"Delete Account",
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
		hx.AttrHxPut:         "/profile",
		hx.AttrHxSwap:        hx.SwapOuterHtml,
		hx.AttrHxDisabledElt: "find button",
		hx.AttrHxIndicator:   "#spinner",
		h.AttrClass:        "space-y-5",
	},
		h.Div(h.KV{h.AttrClass: "space-y-1"},
			h.Label(h.KV{"for": "name", h.AttrClass: "block text-sm font-medium text-fg-secondary"}, "Full Name"),
			h.Input(h.KV{h.AttrType: "text", h.AttrId: "name", h.AttrName: "name", h.AttrRequired: true, h.AttrValue: params.Name, h.AttrPlaceholder: "Enter your full name", h.AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(params.NameErr != nil, h.P(h.KV{h.AttrClass: "text-red-500 text-sm mt-1"}, params.NameErr)),
		),
		h.Div(h.KV{h.AttrClass: "space-y-1"},
			h.Label(h.KV{"for": "username", h.AttrClass: "block text-sm font-medium text-fg-secondary"}, "Username"),
			h.Input(h.KV{h.AttrType: "text", h.AttrId: "username", h.AttrName: "username", h.AttrValue: params.Username, h.AttrPlaceholder: "Choose a username", h.AttrRequired: true, h.AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(params.UsernameErr != nil, h.P(h.KV{h.AttrClass: "text-red-500 text-sm mt-1"}, params.UsernameErr)),
		),
		h.Div(h.KV{h.AttrClass: "space-y-1"},
			h.Label(h.KV{"for": "email", h.AttrClass: "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			h.Input(h.KV{h.AttrType: "email", h.AttrId: "email", h.AttrName: "email", h.AttrValue: params.Email, h.AttrRequired: true, h.AttrPlaceholder: "you@example.com", h.AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			h.If(params.EmailErr != nil, h.P(h.KV{h.AttrClass: "text-red-500 text-sm mt-1"}, params.EmailErr)),
		),
		h.Div(h.KV{h.AttrClass: "space-y-1"},
			h.Label(h.KV{"for": "bio", h.AttrClass: "block text-sm font-medium text-fg-secondary"}, "Bio"),
			h.Textarea(h.KV{h.AttrId: "bio", h.AttrName: "bio", h.AttrPlaceholder: "Tell us about yourself...", h.AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors min-h-24 resize-none"},
				params.Bio,
			),
			h.If(params.BioErr != nil, h.P(h.KV{h.AttrClass: "text-red-500 text-sm mt-1"}, params.BioErr)),
		),
		h.Button(h.KV{h.AttrClass: "w-full rounded-lg bg-blue hover:bg-blue/80 disabled:hover:bg-blue disabled:opacity-50 disabled:cursor-not-allowed text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2 flex items-center justify-center gap-2"},
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

	return h.Div(h.KV{h.AttrId: "sessions-tab", h.AttrClass: "space-y-3"},
		h.MapSlice(params.Sessions, func(session Session) h.Node {
			elementID := fmt.Sprintf("session-%s", session.ID)
			isCurrent := params.CurrentSessionID == session.ID
			spinnerID := fmt.Sprintf("spinner-%s", session.ID)

			return h.Div(h.KV{
				h.AttrId:    elementID,
				h.AttrClass: "flex items-center justify-between p-4 rounded-xl border-2 " + h.IfElse(isCurrent, "border-blue bg-blue/5", "border-bg-tertiary bg-bg-tertiary/30"),
			},
				// Left side: Session info
				h.Div(h.KV{h.AttrClass: "flex flex-col"},
					h.Div(h.KV{h.AttrClass: "flex items-center gap-2"},
						h.P(h.KV{h.AttrClass: "font-semibold text-fg-primary"}, session.Platform),
						h.If(isCurrent,
							h.Span(h.KV{h.AttrClass: "px-2 py-0.5 text-xs font-medium bg-blue text-bg-primary rounded-full"}, "Current"),
						),
					),
					h.P(h.KV{h.AttrClass: "text-sm text-fg-secondary"}, session.Os),
				),
				// Right side: Action button
				h.IfElse(isCurrent,
					h.Button(h.KV{
						hx.AttrHxPost:        "/logout",
						hx.AttrHxDisabledElt: "this",
						hx.AttrHxIndicator:   "#" + spinnerID,
						h.AttrClass:        "px-4 py-2 rounded-lg bg-red-500/10 hover:bg-red-500/20 text-red-500 font-medium transition-colors flex items-center gap-2 cursor-pointer",
					},
						h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>`),
						"Logout", spinner(spinnerID),
					),
					h.Button(h.KV{
						hx.AttrHxDelete:      fmt.Sprintf("/sessions/%s", session.ID),
						hx.AttrHxTarget:      "#" + elementID,
						hx.AttrHxSwap:        hx.SwapDelete,
						hx.AttrHxDisabledElt: "this",
						hx.AttrHxIndicator:   "#" + spinnerID,
						h.AttrClass:        "px-4 py-2 rounded-lg hover:bg-red-500/10 text-fg-secondary hover:text-red-500 font-medium transition-colors flex items-center gap-2 cursor-pointer",
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
		h.AttrId:                "search-modal",
		h.AttrClass:             "fixed inset-0 z-50 flex items-start justify-center bg-black/50 backdrop-blur-sm p-4 pt-20 outline-none",
		hx.AttrHxOn(h.EventClick): "if (event.target === this) this.remove()",
	},
		h.Div(h.KV{h.AttrClass: "bg-bg-secondary rounded-2xl shadow-2xl w-full max-w-xl flex flex-col overflow-hidden"},
			// Header with search input
			h.Div(h.KV{h.AttrClass: "flex items-center gap-3 p-4 border-b border-bg-tertiary"},
				h.Div(h.KV{h.AttrClass: "flex-1 relative"},
					h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="absolute left-3 top-1/2 -translate-y-1/2 text-fg-secondary"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>`),
					h.Input(h.KV{
						h.AttrType:        "text",
						h.AttrPlaceholder: "Search users...",
						h.AttrAutofocus:   true,
						h.AttrName:        "query",
						hx.AttrHxGet:        "/search/users",
						hx.AttrHxTrigger:    "input changed delay:300ms",
						hx.AttrHxTarget:     "#search-results",
						hx.AttrHxSwap:       hx.SwapInnerHtml,
						h.AttrClass:       "w-full bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary pl-10 pr-4 py-3 outline-none transition-colors",
					}),
				),
				h.Button(h.KV{
					hx.AttrHxOn(h.EventClick): "this.closest('#search-modal').remove()",
					h.AttrClass:             "p-2 hover:bg-bg-tertiary rounded-lg transition-colors cursor-pointer",
				},
					h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`),
				),
			),
			// Content area with results
			h.Div(h.KV{h.AttrId: "search-results", h.AttrClass: "flex-1 p-4 overflow-y-auto max-h-96"},
				h.P(h.KV{h.AttrClass: "text-fg-secondary text-center py-3"}, "Search for users by name or username"),
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
		return h.P(h.KV{h.AttrClass: "text-fg-secondary text-center py-3"}, "Search for users by name or username")
	}

	if len(params.Items) == 0 {
		return h.Empty()
	}

	lastID := params.Items[len(params.Items)-1].ID

	return h.Div(h.KV{h.AttrClass: "space-y-1"},
		h.MapSlice(params.Items, func(profile SearchResultItemParams) h.Node {
			return h.Div(h.KV{
				hx.AttrHxGet:     "/search/users/select/" + profile.ID,
				hx.AttrHxTrigger: "click",
				hx.AttrHxTarget:  "#chat-container",
				hx.AttrHxSwap:    hx.SwapInnerHtml,
			},
				h.IfElse(profile.ID == lastID,
					h.Div(h.KV{
						hx.AttrHxGet:     "/search/users?query=" + params.Query + "&cursor=" + lastID,
						hx.AttrHxTrigger: "intersect once",
						hx.AttrHxSwap:    hx.SwapAfterEnd,
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
	return h.Div(h.KV{h.AttrClass: "flex items-center gap-3 p-3 hover:bg-bg-tertiary/50 hover:rounded-lg cursor-pointer transition-colors"},
		h.Div(h.KV{h.AttrClass: "relative w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(params.Name),
		),
		h.Div(h.KV{h.AttrClass: "flex-1 min-w-0"},
			h.P(h.KV{h.AttrClass: "text-fg-primary font-medium truncate"}, params.Name),
			h.P(h.KV{h.AttrClass: "text-fg-secondary text-sm truncate"}, "@"+params.Username),
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
			h.AttrId:    "partners-list",
			h.AttrClass: "space-y-1",
		})
	}

	lastID := params.Partners[len(params.Partners)-1].ID

	return h.Div(h.KV{
		h.AttrId:    "partners-list",
		h.AttrClass: "space-y-1",
	},
		h.MapSlice(params.Partners, func(partner PartnerBlockParams) h.Node {
			attrs := h.IfElse(partner.ID == lastID && params.LastMessageWithLastPartnerID != "", h.KV{
				hx.AttrHxGet:       "/partners?cursor=" + params.LastMessageWithLastPartnerID,
				hx.AttrHxTrigger:   "intersect once",
				hx.AttrHxSwap:      hx.SwapAfterEnd,
				hx.AttrHxIndicator: "#partners-indicator",
				// Disable the sidebar indicator for requests in blocks
				hx.AttrHxDisinherit: hx.AttrHxIndicator,
			}, nil)

			return h.Div(attrs, PartnersListItem(partner))
		}),
	)
}

func PartnersListItem(params PartnerBlockParams) h.Node {
	return h.Div(h.KV{
		h.AttrId:       "partner-" + params.ID,
		h.AttrClass:    "cursor-pointer transition-colors",
		hx.AttrHxGet:     "/chat/" + params.ID,
		hx.AttrHxTrigger: "click",
		hx.AttrHxTarget:  "#chat-container",
		hx.AttrHxSwap:    hx.SwapInnerHtml,
		// Cancel the request if clicking the active partner
		hx.AttrHxOn(hx.EventConfigRequest): `
			if (window.currentActivePartnerId === "` + params.ID + `") {
				event.preventDefault();
				return;
			}
		`,
		hx.AttrHxOn(hx.EventLoad): `
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
		h.AttrClass: "flex items-center gap-3 p-3 hover:bg-bg-tertiary/50 hover:rounded-lg cursor-pointer transition-colors",
	},
		h.Div(h.KV{h.AttrClass: "relative w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(params.Name),
			PartnerBlockPresenceIndicator(params.ID, params.IsOnline),
		),
		h.Div(h.KV{h.AttrClass: "flex-1 min-w-0"},
			h.P(h.KV{h.AttrClass: "text-fg-primary font-medium truncate"}, params.Name),
			h.P(h.KV{h.AttrClass: "text-fg-secondary text-sm truncate"}, "@"+params.Username),
		),
		h.Span(h.KV{
			h.AttrId:    "unread-count-badge-" + params.ID,
			h.AttrClass: "flex-shrink-0 min-w-[20px] h-5 px-1.5 bg-blue text-bg-primary text-xs font-bold rounded-full flex items-center justify-center hidden",
		}),
	)
}

func PartnerBlockPresenceIndicator(partnerID string, isOnline bool) h.Node {
	return h.Div(h.KV{
		h.AttrId: "profile-block-presence-indicator-" + partnerID,
		h.AttrClass: "absolute bottom-0 right-0 w-3.5 h-3.5 bg-green rounded-full border-2 border-bg-primary " +
			h.IfElse(!isOnline, "hidden", ""),
	})
}

type ChatContainerParams struct {
	Partner PartnerBlockParams
}

func ChatContainer(params ChatContainerParams) h.Node {
	return h.Div(h.KV{
		h.AttrClass: "flex-1 flex flex-col bg-bg-primary h-full",
		hx.AttrHxOn(hx.EventLoad): `
			window.currentActivePartnerId = "` + params.Partner.ID + `";
		  document.querySelector("#partners-list [data-active]")?.removeAttribute("data-active");
		  document.getElementById("partner-" + "` + params.Partner.ID + `")?.setAttribute("data-active", "");
		`,
	},
		ChatContainerHeader(params.Partner),
		h.Div(h.KV{
			h.AttrId:    "messages-container",
			h.AttrClass: "flex-1 overflow-y-auto px-6 sm:px-10 py-4 flex flex-col-reverse gap-3",
		},
			h.Div(h.KV{
				h.AttrId: "new-message-inserter-" + params.Partner.ID,
				hx.AttrHxOn(hx.EventOobBeforeSwap): `
					// don't insert the new message if it doesn't come from the active partner
					if (window.currentActivePartnerId !== "` + params.Partner.ID + `") {
						event.preventDefault();
						return;
					}
				`,
			}),
			h.Div(h.KV{
				hx.AttrHxGet:                     fmt.Sprintf("/chat/%s/messages", params.Partner.ID),
				hx.AttrHxTrigger:                 "load",
				hx.AttrHxSwap:                    hx.SwapAfterEnd,
				hx.AttrHxIndicator:               "#messages-indicator",
				hx.AttrHxOn(hx.EventAfterSettle): "this.remove()",
			}),
			h.Div(h.KV{h.AttrClass: "flex justify-center"},
				spinner("messages-indicator"),
			),
		),
		ChatInputForm(ChatInputFormParams{PartnerID: params.Partner.ID}),
	)
}

func ChatContainerHeader(partner PartnerBlockParams) h.Node {
	return h.Div(h.KV{
		h.AttrId:    "chat-container-header-" + partner.ID,
		h.AttrClass: "h-16 px-4 bg-bg-secondary border-b border-bg-tertiary flex items-center gap-3",
	},
		h.Div(h.KV{h.AttrClass: "relative w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(partner.Name),
			ChatContainerPresenceIndicator(partner.ID, partner.IsOnline),
		),
		h.Div(h.KV{h.AttrClass: "flex-1 min-w-0"},
			h.P(h.KV{h.AttrClass: "text-fg-primary font-medium truncate"}, partner.Name),
			h.P(h.KV{h.AttrClass: "text-fg-secondary text-sm truncate"}, "@"+partner.Username),
		),
	)
}

func ChatContainerPresenceIndicator(partnerID string, isOnline bool) h.Node {
	return h.Div(h.KV{
		h.AttrId: "chat-container-presence-indicator-" + partnerID,
		h.AttrClass: "absolute bottom-0 right-0 w-3.5 h-3.5 bg-green rounded-full border-2 border-bg-primary " +
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
					hx.AttrHxGet:       fmt.Sprintf("/chat/%s/messages?cursor=%s", params.PartnerID, cursorMessageID),
					hx.AttrHxTrigger:   "intersect once",
					hx.AttrHxSwap:      hx.SwapAfterEnd,
					hx.AttrHxIndicator: "#messages-indicator",
					// For internal requests not to use this indicator
					hx.AttrHxDisinherit: hx.AttrHxIndicator,
				},
				nil,
			),
			ChatMessage(msg),
		)
	})
}

type ChatMessageStatus int

const (
	StatusPending ChatMessageStatus = iota
	StatusSent
	StatusRead
)

type ChatMessageParams struct {
	ID              string
	ClientMessageID int
	PartnerID       string
	Content         string
	SentAt          time.Time
	Status          ChatMessageStatus
	FromMe          bool
}

func ChatMessage(params ChatMessageParams) h.Node {
	return h.Div(
		h.IfElse(!params.FromMe && params.Status == StatusSent,
			h.KV{
				hx.AttrHxPost:    "/api/v1/chats/" + params.PartnerID + "/mark_as_read?upto_message_id=" + params.ID,
				hx.AttrHxTrigger: "intersect once",
				hx.AttrHxSwap:    hx.SwapNone,
			},
			h.KV{h.AttrId: fmt.Sprintf("pending-chat-message-%d", params.ClientMessageID)},
		),
		h.Div(h.KV{h.AttrClass: "flex w-full " + h.IfElse(params.FromMe, "justify-end", "justify-start")},
			h.Div(h.KV{h.AttrClass: "flex flex-col " + h.IfElse(params.FromMe, "items-end", "items-start") + " max-w-[70%]"},
				h.Div(h.KV{h.AttrClass: "px-4 py-2 " + h.IfElse(params.FromMe, "bg-blue text-bg-primary rounded-l-2xl rounded-tr-2xl", "bg-bg-tertiary text-fg-primary rounded-r-2xl rounded-tl-2xl")},
					h.P(h.KV{h.AttrClass: "whitespace-pre-wrap"}, params.Content),
					h.Div(h.KV{h.AttrClass: "flex items-center gap-1 mt-1 justify-end"},
						h.If(params.Status != StatusPending,
							h.P(h.KV{h.AttrClass: "text-xs " + h.IfElse(params.FromMe, "text-bg-primary/70", "text-fg-secondary")}, params.SentAt.Format("Jan 2, 3:04 PM")),
						),
						h.If(params.FromMe,
							h.IfElse(params.Status == StatusPending,
								PendingMessageIndicator(),
								h.IfElse(params.Status == StatusSent,
									sentMessageIndicator(params.ID),
									ReadMessageIndicator(),
								),
							),
						),
					),
				),
			),
		),
	)
}

func PendingMessageIndicator() h.Node {
	return h.Span(h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-bg-primary animate-spin" style="animation-duration: 1s;"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>`))
}

func ReadMessageIndicator() h.Node {
	return h.Span(h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-bg-primary"><polyline points="16 6 8 14 4 10"></polyline><polyline points="22 6 14 14 10 10"></polyline></svg>`))
}

func sentMessageIndicator(messageID string) h.Node {
	return h.Span(h.RawText(fmt.Sprintf(`
		<svg id="unread-message-indicator-%s" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-bg-primary"><polyline points="20 6 9 17 4 12"></polyline></svg>`,
		messageID,
	)))
}

type ChatInputFormParams struct {
	PartnerID string
}

func ChatInputForm(params ChatInputFormParams) h.Node {
	return h.Form(h.KV{
		hxws.AttrWsSend: true,
		hx.AttrHxOn(hxws.EventWsConfigSend): `
			const content = event.detail.parameters.content.trim();
			if (content == "") {
				event.preventDefault();
				return;
			}
			event.detail.parameters.content = content;
			event.detail.parameters.clientMessageID = window.messageIDGenerator.GetID();
		`,
		hx.AttrHxOn(hxws.EventWsAfterSend): `
			const textarea = event.detail.elt.querySelector('textarea[name="content"]');
			textarea.value = '';
			textarea.style.height = 'auto';
		`,
		h.AttrClass: "flex items-center gap-2 px-3 py-2",
	},
		h.Input(h.KV{h.AttrType: h.TypeHidden, h.AttrName: "kind", h.AttrValue: "SendMessage"}),
		h.Input(h.KV{h.AttrType: h.TypeHidden, h.AttrName: "partnerID", h.AttrValue: params.PartnerID}),
		h.Textarea(h.KV{
			h.AttrClass:       "flex-1 bg-bg-tertiary text-fg-primary resize-none outline-none max-h-[120px] min-h-[44px] py-2.5 px-4 rounded-2xl leading-6 self-center overflow-y-auto",
			h.AttrName:        "content",
			h.AttrPlaceholder: "Write a message...",
			h.AttrRows:        "1",
			h.AttrAutofocus:   true,
			h.AttrOnInput: `
				this.style.height = "auto";
				this.style.height = Math.min(this.scrollHeight, 120)+"px";
			`,
			h.AttrOnKeyDown: `
				if (event.key === "Enter" && !event.shiftKey) {
					event.preventDefault();
					document.getElementById("send-btn").click();
				}
			`,
		}),
		h.Button(h.KV{
			h.AttrId:    "send-btn",
			h.AttrType:  "submit",
			h.AttrClass: "w-12 h-12 rounded-full bg-bg-primary hover:bg-bg-tertiary flex items-center justify-center flex-shrink-0 cursor-pointer transition-colors",
		},
			h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-send-horizontal-icon lucide-send-horizontal text-blue"><path d="M3.714 3.048a.498.498 0 0 0-.683.627l2.843 7.627a2 2 0 0 1 0 1.396l-2.842 7.627a.498.498 0 0 0 .682.627l18-8.5a.5.5 0 0 0 0-.904z"/><path d="M6 12h16"/></svg>`),
		),
	)
}
