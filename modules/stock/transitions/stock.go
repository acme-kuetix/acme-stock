package transitions

import (
	"github.com/kuetix/engine/engine/domain/interfaces"
	"github.com/kuetix/engine/engine/workflow"
)

type stockTransitions struct {
	workflow.BaseServiceTransition
}

func NewStockTransitions() interfaces.ServiceTransitions {
	return &stockTransitions{}
}

const (
	moveStateDraft      = "draft"
	locationsCollection = "stock_locations"
	movesCollection     = "stock_moves"
)
