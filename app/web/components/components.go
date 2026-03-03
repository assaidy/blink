package components

import (
	"fmt"
	"slices"
	"strings"
	"time"

	. "github.com/assaidy/hyper"
	. "github.com/assaidy/hyper/htmx"
	. "github.com/assaidy/hyper/htmx/extensions/ws"
)

func rootLayout(children ...any) HyperNode {
	return Empty(
		DoctypeHtml(),
		Html(
			Head(
				Title("blink"),
				Meta(KV{AttrCharset: "UTF-8"}),
				Meta(KV{AttrName: "viewport", AttrContent: "width=device-width, initial-scale=1.0"}),
				Link(KV{AttrRel: "stylesheet", AttrHref: "/public/css/style.css"}),
				Script(KV{AttrSrc: "/public/js/lib/htmx@2.0.8.js"}),
				Script(RawText(`
					const $cookie = (name) => {
						return document.cookie
							.split('; ')
							.find(row => row.startsWith(name+"="))
							?.split('=')[1]
							?.trim() || '';
					};
				`)),
			),
			Body(KV{
				AttrClass: "bg-bg-primary text-fg-primary",
				AttrHxOn(EventHtmxConfigRequest): `
					event.detail.headers['X-CSRF-Token'] = $cookie("csrf_token");
				`,
			},
				Div(children...),
			),
		),
	)
}

func RegisterPage() HyperNode {
	return rootLayout(
		Div(KV{AttrClass: "min-h-screen flex justify-center items-center bg-bg-primary sm:px-6 lg:px-8"},
			Div(KV{AttrClass: "w-full min-h-screen sm:min-h-0 sm:max-w-md sm:p-8 bg-bg-secondary sm:rounded-lg sm:shadow-lg flex flex-col justify-center"},
				Div(KV{AttrClass: "p-6 sm:p-0"},
					H2(KV{AttrClass: "text-fg-primary text-2xl font-bold text-center mb-8"}, "Create Account"),
					RegisterForm(),
					P(KV{AttrClass: "text-center text-fg-secondary text-sm mt-6"},
						"Already have an account? ", A(KV{AttrHref: "/login", AttrClass: "text-blue hover:underline font-medium"}, "Sign in"),
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

func RegisterForm(params ...RegisterFormParams) HyperNode {
	var p RegisterFormParams
	if len(params) != 0 {
		p = params[0]
	}

	return Form(KV{AttrHxPost: "/register", AttrHxSwap: SwapOuterHtml, AttrClass: "space-y-5"},
		Div(KV{AttrClass: "space-y-1"},
			Label(KV{"for": "name", AttrClass: "block text-sm font-medium text-fg-secondary"}, "Full Name"),
			Input(KV{AttrType: "text", AttrId: "name", AttrName: "name", AttrRequired: true, AttrValue: p.Name, AttrPlaceholder: "Enter your full name", AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			If(p.NameErr != nil, P(KV{AttrClass: "text-red-500 text-sm mt-1"}, p.NameErr)),
		),
		Div(KV{AttrClass: "space-y-1"},
			Label(KV{"for": "username", AttrClass: "block text-sm font-medium text-fg-secondary"}, "Username"),
			Input(KV{AttrType: "text", AttrId: "username", AttrName: "username", AttrValue: p.Username, AttrPlaceholder: "Choose a username", AttrRequired: true, AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			If(p.UsernameErr != nil, P(KV{AttrClass: "text-red-500 text-sm mt-1"}, p.UsernameErr)),
		),
		Div(KV{AttrClass: "space-y-1"},
			Label(KV{"for": "email", AttrClass: "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			Input(KV{AttrType: "email", AttrId: "email", AttrName: "email", AttrValue: p.Email, AttrRequired: true, AttrPlaceholder: "you@example.com", AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			If(p.EmailErr != nil, P(KV{AttrClass: "text-red-500 text-sm mt-1"}, p.EmailErr)),
		),
		Div(KV{AttrClass: "space-y-1"},
			Label(KV{"for": "bio", AttrClass: "block text-sm font-medium text-fg-secondary"}, "Bio"),
			Textarea(KV{AttrId: "bio", AttrName: "bio", AttrPlaceholder: "Tell us about yourself...", AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors min-h-24 resize-none"},
				p.Bio,
			),
			If(p.BioErr != nil, P(KV{AttrClass: "text-red-500 text-sm mt-1"}, p.BioErr)),
		),
		Button(KV{AttrClass: "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
			"Create Account",
		),
	)
}

func LoginPage() HyperNode {
	return rootLayout(
		Div(KV{AttrClass: "min-h-screen flex justify-center items-center bg-bg-primary sm:px-6 lg:px-8"},
			Div(KV{AttrClass: "w-full min-h-screen sm:min-h-0 sm:max-w-md sm:p-8 bg-bg-secondary sm:rounded-lg sm:shadow-lg flex flex-col justify-center"},
				Div(KV{AttrClass: "p-6 sm:p-0"},
					H2(KV{AttrClass: "text-fg-primary text-2xl font-bold text-center mb-8"}, "Sign In"),
					LoginForm(),
					P(KV{AttrClass: "text-center text-fg-secondary text-sm mt-6"},
						"Don't have an account? ", A(KV{AttrHref: "/register", AttrClass: "text-blue hover:underline font-medium"}, "Create one"),
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

func LoginForm(params ...LoginFormParams) HyperNode {
	var p LoginFormParams
	if len(params) != 0 {
		p = params[0]
	}

	return Form(KV{AttrHxPost: "/login", AttrHxSwap: SwapOuterHtml, AttrClass: "space-y-5"},
		Div(KV{AttrClass: "space-y-1"},
			Label(KV{"for": "email", AttrClass: "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			Input(KV{AttrType: "email", AttrId: "email", AttrName: "email", AttrValue: p.Email, AttrRequired: true, AttrPlaceholder: "you@example.com", AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			If(p.EmailErr != nil, P(KV{AttrClass: "text-red-500 text-sm mt-1"}, p.EmailErr)),
		),
		Button(KV{AttrClass: "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
			"Sign In",
		),
	)
}

type OtpFormParams struct {
	OtpID  string
	Otp    string
	OtpErr any
}

func OtpForm(params OtpFormParams) HyperNode {
	return Form(KV{AttrHxPost: "/verify_otp", AttrHxSwap: SwapOuterHtml, AttrClass: "space-y-5"},
		Input(KV{AttrType: TypeHidden, AttrName: "otpID", AttrValue: params.OtpID}),
		Div(KV{AttrClass: "space-y-1"},
			Label(KV{"for": "otp", AttrClass: "block text-sm font-medium text-fg-secondary"}, "Verification Code"),
			P(KV{AttrClass: "text-fg-secondary text-sm mb-2"}, "We've sent a 6-digit code to your email address. Please enter it below to verify your identity."),
			Input(KV{
				AttrType:         "text",
				AttrId:           "otp",
				AttrName:         "otp",
				AttrValue:        params.Otp,
				AttrRequired:     true,
				AttrMaxLength:    "6",
				AttrPattern:      "[0-9]{6}",
				AttrInputMode:    "numeric",
				AttrAutocomplete: "one-time-code",
				AttrPlaceholder:  "000000",
				AttrClass:        "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors text-center text-2xl tracking-widest",
			}),
			If(params.OtpErr != nil, P(KV{AttrClass: "text-red-500 text-sm mt-1"}, params.OtpErr)),
		),
		Button(KV{AttrClass: "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
			"Verify",
		),
	)
}

type ChatPageParams struct {
	User UserBlockParams
}

func ChatPage(params ChatPageParams) HyperNode {
	return rootLayout(
		Script(KV{AttrSrc: "/public/js/lib/htmx_ext_ws@2.0.4.js"}),
		Script(RawText(`
			htmx.createWebSocket = (url) => {
				const csrfToken = $cookie("csrf_token");
				const fullUrl = csrfToken ? url+"?csrf_token="+csrfToken : url;
				return new WebSocket(fullUrl);
			};
		`)),
		Script(RawText(`
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
		Div(KV{AttrId: "unread-manager-anchor"}),
		// Actual page content
		Div(KV{
			AttrClass:     "h-screen flex bg-bg-primary",
			AttrHxExt:     "ws",
			AttrWsConnect: "/ws",
			AttrOnClick: `
				const msgMenu = document.getElementById('message-context-menu');
				const partnerMenu = document.getElementById('partner-context-menu');
				if (msgMenu && !msgMenu.classList.contains('hidden')) msgMenu.classList.add('hidden');
				if (partnerMenu && !partnerMenu.classList.contains('hidden')) partnerMenu.classList.add('hidden');
			`,
		},
			partnerContextMenu(),
			messageContextMenu(),
			// Sidebar
			Div(KV{AttrClass: "w-80 bg-bg-secondary border-r border-bg-tertiary flex flex-col"},
				// Sticky top row
				Div(KV{AttrClass: "sticky top-0 z-10 h-16 bg-aqua/10 border-b border-bg-primary flex items-center"},
					UserBlock(params.User),
					Button(KV{
						AttrHxGet:     "/search_modal",
						AttrHxTrigger: "click",
						AttrHxTarget:  "body",
						AttrHxSwap:    SwapBeforeEnd,
						AttrClass:     "flex items-center justify-center w-16 h-16 flex-shrink-0 hover:bg-aqua/30 transition-colors cursor-pointer",
					},
						RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>`),
					),
				),
				// Partners list container
				Div(KV{
					AttrId:    "partners-container",
					AttrClass: "flex-1 overflow-y-auto px-2 py-2",
				},
					Div(KV{
						AttrHxGet:                      "/partners",
						AttrHxTrigger:                  "load",
						AttrHxSwap:                     SwapAfterEnd,
						AttrHxIndicator:                "#partners-indicator",
						AttrHxOn(EventHtmxAfterSettle): "document.getElementById('partners-container').scrollTop = 0; this.remove()",
					}),
					Div(KV{AttrClass: "flex justify-center"},
						spinner("partners-indicator"),
					),
				),
			),
			Div(KV{
				AttrId:    "chat-container",
				AttrClass: "flex-1 flex flex-col bg-bg-primary",
			},
				ChatContainerPlaceholder(),
			),
		),
	)
}

func ChatContainerPlaceholder() HyperNode {
	return Div(KV{AttrClass: "flex-1 flex items-center justify-center"},
		P(KV{AttrClass: "text-fg-secondary text-lg"}, "Select a chat to start messaging"),
	)
}

func spinner(id string) HyperNode {
	return RawText(fmt.Sprintf(`<svg
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

func UserBlock(params UserBlockParams) HyperNode {
	return Div(KV{
		AttrId:        "user-block",
		AttrHxGet:     "/profile_modal",
		AttrHxTrigger: "click",
		AttrHxTarget:  "body",
		AttrHxSwap:    SwapBeforeEnd,
		AttrClass:     "w-64 h-16 flex-shrink-0 flex items-center gap-3 cursor-pointer hover:bg-aqua/20 transition-colors px-4",
	},
		Div(KV{AttrClass: "w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold flex-shrink-0"},
			getInitials(params.Name),
		),
		Div(KV{AttrClass: "flex-1 min-w-0"},
			P(KV{AttrClass: "text-fg-primary font-medium truncate"}, params.Name),
			P(KV{AttrClass: "text-fg-secondary text-sm truncate"}, "@"+params.Username),
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

func ProfileModal(params ProfileModalParams) HyperNode {
	if params.ActiveTab == tabDefault {
		params.ActiveTab = TabProfile
	}

	profileIcon := RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>`)
	sessionsIcon := RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>`)

	return Div(KV{
		AttrId:      "profile-modal",
		AttrClass:   "fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 outline-none",
		AttrOnClick: "if (event.target === this) this.remove()",
	},
		Div(KV{AttrClass: "bg-bg-secondary rounded-2xl shadow-2xl max-w-3xl w-full flex overflow-hidden", AttrStyle: "height: 80vh; max-height: 700px;"},
			// Left sidebar with tabs
			Div(KV{AttrClass: "w-56 bg-bg-tertiary flex flex-col p-2 flex-shrink-0"},
				// Close button at top
				Button(KV{
					AttrClass:   "self-start p-2 hover:bg-bg-secondary rounded-lg transition-colors cursor-pointer mb-6",
					AttrOnClick: "this.closest('#profile-modal').remove()",
				},
					RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`),
				),
				// Tabs
				Div(KV{AttrClass: "flex flex-col gap-1"},
					Button(KV{
						AttrHxGet:     "/profile_modal?tab=profile",
						AttrHxTarget:  "#profile-modal",
						AttrHxSwap:    SwapOuterHtml,
						AttrHxTrigger: IfElse(params.ActiveTab == TabProfile, SwapNone, "click"),
						AttrClass:     "flex items-center gap-3 px-2 py-2.5 rounded-lg text-left font-medium text-fg-primary " + IfElse(params.ActiveTab == TabProfile, "bg-bg-secondary", "hover:bg-bg-secondary/50"),
					},
						profileIcon, "Profile",
					),
					Button(KV{
						AttrHxGet:     "/profile_modal?tab=sessions",
						AttrHxTarget:  "#profile-modal",
						AttrHxSwap:    SwapOuterHtml,
						AttrHxTrigger: IfElse(params.ActiveTab == TabSessions, SwapNone, "click"),
						AttrClass:     "flex items-center gap-3 px-2 py-2.5 rounded-lg text-left font-medium text-fg-primary " + IfElse(params.ActiveTab == TabSessions, "bg-bg-secondary", "hover:bg-bg-secondary/50"),
					},
						sessionsIcon, "Sessions",
					),
				),
			),
			// Right content area
			Div(KV{AttrClass: "flex-1 flex flex-col min-w-0"},
				// Header
				Div(KV{AttrClass: "px-8 py-6 border-b border-bg-tertiary"},
					H2(KV{AttrClass: "text-xl font-semibold text-fg-primary"},
						IfElse(params.ActiveTab == TabProfile, "Profile", "Sessions"),
					),
				),
				// Content
				Div(KV{AttrId: "tab-content", AttrClass: "flex-1 p-8 overflow-y-auto"},
					IfElse(params.ActiveTab == TabProfile,
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

func ProfileTab(params ProfileTabParams) HyperNode {
	return Div(
		ProfileForm(ProfileFormParams{
			Name:     params.Name,
			Username: params.Username,
			Email:    params.Email,
			Bio:      params.Bio,
		}),
		P(KV{AttrClass: "text-fg-secondary text-sm text-center my-4"},
			"Joined ", params.JoinedAt.Format("January 2, 2006"),
		),
		Div(KV{AttrClass: "flex items-center justify-center gap-2 text-sm"},
			IfElse(params.EmailIsVerified,
				Empty(
					RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-green-500"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>`),
					P(KV{AttrClass: "text-green-600"}, "Email verified"),
				),
				Empty(
					RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-yellow-500"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>`),
					P(KV{AttrClass: "text-yellow-600"}, "Email not verified"),
				),
			),
		),
		Div(KV{AttrClass: "border-t border-bg-tertiary my-6"}),
		Button(KV{
			AttrHxDelete:  "/profile",
			AttrHxSwap:    SwapNone,
			AttrHxConfirm: "Are you sure you wish to delete your account? This action cannot be undone.",
			AttrClass:     "w-full px-4 py-2 rounded-lg !bg-red-600 hover:!bg-red-700 text-white font-medium transition-colors flex items-center justify-center gap-2 cursor-pointer",
		},
			RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>`),
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

func ProfileForm(params ProfileFormParams) HyperNode {
	return Form(KV{
		AttrHxPut:         "/profile",
		AttrHxSwap:        SwapOuterHtml,
		AttrHxDisabledElt: "find button",
		AttrHxIndicator:   "#spinner",
		AttrClass:         "space-y-5",
	},
		Div(KV{AttrClass: "space-y-1"},
			Label(KV{"for": "name", AttrClass: "block text-sm font-medium text-fg-secondary"}, "Full Name"),
			Input(KV{AttrType: "text", AttrId: "name", AttrName: "name", AttrRequired: true, AttrValue: params.Name, AttrPlaceholder: "Enter your full name", AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			If(params.NameErr != nil, P(KV{AttrClass: "text-red-500 text-sm mt-1"}, params.NameErr)),
		),
		Div(KV{AttrClass: "space-y-1"},
			Label(KV{"for": "username", AttrClass: "block text-sm font-medium text-fg-secondary"}, "Username"),
			Input(KV{AttrType: "text", AttrId: "username", AttrName: "username", AttrValue: params.Username, AttrPlaceholder: "Choose a username", AttrRequired: true, AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			If(params.UsernameErr != nil, P(KV{AttrClass: "text-red-500 text-sm mt-1"}, params.UsernameErr)),
		),
		Div(KV{AttrClass: "space-y-1"},
			Label(KV{"for": "email", AttrClass: "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			Input(KV{AttrType: "email", AttrId: "email", AttrName: "email", AttrValue: params.Email, AttrRequired: true, AttrPlaceholder: "you@example.com", AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			If(params.EmailErr != nil, P(KV{AttrClass: "text-red-500 text-sm mt-1"}, params.EmailErr)),
		),
		Div(KV{AttrClass: "space-y-1"},
			Label(KV{"for": "bio", AttrClass: "block text-sm font-medium text-fg-secondary"}, "Bio"),
			Textarea(KV{AttrId: "bio", AttrName: "bio", AttrPlaceholder: "Tell us about yourself...", AttrClass: "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors min-h-24 resize-none"},
				params.Bio,
			),
			If(params.BioErr != nil, P(KV{AttrClass: "text-red-500 text-sm mt-1"}, params.BioErr)),
		),
		Button(KV{AttrClass: "w-full rounded-lg bg-blue hover:bg-blue/80 disabled:hover:bg-blue disabled:opacity-50 disabled:cursor-not-allowed text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2 flex items-center justify-center gap-2"},
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

func SessionsTab(params SessionsTabParams) HyperNode {
	// Put the current session first
	if index := slices.IndexFunc(params.Sessions, func(s Session) bool {
		return s.ID == params.CurrentSessionID
	}); index != -1 {
		params.Sessions[0], params.Sessions[index] = params.Sessions[index], params.Sessions[0]
	}

	return Div(KV{AttrId: "sessions-tab", AttrClass: "space-y-3"},
		MapSlice(params.Sessions, func(session Session) HyperNode {
			elementID := fmt.Sprintf("session-%s", session.ID)
			isCurrent := params.CurrentSessionID == session.ID
			spinnerID := fmt.Sprintf("spinner-%s", session.ID)

			return Div(KV{
				AttrId:    elementID,
				AttrClass: "flex items-center justify-between p-4 rounded-xl border-2 " + IfElse(isCurrent, "border-blue bg-blue/5", "border-bg-tertiary bg-bg-tertiary/30"),
			},
				// Left side: Session info
				Div(KV{AttrClass: "flex flex-col"},
					Div(KV{AttrClass: "flex items-center gap-2"},
						P(KV{AttrClass: "font-semibold text-fg-primary"}, session.Platform),
						If(isCurrent,
							Span(KV{AttrClass: "px-2 py-0.5 text-xs font-medium bg-blue text-bg-primary rounded-full"}, "Current"),
						),
					),
					P(KV{AttrClass: "text-sm text-fg-secondary"}, session.Os),
				),
				// Right side: Action button
				IfElse(isCurrent,
					Button(KV{
						AttrHxPost:        "/logout",
						AttrHxDisabledElt: "this",
						AttrHxIndicator:   "#" + spinnerID,
						AttrClass:         "px-4 py-2 rounded-lg bg-red-500/10 hover:bg-red-500/20 text-red-500 font-medium transition-colors flex items-center gap-2 cursor-pointer",
					},
						RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>`),
						"Logout", spinner(spinnerID),
					),
					Button(KV{
						AttrHxDelete:      fmt.Sprintf("/sessions/%s", session.ID),
						AttrHxTarget:      "#" + elementID,
						AttrHxSwap:        SwapDelete,
						AttrHxDisabledElt: "this",
						AttrHxIndicator:   "#" + spinnerID,
						AttrClass:         "px-4 py-2 rounded-lg hover:bg-red-500/10 text-fg-secondary hover:text-red-500 font-medium transition-colors flex items-center gap-2 cursor-pointer",
					},
						RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>`),
						"Remove", spinner(spinnerID),
					),
				),
			)
		}),
	)
}

func SearchModal() HyperNode {
	return Div(KV{
		AttrId:      "search-modal",
		AttrClass:   "fixed inset-0 z-50 flex items-start justify-center bg-black/50 backdrop-blur-sm p-4 pt-20 outline-none",
		AttrOnClick: "if (event.target === this) this.remove();",
	},
		Div(KV{AttrClass: "bg-bg-secondary rounded-2xl shadow-2xl w-full max-w-xl flex flex-col overflow-hidden"},
			// Header with search input
			Div(KV{AttrClass: "flex items-center gap-3 p-4 border-b border-bg-tertiary"},
				Div(KV{AttrClass: "flex-1 relative"},
					RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="absolute left-3 top-1/2 -translate-y-1/2 text-fg-secondary"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>`),
					Input(KV{
						AttrType:        "text",
						AttrPlaceholder: "Search users...",
						AttrAutofocus:   true,
						AttrName:        "query",
						AttrHxGet:       "/search/users",
						AttrHxTrigger:   "input changed delay:300ms",
						AttrHxTarget:    "#search-results",
						AttrHxSwap:      SwapInnerHtml,
						AttrClass:       "w-full bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary pl-10 pr-4 py-3 outline-none transition-colors",
					}),
				),
				Button(KV{
					AttrOnClick: "this.closest('#search-modal').remove();",
					AttrClass:   "p-2 hover:bg-bg-tertiary rounded-lg transition-colors cursor-pointer",
				},
					RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-fg-secondary"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`),
				),
			),
			// Content area with results
			Div(KV{AttrId: "search-results", AttrClass: "flex-1 p-4 overflow-y-auto max-h-96"},
				P(KV{AttrClass: "text-fg-secondary text-center py-3"}, "Search for users by name or username"),
			),
		),
	)
}

type SearchResultParams struct {
	Query   string
	HasMore bool
	Items   []SearchResultItemParams
}

func SearchResult(params SearchResultParams) HyperNode {
	if params.Query == "" {
		return P(KV{AttrClass: "text-fg-secondary text-center py-3"}, "Search for users by name or username")
	}

	if len(params.Items) == 0 {
		return Empty()
	}

	lastID := params.Items[len(params.Items)-1].ID

	return Div(KV{AttrClass: "space-y-1"},
		MapSlice(params.Items, func(profile SearchResultItemParams) HyperNode {
			return Div(KV{
				AttrHxGet:     "/search/users/select/" + profile.ID,
				AttrHxTrigger: "click",
				AttrHxTarget:  "#chat-container",
				AttrHxSwap:    SwapInnerHtml,
			},
				IfElse(profile.ID == lastID,
					Div(KV{
						AttrHxGet:     "/search/users?query=" + params.Query + "&cursor=" + lastID,
						AttrHxTrigger: "intersect once",
						AttrHxSwap:    SwapAfterEnd,
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

func searchResultItem(params SearchResultItemParams) HyperNode {
	return Div(KV{AttrClass: "flex items-center gap-3 p-3 hover:bg-bg-tertiary/50 hover:rounded-lg cursor-pointer transition-colors"},
		Div(KV{AttrClass: "relative w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(params.Name),
		),
		Div(KV{AttrClass: "flex-1 min-w-0"},
			P(KV{AttrClass: "text-fg-primary font-medium truncate"}, params.Name),
			P(KV{AttrClass: "text-fg-secondary text-sm truncate"}, "@"+params.Username),
		),
	)
}

type PartnersListParams struct {
	Partners []PartnerBlockParams
	// This is the cursor. Empty when no more partners
	LastMessageWithLastPartnerID string
}

func PartnersList(params PartnersListParams) HyperNode {
	if len(params.Partners) == 0 {
		return Div(KV{AttrId: "partners-list", AttrClass: "space-y-1"})
	}

	lastID := params.Partners[len(params.Partners)-1].ID

	return Div(KV{AttrId: "partners-list", AttrClass: "space-y-1"},
		MapSlice(params.Partners, func(partner PartnerBlockParams) HyperNode {
			return Div(
				IfElse(partner.ID == lastID && params.LastMessageWithLastPartnerID != "",
					KV{
						AttrHxGet:       "/partners?cursor=" + params.LastMessageWithLastPartnerID,
						AttrHxTrigger:   "intersect once",
						AttrHxSwap:      SwapAfterEnd,
						AttrHxIndicator: "#partners-indicator",
						// Disable the sidebar indicator for requests in blocks
						AttrHxDisinherit: AttrHxIndicator,
					},
					nil,
				),
				PartnersListItem(partner),
			)
		}),
	)
}

func PartnersListItem(params PartnerBlockParams) HyperNode {
	return Div(KV{
		AttrId:        "partner-" + params.ID,
		AttrClass:     "cursor-pointer transition-colors",
		AttrHxGet:     "/chat/" + params.ID,
		AttrHxTrigger: "click",
		AttrHxTarget:  "#chat-container",
		AttrHxSwap:    SwapInnerHtml,
		// Cancel the request if clicking the active partner
		AttrHxOn(EventHtmxConfigRequest): `
			if (window.currentActivePartnerId === "` + params.ID + `") {
				event.preventDefault();
				return;
			}
		`,
		AttrHxOn(EventHtmxLoad): `
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

func partnerBlock(params PartnerBlockParams) HyperNode {
	return Div(KV{
		AttrClass: "flex items-center gap-3 p-3 hover:bg-bg-tertiary/50 hover:rounded-lg cursor-pointer transition-colors",
		AttrOnContextMenu: `
			event.preventDefault();
			const menu = document.getElementById('partner-context-menu');
			menu.dataset.partnerID = "` + params.ID + `";
			menu.style.left = event.clientX + 'px';
			menu.style.top = event.clientY + 'px';
			menu.classList.remove('hidden');
			const menuRect = menu.getBoundingClientRect();
			const viewportWidth = window.innerWidth;
			const viewportHeight = window.innerHeight;
			if (event.clientX + menuRect.width > viewportWidth) menu.style.left = (event.clientX - menuRect.width) + 'px';
			if (event.clientY + menuRect.height > viewportHeight) menu.style.top = (event.clientY - menuRect.height) + 'px';
		`,
	},
		Div(KV{AttrClass: "relative w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(params.Name),
			PartnerBlockPresenceIndicator(params.ID, params.IsOnline),
		),
		Div(KV{AttrClass: "flex-1 min-w-0"},
			P(KV{AttrClass: "text-fg-primary font-medium truncate"}, params.Name),
			P(KV{AttrClass: "text-fg-secondary text-sm truncate"}, "@"+params.Username),
		),
		Span(KV{
			AttrId:    "unread-count-badge-" + params.ID,
			AttrClass: "flex-shrink-0 min-w-[20px] h-5 px-1.5 bg-blue text-bg-primary text-xs font-bold rounded-full flex items-center justify-center hidden",
		}),
	)
}

func PartnerBlockPresenceIndicator(partnerID string, isOnline bool) HyperNode {
	return Div(KV{
		AttrId: "profile-block-presence-indicator-" + partnerID,
		AttrClass: "absolute bottom-0 right-0 w-3.5 h-3.5 bg-green rounded-full border-2 border-bg-primary " +
			IfElse(!isOnline, "hidden", ""),
	})
}

type ChatContainerParams struct {
	Partner PartnerBlockParams
}

func ChatContainer(params ChatContainerParams) HyperNode {
	return Div(KV{
		AttrId:    "chat-container-" + params.Partner.ID,
		AttrClass: "flex-1 flex flex-col bg-bg-primary h-full",
		AttrHxOn(EventHtmxLoad): `
			window.currentActivePartnerId = "` + params.Partner.ID + `";
		  document.querySelector("#partners-list [data-active]")?.removeAttribute("data-active");
		  document.getElementById("partner-" + "` + params.Partner.ID + `")?.setAttribute("data-active", "");
		`,
	},
		ChatContainerHeader(params.Partner),
		Div(KV{
			AttrId:    "messages-container",
			AttrClass: "flex-1 overflow-y-auto px-6 sm:px-10 py-4 flex flex-col-reverse gap-3",
		},
			Div(KV{
				AttrId: "new-message-inserter-" + params.Partner.ID,
				AttrHxOn(EventHtmxOobBeforeSwap): `
					// don't insert the new message if it doesn't come from the active partner
					if (window.currentActivePartnerId !== "` + params.Partner.ID + `") {
						event.preventDefault();
						return;
					}
				`,
			}),
			Div(KV{
				AttrHxGet:                      fmt.Sprintf("/chat/%s/messages", params.Partner.ID),
				AttrHxTrigger:                  "load",
				AttrHxSwap:                     SwapAfterEnd,
				AttrHxIndicator:                "#messages-indicator",
				AttrHxOn(EventHtmxAfterSettle): "this.remove()",
			}),
			Div(KV{AttrClass: "flex justify-center"},
				spinner("messages-indicator"),
			),
		),
		ChatInputForm(ChatInputFormParams{PartnerID: params.Partner.ID}),
	)
}

func ChatContainerHeader(partner PartnerBlockParams) HyperNode {
	return Div(KV{
		AttrId:    "chat-container-header-" + partner.ID,
		AttrClass: "h-16 px-4 bg-bg-secondary border-b border-bg-tertiary flex items-center gap-3",
	},
		Div(KV{AttrClass: "relative w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(partner.Name),
			ChatContainerPresenceIndicator(partner.ID, partner.IsOnline),
		),
		Div(KV{AttrClass: "flex-1 min-w-0"},
			P(KV{AttrClass: "text-fg-primary font-medium truncate"}, partner.Name),
			P(KV{AttrClass: "text-fg-secondary text-sm truncate"}, "@"+partner.Username),
		),
	)
}

func ChatContainerPresenceIndicator(partnerID string, isOnline bool) HyperNode {
	return Div(KV{
		AttrId: "chat-container-presence-indicator-" + partnerID,
		AttrClass: "absolute bottom-0 right-0 w-3.5 h-3.5 bg-green rounded-full border-2 border-bg-primary " +
			IfElse(!isOnline, "hidden", ""),
	})
}

type ChatMessagesListParams struct {
	PartnerID string
	Messages  []ChatMessageParams
	HasMore   bool
}

func ChatMessagesList(params ChatMessagesListParams) HyperNode {
	if len(params.Messages) == 0 {
		return Empty()
	}

	cursorMessageID := params.Messages[len(params.Messages)-1].ID

	return MapSlice(params.Messages, func(msg ChatMessageParams) HyperNode {
		msg.PartnerID = params.PartnerID
		return IfElse(params.HasMore && msg.ID == cursorMessageID,
			Div(KV{
				AttrHxGet:       fmt.Sprintf("/chat/%s/messages?cursor=%s", params.PartnerID, cursorMessageID),
				AttrHxTrigger:   "intersect once",
				AttrHxSwap:      SwapAfterEnd,
				AttrHxIndicator: "#messages-indicator",
				// For internal requests not to use this indicator
				AttrHxDisinherit: AttrHxIndicator,
			},
				ChatMessage(msg),
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

func ChatMessage(params ChatMessageParams) HyperNode {
	return Div(
		IfElse(params.Status == StatusPending,
			KV{AttrId: fmt.Sprintf("pending-chat-message-%d", params.ClientMessageID)},
			IfElse(!params.FromMe && params.Status != StatusRead,
				KV{
					AttrId:        "chat-message-" + params.ID,
					AttrHxPost:    "/api/v1/chats/" + params.PartnerID + "/messages/mark_as_read?upto_message_id=" + params.ID,
					AttrHxTrigger: "intersect once",
					AttrHxSwap:    SwapNone,
				},
				KV{AttrId: "chat-message-" + params.ID},
			),
		),
		Div(KV{
			AttrClass: "flex w-full " + IfElse(params.FromMe, "justify-end", "justify-start"),
			AttrOnContextMenu: `
				event.preventDefault();
				const menu = document.getElementById('message-context-menu');
				menu.dataset.messageID = "` + params.ID + `";
				menu.dataset.partnerID = "` + params.PartnerID + `";
				const fromMe = ` + IfElse(params.FromMe, "true", "false") + `;
				document.getElementById('ctx-edit').style.display = fromMe ? 'flex' : 'none';
				document.getElementById('ctx-delete').style.display = fromMe ? 'flex' : 'none';

				menu.style.left = event.clientX + 'px';
				menu.style.top = event.clientY + 'px';
				menu.classList.remove('hidden');
				const menuRect = menu.getBoundingClientRect();
				const viewportWidth = window.innerWidth;
				const viewportHeight = window.innerHeight;
				// Adjust if overflows right edge
				if (event.clientX + menuRect.width > viewportWidth) menu.style.left = (event.clientX - menuRect.width) + 'px';
				// Adjust if overflows bottom edge  
				if (event.clientY + menuRect.height > viewportHeight) menu.style.top = (event.clientY - menuRect.height) + 'px';
			`,
		},
			Div(KV{AttrClass: "flex flex-col " + IfElse(params.FromMe, "items-end", "items-start") + " max-w-[70%]"},
				Div(KV{AttrClass: "px-4 py-2 " + IfElse(params.FromMe, "bg-blue text-bg-primary rounded-l-2xl rounded-tr-2xl", "bg-bg-tertiary text-fg-primary rounded-r-2xl rounded-tl-2xl")},
					P(KV{AttrClass: "message-content whitespace-pre-wrap"}, params.Content),
					Div(KV{AttrClass: "flex items-center gap-1 mt-1 justify-end"},
						If(params.Status != StatusPending,
							P(KV{AttrClass: "text-xs " + IfElse(params.FromMe, "text-bg-primary/70", "text-fg-secondary")}, params.SentAt.Format("Jan 2, 3:04 PM")),
						),
						If(params.FromMe,
							IfElse(params.Status == StatusPending,
								PendingMessageIndicator(),
								IfElse(params.Status == StatusSent,
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

func PendingMessageIndicator() HyperNode {
	return Span(RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-bg-primary animate-spin" style="animation-duration: 1s;"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>`))
}

func ReadMessageIndicator() HyperNode {
	return Span(RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-bg-primary"><polyline points="16 6 8 14 4 10"></polyline><polyline points="22 6 14 14 10 10"></polyline></svg>`))
}

func sentMessageIndicator(messageID string) HyperNode {
	return Span(RawText(fmt.Sprintf(`
		<svg id="unread-message-indicator-%s" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-bg-primary"><polyline points="20 6 9 17 4 12"></polyline></svg>`,
		messageID,
	)))
}

type ChatInputFormParams struct {
	PartnerID string

	ForEdit    bool
	MessageID  string
	OldContent string
}

func ChatInputForm(params ChatInputFormParams) HyperNode {
	textarea := Textarea(KV{
		AttrClass:       "w-full bg-bg-tertiary text-fg-primary resize-none outline-none max-h-[120px] min-h-[44px] py-2.5 px-4 leading-6 self-center overflow-y-auto " + IfElse(params.ForEdit, "rounded-b-2xl rounded-t-none", "rounded-2xl"),
		AttrName:        "content",
		AttrPlaceholder: "Write a message...",
		AttrRows:        "1",
		AttrAutofocus:   true,
		AttrOnInput: `
			this.style.height = "auto";
			this.style.height = Math.min(this.scrollHeight, 120)+"px";
		`,
		AttrOnKeyDown: `
			if (event.key === "Enter" && !event.shiftKey) {
				event.preventDefault();
				document.getElementById("send-btn").click();
			}
		`,
	},
		IfElse(params.ForEdit, params.OldContent, ""),
	)

	return Form(KV{AttrId: "message-input-form", AttrClass: "flex items-center gap-2 px-3 py-2"},
		IfElse(params.ForEdit,
			KV{
				AttrHxPut:  fmt.Sprintf("/chat/%s/messages/%s", params.PartnerID, params.MessageID),
				AttrHxSwap: SwapOuterHtml,
				AttrHxOn(EventHtmxConfigRequest): `
					const content = event.detail.parameters.content.trim();
					if (content == "") {
						event.preventDefault();
						return;
					}
					event.detail.parameters.content = content;
				`,
			},
			KV{
				AttrWsSend: true,
				AttrHxOn(EventHtmxWsConfigSend): `
					const content = event.detail.parameters.content.trim();
					if (content == "") {
						event.preventDefault();
						return;
					}
					event.detail.parameters.content = content;
					event.detail.parameters.clientMessageID = window.messageIDGenerator.GetID();
				`,
				AttrHxOn(EventHtmxWsAfterSend): `
					const textarea = event.detail.elt.querySelector('textarea[name="content"]');
					textarea.value = '';
					textarea.style.height = 'auto';
				`,
			},
		),
		If(!params.ForEdit, Empty(
			Input(KV{AttrType: TypeHidden, AttrName: "kind", AttrValue: "SendMessage"}),
			Input(KV{AttrType: TypeHidden, AttrName: "partnerID", AttrValue: params.PartnerID}),
		)),
		IfElse(params.ForEdit,
			Div(KV{AttrClass: "flex flex-col flex-1"},
				Div(KV{AttrClass: "flex items-center justify-between px-3 py-1.5 bg-blue/10 rounded-t-xl rounded-b-none"},
					P(KV{AttrClass: "text-sm text-fg-secondary"}, "Editing message"),
					Button(KV{
						AttrClass:    "px-3 py-1 text-sm bg-bg-tertiary hover:bg-red-500/20 hover:text-red-500 rounded transition-colors cursor-pointer",
						AttrHxGet:    fmt.Sprintf("/chat/%s/message_input_form", params.PartnerID),
						AttrHxTarget: "#message-input-form",
						AttrHxSwap:   SwapOuterHtml,
					},
						"Cancel",
					),
				),
				textarea,
			),
			textarea,
		),
		Button(KV{
			AttrId:    "send-btn",
			AttrType:  "submit",
			AttrClass: "w-12 h-12 rounded-full bg-bg-primary hover:bg-bg-tertiary flex items-center justify-center flex-shrink-0 cursor-pointer transition-colors",
		},
			RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-send-horizontal-icon lucide-send-horizontal text-blue"><path d="M3.714 3.048a.498.498 0 0 0-.683.627l2.843 7.627a2 2 0 0 1 0 1.396l-2.842 7.627a.498.498 0 0 0 .682.627l18-8.5a.5.5 0 0 0 0-.904z"/><path d="M6 12h16"/></svg>`),
		),
	)
}

func messageContextMenu() HyperNode {
	return Div(KV{
		AttrId:    "message-context-menu",
		AttrClass: "hidden fixed z-50 bg-bg-secondary border border-bg-tertiary rounded-lg shadow-lg py-1 min-w-[120px]",
	},
		Button(KV{
			AttrClass: "w-full px-4 py-2 text-left text-sm text-fg-primary hover:bg-bg-tertiary flex items-center gap-2",
			AttrOnClick: `
				const messageID = document.getElementById("message-context-menu").dataset.messageID;
				const content = document.getElementById("chat-message-"+messageID).querySelector('.message-content').textContent;
				navigator.clipboard.writeText(content);
			`,
		},
			RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2"/></svg>`),
			"Copy",
		),
		Button(KV{
			AttrId:    "ctx-edit",
			AttrClass: "w-full px-4 py-2 text-left text-sm text-fg-primary hover:bg-bg-tertiary flex items-center gap-2",
			AttrOnClick: `
				const messageID = document.getElementById('message-context-menu').dataset.messageID;
				const partnerID = document.getElementById('message-context-menu').dataset.partnerID;
				const content = document.getElementById("chat-message-"+messageID).querySelector('.message-content').textContent;
				htmx.ajax('GET', '/chat/'+partnerID+'/edit_message_input_form/'+messageID, {
					target: "#message-input-form",
					swap:   "outerHTML",
					values: {content},
				});
			`,
		},
			RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.174 6.812a1 1 0 0 0-3.986-3.987L3.842 16.174a2 2 0 0 0-.5.83l-1.321 4.352a.5.5 0 0 0 .623.622l4.353-1.32a2 2 0 0 0 .83-.497z"/></svg>`),
			"Edit",
		),
		Button(KV{
			AttrId:    "ctx-delete",
			AttrClass: "w-full px-4 py-2 text-left text-sm text-fg-primary hover:bg-bg-tertiary flex items-center gap-2",
			AttrOnClick: `
				const messageID = document.getElementById('message-context-menu').dataset.messageID;
				htmx.ajax('DELETE', '/api/v1/chats/messages/'+messageID, { swap: 'none' });
			`,
		},
			RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/></svg>`),
			"Delete",
		),
	)
}

func partnerContextMenu() HyperNode {
	return Div(KV{
		AttrId:    "partner-context-menu",
		AttrClass: "hidden fixed z-50 bg-bg-secondary border border-bg-tertiary rounded-lg shadow-lg py-1 min-w-[120px]",
	},
		Button(KV{
			AttrClass: "w-full px-4 py-2 text-left text-sm text-fg-primary hover:bg-bg-tertiary flex items-center gap-2",
			AttrOnClick: `
				const partnerID = document.getElementById('partner-context-menu').dataset.partnerID;
				htmx.ajax('DELETE', '/api/v1/chats/'+partnerID, { swap: 'none' });
			`,
		},
			RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/></svg>`),
			"Delete",
		),
	)
}
