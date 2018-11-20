package utils

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	lg, _ := zap.NewDevelopment()
	Logger = lg
}
