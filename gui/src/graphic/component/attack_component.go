package component

import (
	"strings"

	"tap-gui/src/graphic/component/character"
	"tap-gui/src/graphic/component/ui"
	"tap-gui/src/model/npc"
	"tap-gui/src/service"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AttackComponent struct {
	Manager         *service.AttackManager
	PathResolver    *utils.PathResolver
	MainChar        *character.MainCharacter
	MonsterPreviews map[string]*character.Monster
	PlayerHPBar     *ui.HPBar
	NPCHPBar        *ui.HPBar
}

func NewAttackComponent(mgr *service.AttackManager, pathResolver *utils.PathResolver) *AttackComponent {
	assetPath := pathResolver.Resolve("characters", "main_character")
	mainChar := character.NewMainCharacter(service.CharScale, pathResolver, assetPath)

	return &AttackComponent{
		Manager:         mgr,
		PathResolver:    pathResolver,
		MainChar:        mainChar,
		MonsterPreviews: make(map[string]*character.Monster),
		PlayerHPBar:     ui.NewHPBar(0, 0, 220, 26),
		NPCHPBar:        ui.NewHPBar(0, 0, 220, 26),
	}
}

func (ac *AttackComponent) Unload() {
	if ac.MainChar != nil {
		ac.MainChar.UnloadAnimations()
	}
	for _, monster := range ac.MonsterPreviews {
		if monster != nil {
			monster.UnloadAnimations()
		}
	}
}

func (ac *AttackComponent) getOrCreateMonsterPreview(npc npc.NPCConfig) *character.Monster {
	if monster, exists := ac.MonsterPreviews[npc.Key]; exists {
		return monster
	}

	scale := float32(0.25)
	imageType := strings.ToLower(npc.ImageType)
	if imageType == "" {
		imageType = strings.ToLower(npc.Type)
	}
	if imageType == "" {
		imageType = "gobelin"
	}

	assetPath := ac.PathResolver.Resolve("characters", imageType)
	monster := character.NewMonster(scale, ac.PathResolver, assetPath, npc.Name)
	ac.MonsterPreviews[npc.Key] = monster
	return monster
}

func (ac *AttackComponent) Render(winWidth, winHeight int32) {
	if ac.Manager == nil || !ac.Manager.IsOpen {
		return
	}

	dt := rl.GetFrameTime()

	rl.DrawRectangle(0, 0, winWidth, winHeight, rl.NewColor(utils.ThemeBackground.R, utils.ThemeBackground.G, utils.ThemeBackground.B, 210))

	modalWidth := float32(900)
	modalHeight := float32(600)
	modalX := (float32(winWidth) - modalWidth) / 2
	modalY := (float32(winHeight) - modalHeight) / 2
	modalRect := rl.NewRectangle(modalX, modalY, modalWidth, modalHeight)

	rl.DrawRectangleRounded(modalRect, 0.04, 12, utils.ThemeSurface)
	rl.DrawRectangleRoundedLinesEx(modalRect, 0.04, 12, 3, utils.ThemeBorderStrong)

	topH := float32(280)
	topRect := rl.NewRectangle(modalX+12, modalY+12, modalWidth-24, topH)

	rl.DrawRectangleRounded(topRect, 0.03, 8, utils.ThemeSurfaceInset)
	rl.DrawRectangleRoundedLinesEx(topRect, 0.03, 8, 2, utils.ThemeBorder)

	ac.PlayerHPBar.Position = rl.NewVector2(topRect.X+20, topRect.Y+16)
	ac.PlayerHPBar.Width = 220
	ac.PlayerHPBar.Height = 26
	ac.PlayerHPBar.SetMaxHP(float32(ac.Manager.PlayerMaxHP))
	ac.PlayerHPBar.SetHP(float32(ac.Manager.PlayerHP))
	ac.PlayerHPBar.Update(dt)
	ac.PlayerHPBar.Render()

	ac.NPCHPBar.Position = rl.NewVector2(topRect.X+topRect.Width-240, topRect.Y+16)
	ac.NPCHPBar.Width = 220
	ac.NPCHPBar.Height = 26
	ac.NPCHPBar.SetMaxHP(float32(ac.Manager.NPCMaxHP))
	ac.NPCHPBar.SetHP(float32(ac.Manager.NPCHP))
	ac.NPCHPBar.Update(dt)
	ac.NPCHPBar.Render()

	playerPosX := topRect.X + 260
	monsterPosX := topRect.X + topRect.Width - 260
	arenaPosY := topRect.Y + 170

	if ac.MainChar != nil {
		ac.MainChar.Direction = character.DirRight
		ac.MainChar.SetPosition(playerPosX, arenaPosY)

		if ac.MainChar.Action == character.ActionAttacking {
			if ac.MainChar.AdvanceFrame(dt, false) {
				ac.MainChar.SetAction(character.ActionIdle)
			}
		} else {
			ac.MainChar.AdvanceFrame(dt, true)
		}
		ac.MainChar.Render()
	}

	monster := ac.getOrCreateMonsterPreview(ac.Manager.NPC)
	if monster != nil {
		monster.Direction = character.DirLeft
		monster.SetPosition(monsterPosX, arenaPosY)

		if monster.Action == character.ActionAttacking {
			if monster.AdvanceFrame(dt, false) {
				monster.SetAction(character.ActionIdle)
			}
		} else {
			monster.AdvanceFrame(dt, true)
		}
		monster.Render()
	}

	botY := topRect.Y + topH + 12
	botH := modalHeight - topH - 36
	btnGridW := float32(300)
	logW := modalWidth - btnGridW - 44

	btnGridRect := rl.NewRectangle(modalX+12, botY, btnGridW, botH)
	logRect := rl.NewRectangle(modalX+btnGridW+28, botY, logW, botH)

	btnW := (btnGridW - 12) / 2
	btnH := (botH - 12) / 2

	fleeBtn := ui.NewButton(btnGridRect.X, btnGridRect.Y, btnW, btnH, "FLEE", 16)
	ac.styleActionButton(fleeBtn)
	fleeBtn.Render()
	if fleeBtn.IsClicked {
		ac.Manager.Flee()
	}

	statusBtn := ui.NewButton(btnGridRect.X+btnW+12, btnGridRect.Y, btnW, btnH, "STATUS", 16)
	ac.styleActionButton(statusBtn)
	statusBtn.Render()
	if statusBtn.IsClicked {
		ac.Manager.Status()
	}

	invBtn := ui.NewButton(btnGridRect.X, btnGridRect.Y+btnH+12, btnW, btnH, "INVENTORY", 16)
	ac.styleActionButton(invBtn)
	invBtn.Render()
	if invBtn.IsClicked {
		ac.Manager.UseInventory()
	}

	attackBtn := ui.NewButton(btnGridRect.X+btnW+12, btnGridRect.Y+btnH+12, btnW, btnH, "ATTACK", 16)
	ac.styleActionButton(attackBtn)
	attackBtn.Render()
	if attackBtn.IsClicked {
		ac.Manager.Attack()
		if ac.MainChar != nil {
			ac.MainChar.SetAction(character.ActionAttacking)
		}
		if monster != nil {
			monster.SetAction(character.ActionAttacking)
		}
	}

	rl.DrawRectangleRounded(logRect, 0.08, 8, utils.ThemeSurfaceInset)
	rl.DrawRectangleRoundedLinesEx(logRect, 0.08, 8, 2, utils.ThemeBorder)

	logY := logRect.Y + 10
	lineHeight := float32(18)
	visibleLines := int((logRect.Height - 20) / lineHeight)
	startIdx := 0
	if len(ac.Manager.CombatLog) > visibleLines {
		startIdx = len(ac.Manager.CombatLog) - visibleLines
	}

	for i := startIdx; i < len(ac.Manager.CombatLog); i++ {
		rl.DrawText(ac.Manager.CombatLog[i], int32(logRect.X+12), int32(logY), 12, utils.ThemeText)
		logY += lineHeight
	}
}

func (ac *AttackComponent) styleActionButton(btn *ui.Button) {
	btn.BgColor = utils.ThemeSurfaceRaised
	btn.HoverBgColor = utils.ThemeAccentHover
	btn.ClickedColor = utils.ThemeAccentPressed
	btn.TextColor = utils.ThemeText
	btn.HoverTextColor = utils.ThemeBackground
	btn.BorderColor = utils.ThemeBorder
	btn.BorderWidth = 2
	btn.BorderRadius = 0.2
}
