package global

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var (
	// zap variable
	GlobalZapLog *zap.Logger
	// validator variable
	GlobalValidator *validator.Validate
)