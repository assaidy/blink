package components

import (
	"github.com/assaidy/hyper"
)

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

func ChatInputForm() h.Node {
	return h.Form(h.KV{
		"ws-send": true,
		"hx-on::ws-config-send": `
			const content = event.detail.parameters.content.trim();
			if (content == "") event.preventDefault();
			event.detail.parameters.content = content;
		`,
		"class": "flex items-center gap-2 px-3 py-2",
	},
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
