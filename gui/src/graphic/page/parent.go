package page

import (
	"tap-gui/src/model"
	"tap-gui/src/model/enums"
)

type Page interface {
	Init(ctx *model.GameContext)
	Unload()
	Render()
	GetState() enums.PageState
	GetNextState() enums.PageState
	SetNextState(state enums.PageState)
	GetContext() *model.GameContext
}

type BasePage struct {
	Context    *model.GameContext
	State      enums.PageState
	NextState  enums.PageState
	IsUnloaded bool
}

func (p *BasePage) Init(ctx *model.GameContext) {
	p.NextState = p.State
	p.Context = ctx
}

func (p *BasePage) Unload() {
	if p.IsUnloaded {
		return
	}
	p.IsUnloaded = true
}

func (p *BasePage) GetState() enums.PageState {
	return p.State
}

func (p *BasePage) GetNextState() enums.PageState {
	return p.NextState
}

func (p *BasePage) SetNextState(state enums.PageState) {
	p.NextState = state
}

func (p *BasePage) GetContext() *model.GameContext {
	return p.Context
}