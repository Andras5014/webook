package sms

import "context"

type Service interface {
	Send(ctx context.Context, tplToken string, args []NamedArg, numbers ...string) error
}
type NamedArg struct {
	Name  string
	Value string
}
