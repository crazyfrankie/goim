package tx

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/crazyfrankie/goim/infra/contract/tx"
	"github.com/crazyfrankie/goim/pkg/errorx"
)

type mongoTx struct {
	client *mongo.Client
	tx     func(context.Context, func(ctx context.Context) error) error
}

func NewMongoTx(ctx context.Context, client *mongo.Client) (tx.Tx, error) {
	mtx := mongoTx{
		client: client,
	}
	if err := mtx.init(ctx); err != nil {
		return nil, err
	}
	return &mtx, nil
}

func (m *mongoTx) init(ctx context.Context) error {
	var res map[string]any
	if err := m.client.Database("admin").RunCommand(ctx, bson.M{"isMaster": 1}).Decode(&res); err != nil {
		return errorx.Wrapf(err, "check whether mongo is deployed in a cluster")
	}
	if _, allowTx := res["setName"]; !allowTx {
		return nil // non-clustered transactions are not supported
	}
	m.tx = func(fnctx context.Context, fn func(ctx context.Context) error) error {
		sess, err := m.client.StartSession()
		if err != nil {
			return errorx.Wrapf(err, "mongodb start session failed")
		}
		defer sess.EndSession(fnctx)
		_, err = sess.WithTransaction(fnctx, func(sessCtx mongo.SessionContext) (any, error) {
			return nil, fn(sessCtx)
		})
		return errorx.Wrapf(err, "mongodb transaction failed")
	}
	return nil
}

func (m *mongoTx) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.tx == nil {
		return fn(ctx)
	}
	return m.tx(ctx, fn)
}
