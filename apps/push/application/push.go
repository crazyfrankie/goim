package application

import (
	"context"

	pushv1 "github.com/crazyfrankie/goim/protocol/push/v1"
)

type PushApplicationService struct {
	pushv1.UnimplementedPushServiceServer
}

func NewPushApplicationService() *PushApplicationService {
	return &PushApplicationService{}
}

func (p *PushApplicationService) DelUserPushToken(ctx context.Context, request *pushv1.DelUserPushTokenRequest) (*pushv1.DelUserPushTokenResponse, error) {
	//TODO implement me
	panic("implement me")
}
