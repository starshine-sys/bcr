package bcr

import (
	"emperror.dev/errors"
)

const (
	ErrNotCommand = errors.Sentinel("not a command interaction")
	ErrNotModal   = errors.Sentinel("not a modal interaction")
)

const ErrUnknownCommand = errors.Sentinel("no command with that path found")

type HasContext interface {
	*CommandContext | *ModalContext
}

type HandlerFunc[T HasContext] func(T) error

type handler[T HasContext] struct {
	check   Check[T]
	handler HandlerFunc[T]
}
