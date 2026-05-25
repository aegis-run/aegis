package core

import "context"

type DB interface {
	Engine() string
	Close() error
	IsReady(context.Context) (bool, error)
}
