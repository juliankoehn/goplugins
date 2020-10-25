// this is a example plugin
package main

import (
	"fmt"
	"goplugins/core/routing"
	"net/http"
)

var Plugin UserPlugin

type UserPlugin string

func (p *UserPlugin) Install() {
	// install database tables blabla
	fmt.Println("Install Hook")
}
func (p *UserPlugin) PostInstall() {
	fmt.Println("PostInstall Hook")
}
func (p *UserPlugin) Update()     {}
func (p *UserPlugin) PostUpdate() {}
func (p *UserPlugin) Activate()   {}
func (p *UserPlugin) Deactivate() {
	// DROP database tables and so on...
}

// ConfigureRoutes is where you can install routes to our mux
func (p *UserPlugin) ConfigureRoutes(r *routing.Mux) {
	r.GET("/product", func(c routing.Context) error {
		return c.String(http.StatusOK, "Hello from ProductPlugin")
	})
}
