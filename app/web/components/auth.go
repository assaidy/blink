package components

import (
	"github.com/assaidy/gg"
)

func RegisterPage() gg.Node {
	return rootLayout(
		gg.Div(gg.KV{"class": "min-h-screen flex justify-center items-center bg-bg-primary sm:px-6 lg:px-8"},
			gg.Div(gg.KV{"class": "w-full min-h-screen sm:min-h-0 sm:max-w-md sm:p-8 bg-bg-secondary sm:rounded-lg sm:shadow-lg flex flex-col justify-center"},
				gg.Div(gg.KV{"class": "p-6 sm:p-0"},
					gg.H2(gg.KV{"class": "text-fg-primary text-2xl font-bold text-center mb-8"}, "Create Account"),
					RegisterForm(),
					gg.P(gg.KV{"class": "text-center text-fg-secondary text-sm mt-6"},
						"Already have an account? ", gg.A(gg.KV{"href": "/login", "class": "text-blue hover:underline font-medium"}, "Sign in"),
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

func RegisterForm(params ...RegisterFormParams) gg.Node {
	var p RegisterFormParams
	if len(params) != 0 {
		p = params[0]
	}

	return gg.Form(gg.KV{"hx-post": "/register", "hx-swap": "outerHTML", "class": "space-y-5"},
		gg.Div(gg.KV{"class": "space-y-1"},
			gg.Label(gg.KV{"for": "name", "class": "block text-sm font-medium text-fg-secondary"}, "Full Name"),
			gg.Input(gg.KV{"type": "text", "id": "name", "name": "name", "required": true, "value": p.Name, "placeholder": "Enter your full name", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			gg.If(p.NameErr != nil, gg.P(gg.KV{"class": "text-red-500 text-sm mt-1"}, p.NameErr)),
		),
		gg.Div(gg.KV{"class": "space-y-1"},
			gg.Label(gg.KV{"for": "username", "class": "block text-sm font-medium text-fg-secondary"}, "Username"),
			gg.Input(gg.KV{"type": "text", "id": "username", "name": "username", "value": p.Username, "placeholder": "Choose a username", "required": true, "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			gg.If(p.UsernameErr != nil, gg.P(gg.KV{"class": "text-red-500 text-sm mt-1"}, p.UsernameErr)),
		),
		gg.Div(gg.KV{"class": "space-y-1"},
			gg.Label(gg.KV{"for": "email", "class": "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			gg.Input(gg.KV{"type": "email", "id": "email", "name": "email", "value": p.Email, "required": true, "placeholder": "you@example.com", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			gg.If(p.EmailErr != nil, gg.P(gg.KV{"class": "text-red-500 text-sm mt-1"}, p.EmailErr)),
		),
		gg.Div(gg.KV{"class": "space-y-1"},
			gg.Label(gg.KV{"for": "bio", "class": "block text-sm font-medium text-fg-secondary"}, "Bio"),
			gg.Textarea(gg.KV{"id": "bio", "name": "bio", "value": p.Bio, "placeholder": "Tell us about yourself...", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors min-h-24 resize-none"}),
			gg.If(p.BioErr != nil, gg.P(gg.KV{"class": "text-red-500 text-sm mt-1"}, p.BioErr)),
		),
		gg.Button(gg.KV{"class": "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
			"Create Account",
		),
	)
}

func LoginPage() gg.Node {
	return rootLayout(
		gg.Div(gg.KV{"class": "min-h-screen flex justify-center items-center bg-bg-primary sm:px-6 lg:px-8"},
			gg.Div(gg.KV{"class": "w-full min-h-screen sm:min-h-0 sm:max-w-md sm:p-8 bg-bg-secondary sm:rounded-lg sm:shadow-lg flex flex-col justify-center"},
				gg.Div(gg.KV{"class": "p-6 sm:p-0"},
					gg.H2(gg.KV{"class": "text-fg-primary text-2xl font-bold text-center mb-8"}, "Sign In"),
					LoginForm(),
					gg.P(gg.KV{"class": "text-center text-fg-secondary text-sm mt-6"},
						"Don't have an account? ", gg.A(gg.KV{"href": "/register", "class": "text-blue hover:underline font-medium"}, "Create one"),
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

func LoginForm(params ...LoginFormParams) gg.Node {
	var p LoginFormParams
	if len(params) != 0 {
		p = params[0]
	}

	return gg.Form(gg.KV{"hx-post": "/login", "hx-swap": "outerHTML", "class": "space-y-5"},
		gg.Div(gg.KV{"class": "space-y-1"},
			gg.Label(gg.KV{"for": "email", "class": "block text-sm font-medium text-fg-secondary"}, "Email Address"),
			gg.Input(gg.KV{"type": "email", "id": "email", "name": "email", "value": p.Email, "required": true, "placeholder": "you@example.com", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
			gg.If(p.EmailErr != nil, gg.P(gg.KV{"class": "text-red-500 text-sm mt-1"}, p.EmailErr)),
		),
		gg.Button(gg.KV{"class": "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
			"Sign In",
		),
	)
}

type OtpFormParams struct {
	OtpID  string
	Otp    string
	OtpErr any
}

func OtpForm(params OtpFormParams) gg.Node {
	return gg.Form(gg.KV{"hx-post": "/verify_otp", "hx-swap": "outerHTML", "class": "space-y-5"},
		gg.Input(gg.KV{"type": "hidden", "name": "otpID", "value": params.OtpID}),
		gg.Div(gg.KV{"class": "space-y-1"},
			gg.Label(gg.KV{"for": "otp", "class": "block text-sm font-medium text-fg-secondary"}, "Verification Code"),
			gg.P(gg.KV{"class": "text-fg-secondary text-sm mb-2"}, "We've sent a 6-digit code to your email address. Please enter it below to verify your identity."),
			gg.Input(gg.KV{
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
			gg.If(params.OtpErr != nil, gg.P(gg.KV{"class": "text-red-500 text-sm mt-1"}, params.OtpErr)),
		),
		gg.Button(gg.KV{"class": "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
			"Verify",
		),
	)
}
