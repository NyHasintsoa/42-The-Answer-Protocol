package component

import (
	"strings"

	"tap-gui/src/model/enums"
	"tap-gui/src/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ItemComponent struct {
	Resolver *utils.PathResolver
	textures map[string]rl.Texture2D
}

func NewItemComponent(resolver *utils.PathResolver) *ItemComponent {
	return &ItemComponent{
		Resolver: resolver,
		textures: make(map[string]rl.Texture2D),
	}
}

func (ic *ItemComponent) GetTextureByItemType(itemType enums.InventoryItemType) rl.Texture2D {
	return ic.GetTextureByName(itemType.Filename())
}

func (ic *ItemComponent) GetTextureByName(filename string) rl.Texture2D {
	if !strings.HasSuffix(filename, ".png") {
		filename += ".png"
	}
	if tex, exists := ic.textures[filename]; exists {
		return tex
	}

	var imagePath string
	if ic.Resolver != nil {
		imagePath = ic.Resolver.Resolve("inventory", filename)
	} else {
		imagePath = "assets/inventory/" + filename
	}

	tex := ic.Resolver.ImageManager.Load(imagePath)
	ic.textures[filename] = tex
	return tex
}

func (ic *ItemComponent) RenderItemIcon(filename string, destRect rl.Rectangle) {
	tex := ic.GetTextureByName(filename)
	if tex.ID > 0 {
		srcRec := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
		rl.DrawTexturePro(tex, srcRec, destRect, rl.NewVector2(0, 0), 0, rl.White)
	} else {
		rl.DrawRectangleRounded(destRect, 0.2, 4, rl.NewColor(50, 60, 80, 255))
	}
}

func (ic *ItemComponent) Unload() {
	ic.textures = make(map[string]rl.Texture2D)
}
