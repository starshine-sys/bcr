package bcr

import "github.com/diamondburned/arikawa/v3/discord"

// Check is a check for slash commands.
// If err != nil:
// - if err implements CheckError, respond with the content and embeds from that method
// - otherwise, print the string representation of the error
type Check[T HasContext] func(ctx T) (err error)

// And combines all the given checks into a single check.
// The first one to fail is returned.
func And[T HasContext](checks ...Check[T]) Check[T] {
	return func(ctx T) error {
		for _, check := range checks {
			if err := check(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// Or checks all given checks and returns nil if at least one of them succeeds.
func Or[T HasContext](checks ...Check[T]) Check[T] {
	return func(ctx T) (err error) {
		for _, check := range checks {
			err = check(ctx)
			if err == nil {
				return nil
			}
		}
		return err
	}
}

type CheckError[T HasContext] interface {
	CheckError(T) (string, []discord.Embed)
}
