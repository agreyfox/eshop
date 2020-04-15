package content

import (
	"github.com/agreyfox/eshop/system/logs"
	"go.uber.org/zap"
)

var (
	err    error
	logger *zap.SugaredLogger = logs.Log.Sugar()
)
