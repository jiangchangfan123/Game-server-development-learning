package main

import (
	"image"
	"image/color"
	"log"
	"start/Game/entities"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	player      *entities.Player
	enemies     *[]entities.Enemy //敌人群
	potion      *entities.Potion
	tileMapJson *TileMapJson
	tileMapImg  *ebiten.Image
	cam         *Camera
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

	for _, spirte := range *g.enemies {
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

	g.cam.FollowTarget(g.player.X+8, g.player.Y+8, 640, 480)
	g.cam.Constrain(
		float64(g.tileMapJson.Layer[0].Width)*16.0,
		float64(g.tileMapJson.Layer[0].Height)*16.0,
		640,
		480,
	)

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

			mapOpts.GeoM.Translate(float64(x)+g.cam.X, float64(y)+g.cam.Y)

			screen.DrawImage(
				g.tileMapImg.SubImage(image.Rect(srcX, srcY, srcX+16, srcY+16)).(*ebiten.Image),
				&mapOpts,
			)

			mapOpts.GeoM.Reset()
		}
	}

	// 绘制玩家
	playerOpts := ebiten.DrawImageOptions{}
	playerOpts.GeoM.Translate(g.player.X+g.cam.X, g.player.Y+g.cam.Y)
	screen.DrawImage(
		g.player.Img.SubImage(
			image.Rect(0, 0, 16, 16),
		).(*ebiten.Image),
		&playerOpts,
	)

	// 绘制药水
	if g.potion != nil {
		potionOpts := ebiten.DrawImageOptions{}
		potionOpts.GeoM.Translate(g.potion.X+g.cam.X, g.potion.Y+g.cam.Y)
		screen.DrawImage(
			g.potion.Img.SubImage(
				image.Rect(0, 0, 16, 16),
			).(*ebiten.Image),
			&potionOpts,
		)
	}

	// 绘制其他精灵（静态装饰/NPC 等）
	for _, sprite := range *g.enemies {
		opts := ebiten.DrawImageOptions{}
		opts.GeoM.Translate(sprite.X+g.cam.X, sprite.Y+g.cam.Y)
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
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Img:   playerImg,
				X:     150,
				Y:     150,
				Speed: 2,
			},
			HP: 3,
		},
		potion: &entities.Potion{
			Sprite: &entities.Sprite{
				Img:   potionImg,
				X:     300,
				Y:     200,
				Speed: 0,
			},
			AmtHeal: 1,
		},
		tileMapJson: tilemapJson,
		tileMapImg:  tilemapImg,
		enemies: &[]entities.Enemy{
			{
				Sprite: &entities.Sprite{
					Img:   skeletonImg,
					X:     50,
					Y:     50,
					Speed: 1,
				},
				FollowPlayer: true,
			},

			{
				Sprite: &entities.Sprite{
					Img:   skeletonImg,
					X:     100,
					Y:     100,
					Speed: 1,
				},
				FollowPlayer: true,
			},

			{
				Sprite: &entities.Sprite{
					Img:   skeletonImg,
					X:     200,
					Y:     200,
					Speed: 1,
				},
				FollowPlayer: false,
			},
		},
		cam: NewCamera(0, 0),
	}); err != nil {
		log.Fatal(err)
	}
}
