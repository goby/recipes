//go:generate goagen bootstrap -d github.com/goby/recipes/web/design

package main

import (
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	"github.com/goby/recipes/web/app"
)

func main() {
	// Create service
	service := goa.New("adder")

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	// Mount "operands" controller
	c := NewOperandsController(service)
	app.MountOperandsController(service, c)
	// Mount "swagger" controller
	c2 := NewSwaggerController(service)
	app.MountSwaggerController(service, c2)

	// Start service
	if err := service.ListenAndServe(":9999"); err != nil {
		service.LogError("startup", "err", err)
	}
}
