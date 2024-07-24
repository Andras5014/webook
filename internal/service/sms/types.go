package sms

import "context"

type Service interface {
	Send(ctx context.Context, signature string, args []string, numbers ...string) error
}
