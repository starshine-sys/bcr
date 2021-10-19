package bcr

// CommandFunc is a command function.
type CommandFunc func(*CommandContext) error

// CommandMiddleware is a command middleware.
type CommandMiddleware func(next CommandFunc) CommandFunc

// Command is an application command.
type Command struct {
	Name        string
	Description string

	Middlewares []CommandMiddleware

	Execute CommandFunc
}
