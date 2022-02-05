package bcr

type command struct {
	Check Check[*CommandContext]

	Command func(*CommandContext) error
}

type commandBuilder struct {
	r     *Router
	path  string
	check Check[*CommandContext]
}

func (r *Router) Command(path string) *commandBuilder {
	return &commandBuilder{
		r:    r,
		path: path,
	}
}

func (s *commandBuilder) Check(c Check[*CommandContext]) *commandBuilder {
	s.check = c
	return s
}

func (s *commandBuilder) Exec(cmd func(*CommandContext) error) {
	s.r.commands[s.path] = &command{
		Check:   s.check,
		Command: cmd,
	}
}
