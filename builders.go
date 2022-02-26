package bcr

import "github.com/diamondburned/arikawa/v3/discord"

type CommandBuilder struct {
	r     *Router
	path  string
	check Check[*CommandContext]
}

// Check adds a check to this command handler.
func (c *CommandBuilder) Check(check Check[*CommandContext]) *CommandBuilder {
	c.check = check
	return c
}

// Exec adds this command to the router.
func (c *CommandBuilder) Exec(hn HandlerFunc[*CommandContext]) {
	c.r.commands[c.path] = &handler[*CommandContext]{
		check:   c.check,
		handler: hn,
	}
}

type ModalBuilder struct {
	r     *Router
	id    discord.ComponentID
	check Check[*ModalContext]
}

// Check adds a check to this modal handler.
func (m *ModalBuilder) Check(check Check[*ModalContext]) *ModalBuilder {
	m.check = check
	return m
}

// Exec adds this modal to the router.
func (m *ModalBuilder) Exec(hn HandlerFunc[*ModalContext]) {
	m.r.modals[m.id] = &handler[*ModalContext]{
		check:   m.check,
		handler: hn,
	}
}

type ButtonBuilder struct {
	r     *Router
	id    discord.ComponentID
	check Check[*ButtonContext]
	once  bool
	msgID discord.MessageID
}

// Once changes this button interaction to only be listened for once.
// If the check fails, it will be silent, and the button will not be removed.
func (b *ButtonBuilder) Once() *ButtonBuilder {
	b.once = true
	return b
}

// Message changes this button interaction to be limited to a single message ID.
func (b *ButtonBuilder) Message(id discord.MessageID) *ButtonBuilder {
	b.msgID = id
	return b
}

// Check adds a check to this button handler.
// The behaviour of the passed check is controlled by b.Once().
func (b *ButtonBuilder) Check(check Check[*ButtonContext]) *ButtonBuilder {
	b.check = check
	return b
}

// Exec adds this button to the router.
func (b *ButtonBuilder) Exec(hn HandlerFunc[*ButtonContext]) {
	b.r.componentsMu.Lock()
	defer b.r.componentsMu.Unlock()

	b.r.buttons[componentKey{b.id, b.msgID}] = &handler[*ButtonContext]{
		check:   b.check,
		handler: hn,
		once:    b.once,
	}
}
