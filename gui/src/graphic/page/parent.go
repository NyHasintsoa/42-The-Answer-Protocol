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

func (p *BasePage) GetState() pagemodel.PageState {
	return p.State
}

func (p *BasePage) GetNextState() pagemodel.PageState {
	return p.NextState
}

func (p *BasePage) SetNextState(state pagemodel.PageState) {
	p.NextState = state
}

func (p *BasePage) GetContext() *model.GameContext {
	return p.Context
}