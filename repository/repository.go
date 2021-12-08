package repository

import (
	"dang/model"
)

type Repository interface {
	InsertDecision(model.Decision) error
}