package safego

import (
	"context"
)

func Go(ctx context.Context, fn func()) {
	go func() {
		defer recovery(ctx)

		fn()
	}()
}
