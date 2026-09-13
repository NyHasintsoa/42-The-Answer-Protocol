package component

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type HPBar struct {
	Position  rl.Vector2
	Width     float32
	Height    float32
	MaxHP     float32
	CurrentHP float32
	TargetHP  float32
	Speed     float32 
}

func NewHPBar(x, y, width, height float32) *HPBar {
	return &HPBar{
		Position:  rl.Vector2{X: x, Y: y},
		Width:     width,
		Height:    height,
		MaxHP:     100.0,
		CurrentHP: 100.0,
		TargetHP:  100.0,
		Speed:     80.0, 
	}
}


func (hp *HPBar) SetHP(val float32) {
	if val < 0 {
		val = 0
	}
	if val > hp.MaxHP {
		val = hp.MaxHP
	}
	hp.TargetHP = val
}


func (hp *HPBar) SetMaxHP(maxHP float32) {
	if maxHP <= 0 {
		maxHP = 1
	}
	hp.MaxHP = maxHP
	if hp.CurrentHP > hp.MaxHP {
		hp.CurrentHP = hp.MaxHP
	}
	if hp.TargetHP > hp.MaxHP {
		hp.TargetHP = hp.MaxHP
	}
}

func (hp *HPBar) Update(dt float32) {
	
	if hp.CurrentHP != hp.TargetHP {
		diff := hp.TargetHP - hp.CurrentHP
		step := hp.Speed * dt
		if float32(math.Abs(float64(diff))) <= step {
			hp.CurrentHP = hp.TargetHP
		} else if diff > 0 {
			hp.CurrentHP += step
		} else {
			hp.CurrentHP -= step
		}
	}
}

func (hp *HPBar) Render() {
	x := hp.Position.X
	y := hp.Position.Y
	w := hp.Width
	h := hp.Height

	pct := hp.CurrentHP / hp.MaxHP
	if pct < 0 {
		pct = 0
	} else if pct > 1 {
		pct = 1
	}

	arrowOffset := h * 0.6 

	
	outerBorderColor := rl.NewColor(0, 102, 204, 255)   
	cyanHighlightColor := rl.NewColor(0, 180, 255, 255) 
	emptyBgColor := rl.NewColor(0, 80, 160, 255)        
	emptyBgDark := rl.NewColor(0, 50, 110, 255)

	hpRedLight := rl.NewColor(255, 60, 90, 255) 
	hpRedDark := rl.NewColor(220, 20, 50, 255)  

	
	borderThickness := float32(3.0)

	outerPoly := []rl.Vector2{
		{X: x, Y: y},
		{X: x + w - arrowOffset, Y: y},
		{X: x + w, Y: y + h/2},
		{X: x + w - arrowOffset, Y: y + h},
		{X: x, Y: y + h},
	}

	
	for i := 0; i < len(outerPoly); i++ {
		next := (i + 1) % len(outerPoly)
		rl.DrawLineEx(outerPoly[i], outerPoly[next], borderThickness+2, cyanHighlightColor)
	}

	
	innerPoly := []rl.Vector2{
		{X: x + borderThickness, Y: y + borderThickness},
		{X: x + w - arrowOffset - borderThickness/2, Y: y + borderThickness},
		{X: x + w - borderThickness*1.5, Y: y + h/2},
		{X: x + w - arrowOffset - borderThickness/2, Y: y + h - borderThickness},
		{X: x + borderThickness, Y: y + h - borderThickness},
	}

	
	rl.DrawTriangle(innerPoly[0], innerPoly[4], innerPoly[1], outerBorderColor)
	rl.DrawTriangle(innerPoly[1], innerPoly[4], innerPoly[3], outerBorderColor)
	rl.DrawTriangle(innerPoly[1], innerPoly[3], innerPoly[2], outerBorderColor)

	
	pad := float32(4.0)
	barX := x + pad
	barY := y + pad
	barW := w - pad*2 - arrowOffset
	barH := h - pad*2

	
	rl.DrawRectangleRec(rl.NewRectangle(barX, barY, barW, barH/2), emptyBgColor)
	rl.DrawRectangleRec(rl.NewRectangle(barX, barY+barH/2, barW, barH/2), emptyBgDark)

	
	tipPolyTop := []rl.Vector2{
		{X: barX + barW, Y: barY},
		{X: barX + barW + arrowOffset - pad, Y: barY + barH/2},
		{X: barX + barW, Y: barY + barH/2},
	}
	tipPolyBot := []rl.Vector2{
		{X: barX + barW, Y: barY + barH/2},
		{X: barX + barW + arrowOffset - pad, Y: barY + barH/2},
		{X: barX + barW, Y: barY + barH},
	}
	rl.DrawTriangle(tipPolyTop[0], tipPolyTop[1], tipPolyTop[2], emptyBgColor)
	rl.DrawTriangle(tipPolyBot[0], tipPolyBot[1], tipPolyBot[2], emptyBgDark)

	
	if pct > 0 {
		fillW := barW * pct
		if fillW > barW {
			fillW = barW
		}

		
		rl.DrawRectangleRec(rl.NewRectangle(barX, barY, fillW, barH/2), hpRedLight)
		
		rl.DrawRectangleRec(rl.NewRectangle(barX, barY+barH/2, fillW, barH/2), hpRedDark)

		
		rl.DrawRectangleRec(rl.NewRectangle(barX, barY, fillW, 2), rl.NewColor(255, 180, 200, 220))
	}

	
	heartRadius := h * 0.7
	heartCenterX := x - 4
	heartCenterY := y + h/2

	
	drawHeartShape(heartCenterX, heartCenterY, heartRadius+2, rl.NewColor(100, 10, 30, 255))
	
	drawHeartShape(heartCenterX, heartCenterY, heartRadius, rl.NewColor(235, 30, 70, 255))
	
	rl.DrawCircle(int32(heartCenterX-heartRadius*0.35), int32(heartCenterY-heartRadius*0.35), heartRadius*0.22, rl.NewColor(255, 180, 200, 255))

	
	hpText := fmt.Sprintf("%d / %d", int(hp.CurrentHP), int(hp.MaxHP))
	textSize := int32(h * 0.5)
	textWidth := rl.MeasureText(hpText, textSize)
	textX := int32(barX + barW/2 - float32(textWidth)/2)
	textY := int32(barY + barH/2 - float32(textSize)/2)

	
	rl.DrawText(hpText, textX+1, textY+1, textSize, rl.NewColor(0, 0, 0, 180))
	rl.DrawText(hpText, textX, textY, textSize, rl.White)
}

func drawHeartShape(cx, cy, r float32, col rl.Color) {
	
	rl.DrawCircle(int32(cx-r*0.45), int32(cy-r*0.25), r*0.55, col)
	
	rl.DrawCircle(int32(cx+r*0.45), int32(cy-r*0.25), r*0.55, col)
	
	v1 := rl.Vector2{X: cx - r * 0.95, Y: cy - r * 0.1}
	v2 := rl.Vector2{X: cx + r * 0.95, Y: cy - r * 0.1}
	v3 := rl.Vector2{X: cx, Y: cy + r * 1.05}
	rl.DrawTriangle(v1, v3, v2, col)
}