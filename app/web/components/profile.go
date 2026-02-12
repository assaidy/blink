package components

import (
	"fmt"
	"slices"
	"time"

	"github.com/assaidy/h"
)

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
					h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-green-500"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>`),
					h.P(h.KV{"class": "text-green-600"}, "Email verified"),
				),
				h.Empty(
					h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-yellow-500"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>`),
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
						h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>`),
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
						h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>`),
						"Remove", spinner(spinnerID),
					),
				),
			)
		}),
	)
}
