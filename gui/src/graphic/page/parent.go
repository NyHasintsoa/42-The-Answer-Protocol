// Package page implements the lifecycle and screens of the graphical client.
package page

import (
	"tap-gui/src/model"
	pagemodel "tap-gui/src/model/page"
)

type Page interface {
	Init(ctx *model.GameContext)
	Unload()
	Render()
	GetState() pagemodel.PageState
	GetNextState() pagemodel.PageState
	SetNextState(state pagemodel.PageState)
	GetContext() *model.GameContext
}

type BasePage struct {
	Context    *model.GameContext
	State      pagemodel.PageState
	NextState  pagemodel.PageState
	IsUnloaded bool
}

// Init stores ctx and resets the pending page transition to the page's current state.
func (p *BasePage) Init(ctx *model.GameContext) {
	p.NextState = p.State
	p.Context = ctx
}

// Unload marks the page as unloaded; repeated calls have no effect.
func (p *BasePage) Unload() {
	if p.IsUnloaded {
		return
	}
	p.IsUnloaded = true
}

// GetState returns the page's current state.
func (p *BasePage) GetState() pagemodel.PageState {
	return p.State
}

// GetNextState returns the state requested for the next page transition.
func (p *BasePage) GetNextState() pagemodel.PageState {
	return p.NextState
}

// SetNextState requests a transition to state.
func (p *BasePage) SetNextState(state pagemodel.PageState) {
	p.NextState = state
}

// GetContext returns the game context assigned to the page.
func (p *BasePage) GetContext() *model.GameContext {
	return p.Context
}
