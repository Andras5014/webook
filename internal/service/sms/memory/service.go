package memory

import (
	"context"
	"fmt"
	"github.com/Andras5014/gohub/internal/service/sms"
)

type Service struct{}

func NewService() sms.Service {
	return &Service{}
}
func (s *Service) Send(ctx context.Context, tplToken string, args []sms.NamedArg, numbers ...string) error {
	fmt.Println("发送短信到xx")
	return nil
}
