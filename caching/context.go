package caching

import "context"

func checkedCacheContext(ctx context.Context) (context.Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return ctx, err
	}
	return ctx, nil
}
