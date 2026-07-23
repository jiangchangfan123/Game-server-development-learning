package main

import (
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// 精灵图
type Sprite struct {
	Img   *ebiten.Image
	X     float64
	Y     float64
	Speed float64
}

// 玩家
type Player struct {
	*Sprite
	HP int
}

// 敌人
type Enemy struct {
	*Sprite
	FollowPlayer bool
}

// 药水
type Potion struct {
	*Sprite
	AmtHeal uint
}

type Game struct {
	player      *Player
	sprites     *[]Enemy //敌人群
	potion      *Potion
	tileMapJson *TileMapJson
	tileMapImg  *ebiten.Image
}

/**
只要实现了Update()、Draw(screen *ebiten.Image)、Layout(outsideWidth, outsideHeight int)
这个窗口就被视为Ebat引擎游戏
**/

func (g *Game) Update() error {

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.X += g.player.Speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.X -= g.player.Speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Y += g.player.Speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Y -= g.player.Speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
		g.player.Speed = 3
	} else {
		g.player.Speed = 2
	}

	for _, spirte := range *g.sprites {
		if spirte.FollowPlayer {
			if spirte.X < g.player.X {
				spirte.X = spirte.X + spirte.Speed
			} else {
				spirte.X = spirte.X - spirte.Speed
			}

			if spirte.Y < g.player.Y {
				spirte.Y = spirte.Y + spirte.Speed
			} else {
				spirte.Y = spirte.Y - spirte.Speed
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(color.RGBA{120, 180, 255, 255}) //将屏幕设置成浅蓝色

	// ebitenutil.DebugPrint(screen, "Hello, World!") //窗口上添加文字

	//绘制地图

	mapOpts := ebiten.DrawImageOptions{}
	for _, layer := range g.tileMapJson.Layer {
		for index, id := range layer.Data {
			x := index % layer.Width
			y := index / layer.Width

			x *= 16
			y *= 16

			srcX := (id - 1) % 22
			srcY := (id - 1) / 22

			srcX *= 16
			srcY *= 16

			mapOpts.GeoM.Translate(float64(x), float64(y))

			screen.DrawImage(
				g.tileMapImg.SubImage(image.Rect(srcX, srcY, srcX+16, srcY+16)).(*ebiten.Image),
				&mapOpts,
			)

			mapOpts.GeoM.Reset()
		}
	}

	// 绘制玩家
	playerOpts := ebiten.DrawImageOptions{}
	playerOpts.GeoM.Translate(g.player.X, g.player.Y)
	screen.DrawImage(
		g.player.Img.SubImage(
			image.Rect(0, 0, 16, 16),
		).(*ebiten.Image),
		&playerOpts,
	)

	// 绘制药水
	if g.potion != nil {
		potionOpts := ebiten.DrawImageOptions{}
		potionOpts.GeoM.Translate(g.potion.X, g.potion.Y)
		screen.DrawImage(
			g.potion.Img.SubImage(
				image.Rect(0, 0, 16, 16),
			).(*ebiten.Image),
			&potionOpts,
		)
	}

	// 绘制其他精灵（静态装饰/NPC 等）
	for _, sprite := range *g.sprites {
		opts := ebiten.DrawImageOptions{}
		opts.GeoM.Translate(sprite.X, sprite.Y)
		screen.DrawImage(
			sprite.Img.SubImage(
				image.Rect(0, 0, 16, 16),
			).(*ebiten.Image),
			&opts,
		)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ebiten.WindowSize()
}

func main() {
	ebiten.SetWindowSize(640, 480)                                 //设置窗口大小
	ebiten.SetWindowTitle("Hello, World!")                         //设置窗口标题
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled) //设置窗口可自行调节大小

	playerImg, _, err := ebitenutil.NewImageFromFile("assets/images/ninja.png")
	if err != nil {
		log.Fatal(err)
	}

	skeletonImg, _, err := ebitenutil.NewImageFromFile("assets/images/skeleton.png")
	if err != nil {
		log.Fatal(err)
	}

	potionImg, _, err := ebitenutil.NewImageFromFile("assets/images/potion.png")
	if err != nil {
		log.Fatal(err)
	}

	tilemapImg, _, err := ebitenutil.NewImageFromFile("assets/images/TilesetFloor.png")

	tilemapJson, err := NewTileMapJSON("assets/maps/spawn.json")
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(&Game{
		player: &Player{
			Sprite: &Sprite{
				Img:   playerImg,
				X:     150,
				Y:     150,
				Speed: 2,
			},
			HP: 3,
		},
		potion: &Potion{
			Sprite: &Sprite{
				Img:   potionImg,
				X:     300,
				Y:     200,
				Speed: 0,
			},
			AmtHeal: 1,
		},
		tileMapJson: tilemapJson,
		tileMapImg:  tilemapImg,
		sprites: &[]Enemy{
			{
				&Sprite{
					Img:   skeletonImg,
					X:     50,
					Y:     50,
					Speed: 1,
				},
				true,
			},

			{
				&Sprite{
					Img:   skeletonImg,
					X:     100,
					Y:     100,
					Speed: 1,
				},
				true,
			},

			{
				&Sprite{
					Img:   skeletonImg,
					X:     200,
					Y:     200,
					Speed: 1,
				},
				false,
			},
		},
	}); err != nil {
		log.Fatal(err)
	}
}
