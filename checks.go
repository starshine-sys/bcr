package bcr

type Context interface {
	*CommandContext
}

// Check is a check for slash commands.
// If err != nil:
// - if err implements Checker, respond with the content and embeds from that method
// - otherwise, print the string representation of the error
type Check[T Context] func(ctx T) (err error)

// And combines all the given checks into a single check.
// The first one to fail is returned.
func And[T Context](checks ...Check[T]) Check[T] {
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
func Or[T Context](checks ...Check[T]) Check[T] {
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
