package ctxutil

import (
	"context"
	"strconv"

	"github.com/crazyfrankie/goim/pkg/errorx"
	"github.com/crazyfrankie/goim/types/errno"
)

func CheckAccess(ctx context.Context, ownerUserID int64) error {
	if MustGetUserIDFromCtx(ctx) == ownerUserID {
		return nil
	}

	return errorx.New(errno.ErrNoPermissionCode, errorx.KV("ownerID", strconv.FormatInt(ownerUserID, 10)))
}
