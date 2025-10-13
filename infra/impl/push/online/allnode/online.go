package allnode

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	"github.com/crazyfrankie/goim/pkg/lang/conv"
	"github.com/crazyfrankie/goim/pkg/lang/slice"
	"github.com/crazyfrankie/goim/pkg/logs"
	msgv1 "github.com/crazyfrankie/goim/protocol/msg/v1"
	msggatewayv1 "github.com/crazyfrankie/goim/protocol/msggateway/v1"
	"github.com/crazyfrankie/goim/types/consts"
)

type DefaultAllNode struct {
	disCov discovery.Conn

	maxConcurrentWorkers int64
}

func NewDefaultAllNode(disCov discovery.Conn, maxConcurrentWorkers int64) *DefaultAllNode {
	return &DefaultAllNode{disCov: disCov, maxConcurrentWorkers: maxConcurrentWorkers}
}

func (d *DefaultAllNode) GetConnsAndOnlinePush(ctx context.Context, msg *msgv1.Message,
	pushToUserIDs []string) ([]*msggatewayv1.SingleMsgToUserResults, error) {
	var res []*msggatewayv1.SingleMsgToUserResults

	conns, err := d.disCov.GetConns(ctx, consts.MsgGatewayName)
	if len(conns) == 0 {
		logs.CtxWarnf(ctx, "get gateway conn 0 ")
	} else {
		logs.CtxDebugf(ctx, "get gateway conn, conn length: %d", len(conns))
	}

	if err != nil {
		return nil, err
	}

	var (
		mu         sync.Mutex
		wg         = errgroup.Group{}
		input      = &msggatewayv1.OnlineBatchPushOneMsgRequest{Message: msg, PushToUserIDs: pushToUserIDs}
		maxWorkers = d.maxConcurrentWorkers
	)

	if maxWorkers < 3 {
		maxWorkers = 3
	}

	wg.SetLimit(int(maxWorkers))

	// Online push message
	for _, conn := range conns {
		conn := conn // loop var safe
		ctx := ctx
		wg.Go(func() error {
			msgClient := msggatewayv1.NewMsgGatewayClient(conn)
			reply, err := msgClient.SuperGroupOnlineBatchPushOneMsg(ctx, input)
			if err != nil {
				logs.CtxErrorf(ctx, "SuperGroupOnlineBatchPushOneMsg, err :%v, req : %s", err, input.String())
				return nil
			}

			logs.CtxDebugf(ctx, "push result, reply: %v", reply)
			if reply != nil && reply.SinglePushResult != nil {
				mu.Lock()
				res = append(res, reply.SinglePushResult...)
				mu.Unlock()
			}

			return nil
		})
	}

	_ = wg.Wait()

	// always return nil
	return res, nil
}

func (d *DefaultAllNode) GetOnlinePushFailedUserIDs(ctx context.Context, msg *msgv1.Message,
	wsResults []*msggatewayv1.SingleMsgToUserResults, pushToUserIDs *[]string) []string {

	sendID := conv.Int64ToStr(msg.SendID)

	onlineSuccessUserIDs := []string{sendID}
	for _, v := range wsResults {
		//message sender do not need offline push
		if sendID == v.UserID {
			continue
		}
		// mobile online push success
		if v.OnlinePush {
			onlineSuccessUserIDs = append(onlineSuccessUserIDs, v.UserID)
		}

	}

	return slice.SubSlice(*pushToUserIDs, onlineSuccessUserIDs)
}
