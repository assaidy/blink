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
					gg.Form(gg.KV{"class": "space-y-5"},
						gg.Div(gg.KV{"class": "space-y-1"},
							gg.Label(gg.KV{"for": "name", "class": "block text-sm font-medium text-fg-secondary"}, "Full Name"),
							gg.Input(gg.KV{"type": "text", "id": "name", "name": "name", "required": true, "placeholder": "Enter your full name", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
						),
						gg.Div(gg.KV{"class": "space-y-1"},
							gg.Label(gg.KV{"for": "username", "class": "block text-sm font-medium text-fg-secondary"}, "Username"),
							gg.Input(gg.KV{"type": "text", "id": "username", "name": "username", "required": true, "placeholder": "Choose a username", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
						),
						gg.Div(gg.KV{"class": "space-y-1"},
							gg.Label(gg.KV{"for": "email", "class": "block text-sm font-medium text-fg-secondary"}, "Email Address"),
							gg.Input(gg.KV{"type": "email", "id": "email", "name": "email", "required": true, "placeholder": "you@example.com", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors"}),
						),
						gg.Div(gg.KV{"class": "space-y-1"},
							gg.Label(gg.KV{"class": "block text-sm font-medium text-fg-secondary"}, "Bio"),
							gg.Textarea(gg.KV{"placeholder": "Tell us about yourself...", "class": "bg-bg-tertiary border-2 border-bg-tertiary focus:border-blue rounded-lg text-fg-primary w-full px-4 py-3 outline-none transition-colors min-h-24 resize-none"}),
						),
						gg.Button(gg.KV{"class": "w-full rounded-lg bg-blue hover:bg-blue/80 text-bg-primary font-semibold py-3 px-4 cursor-pointer transition-colors mt-2"},
							"Create Account",
						),
					),
					gg.P(gg.KV{"class": "text-center text-fg-secondary text-sm mt-6"},
						"Already have an account? ",
						gg.A(gg.KV{"href": "/login", "class": "text-blue hover:underline font-medium"}, "Sign in"),
					),
				),
			),
		),
	)
}

func LoginPage() gg.Node {
	return rootLayout(
		gg.Div(gg.KV{"class": "h-screen flex justify-center items-center"},
			gg.H1(gg.KV{"class": "text-5xl font-bold"}, "Login Page"),
		),
	)
}
