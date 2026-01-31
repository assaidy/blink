package components

import (
	"github.com/assaidy/gg"
)

func ChatPage() gg.Node {
	return rootLayout(
		gg.Div(gg.KV{"class": "h-screen flex justify-center items-center"},
			gg.H1(gg.KV{"class": "text-5xl font-bold"}, "Chat Page"),
		),
	)
}
