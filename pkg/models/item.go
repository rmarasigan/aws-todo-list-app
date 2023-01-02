package models

import (
	"context"

	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
	"github.com/rmarasigan/aws-todo-list-app/pkg/service"
)

// ItemCount returns the number of items in your table.
func ItemCount(ctx context.Context, tablename string) int {
	result, err := service.DynamoScan(ctx, tablename)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "ItemCount", Message: "Failed to scan the table"})
		return 0
	}

	return int(result.Count)
}
