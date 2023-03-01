// Package that holds a refernce to a logger object (singleton like architecture)
package log

import (
	"go.uber.org/zap"
)

var (
	Logger *zap.Logger
)
