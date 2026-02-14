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
