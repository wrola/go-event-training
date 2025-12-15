package service

import (
	"context"

	_ "github.com/lib/pq"

	"remove_sagas/common"
	"remove_sagas/orders"
)

func Run(ctx context.Context) {
	common.StartService(
		ctx,
		[]common.AddHandlersFn{
			orders.Initialize,
		},
	)
}
