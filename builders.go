package bcr

import "github.com/diamondburned/arikawa/v3/discord"

type commandBuilder struct {
	r     *Router
	path  string
	check Check[*CommandContext]
}

func (s *commandBuilder) Check(c Check[*CommandContext]) *commandBuilder {
	s.check = c
	return s
}

func (s *commandBuilder) Exec(hn HandlerFunc[*CommandContext]) {
	s.r.commands[s.path] = &handler[*CommandContext]{
		check:   s.check,
		handler: hn,
	}
}

type modalBuilder struct {
	r     *Router
	id    discord.ComponentID
	check Check[*ModalContext]
}

func (s *modalBuilder) Check(c Check[*ModalContext]) *modalBuilder {
	s.check = c
	return s
}

func (s *modalBuilder) Exec(hn HandlerFunc[*ModalContext]) {
	s.r.modals[s.id] = &handler[*ModalContext]{
		check:   s.check,
		handler: hn,
	}
}
