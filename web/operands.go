package main

import (
	"strconv"

	"github.com/goadesign/goa"
	"github.com/goby/recipes/web/app"
)

// OperandsController implements the operands resource.
type OperandsController struct {
	*goa.Controller
}

// NewOperandsController creates a operands controller.
func NewOperandsController(service *goa.Service) *OperandsController {
	return &OperandsController{Controller: service.NewController("OperandsController")}
}

// Add runs the add action.
func (c *OperandsController) Add(ctx *app.AddOperandsContext) error {
	// OperandsController_Add: start_implement

	// Put your logic here
	sum := ctx.Left + ctx.Right

	// OperandsController_Add: end_implement
	return ctx.OK([]byte(strconv.Itoa(sum)))
}

// Add runs the add action.
func (c *OperandsController) Minus(ctx *app.MinusOperandsContext) error {
	// OperandsController_Add: start_implement

	// Put your logic here
	res := ctx.Left - ctx.Right

	// OperandsController_Add: end_implement
	return ctx.OK([]byte(strconv.Itoa(res)))
}
