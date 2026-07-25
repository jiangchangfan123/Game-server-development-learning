package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"start/Game/entities"
)

type Game struct {
	player      *entities.Player
	enemies     []*entities.Enemy
	potions     []*entities.Potion
	tilemapJSON *TileMapJson
	tilesets    []Tileset
	tilemapImg  *ebiten.Image
	cam         *Camera
	colliders   []image.Rectangle
}

func CheckCollisionHorizontal(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		if collider.Overlaps(
			image.Rect(
				int(sprite.X),
				int(sprite.Y),
				int(sprite.X+16),
				int(sprite.Y+16),
			),
		) {
			if sprite.Dx > 0.0 {
				sprite.X = float64(collider.Min.X) - 16.0
			} else if sprite.Dx < 0.0 {
				sprite.X = float64(collider.Max.X)
			}
		}
	}
}

func CheckCollisionVertical(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		if collider.Overlaps(
			image.Rect(
				int(sprite.X),
				int(sprite.Y),
				int(sprite.X+16),
				int(sprite.Y+16),
			),
		) {
			if sprite.Dy > 0.0 {
				sprite.Y = float64(collider.Min.Y) - 16.0
			} else if sprite.Dy < 0.0 {
				sprite.Y = float64(collider.Max.Y)
			}
		}
	}
}

func (g *Game) Update() error {
	// 重置本帧位移量
	g.player.Dx = 0
	g.player.Dy = 0

	// 水平移动
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.Dx = -g.player.Speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.Dx = g.player.Speed
	}
	g.player.X += g.player.Dx
	CheckCollisionHorizontal(g.player.Sprite, g.colliders)

	// 垂直移动
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Dy = -g.player.Speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Dy = g.player.Speed
	}
	g.player.Y += g.player.Dy
	CheckCollisionVertical(g.player.Sprite, g.colliders)

	// 敌人追踪玩家 + 碰撞检测
	for _, sprite := range g.enemies {
		sprite.Dx = 0
		sprite.Dy = 0

		if sprite.FollowPlayer {
			if sprite.X < g.player.X {
				sprite.Dx = sprite.Speed
			} else if sprite.X > g.player.X {
				sprite.Dx = -sprite.Speed
			}
			if sprite.Y < g.player.Y {
				sprite.Dy = sprite.Speed
			} else if sprite.Y > g.player.Y {
				sprite.Dy = -sprite.Speed
			}
		}

		sprite.X += sprite.Dx
		CheckCollisionHorizontal(sprite.Sprite, g.colliders)
		sprite.Y += sprite.Dy
		CheckCollisionVertical(sprite.Sprite, g.colliders)
	}

	// 药水拾取
	for _, potion := range g.potions {
		potionRect := image.Rect(
			int(potion.X), int(potion.Y),
			int(potion.X+16), int(potion.Y+16),
		)
		playerRect := image.Rect(
			int(g.player.X), int(g.player.Y),
			int(g.player.X+16), int(g.player.Y+16),
		)
		if playerRect.Overlaps(potionRect) {
			g.player.HP += int(potion.AmtHeal)
			fmt.Printf("Picked up potion! Health: %d\n", g.player.HP)
		}
	}

	g.cam.FollowTarget(g.player.X+8, g.player.Y+8, 320, 240)
	g.cam.Constrain(
		float64(g.tilemapJSON.Layer[0].Width)*16.0,
		float64(g.tilemapJSON.Layer[0].Height)*16.0,
		320,
		240,
	)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{120, 180, 255, 255})

	opts := ebiten.DrawImageOptions{}

	for layerIndex, layer := range g.tilemapJSON.Layer {
		for index, id := range layer.Data {
			if id == 0 {
				continue
			}

			x := float64((index % layer.Width) * 16)
			y := float64((index / layer.Width) * 16)

			img := g.tilesets[layerIndex].Img(id)
			if img == nil {
				for i := len(g.tilesets) - 1; i >= 0; i-- {
					img = g.tilesets[i].Img(id)
					if img != nil {
						break
					}
				}
			}
			if img == nil {
				continue
			}

			opts.GeoM.Translate(x, y)
			opts.GeoM.Translate(0.0, -(float64(img.Bounds().Dy()) - 16))
			opts.GeoM.Translate(g.cam.X, g.cam.Y)
			screen.DrawImage(img, &opts)
			opts.GeoM.Reset()
		}
	}

	// 玩家
	opts.GeoM.Translate(g.player.X, g.player.Y)
	opts.GeoM.Translate(g.cam.X, g.cam.Y)
	screen.DrawImage(
		g.player.Img.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image),
		&opts,
	)
	opts.GeoM.Reset()

	// 敌人
	for _, sprite := range g.enemies {
		opts.GeoM.Translate(sprite.X, sprite.Y)
		opts.GeoM.Translate(g.cam.X, g.cam.Y)
		screen.DrawImage(
			sprite.Img.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image),
			&opts,
		)
		opts.GeoM.Reset()
	}

	// 药水
	for _, sprite := range g.potions {
		opts.GeoM.Translate(sprite.X, sprite.Y)
		opts.GeoM.Translate(g.cam.X, g.cam.Y)
		screen.DrawImage(
			sprite.Img.SubImage(image.Rect(0, 0, 16, 16)).(*ebiten.Image),
			&opts,
		)
		opts.GeoM.Reset()
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func NewGame() *Game {
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
	if err != nil {
		log.Fatal(err)
	}

	tilemapJSON, err := NewTileMapJSON("assets/maps/spawn.json")
	if err != nil {
		log.Fatal(err)
	}

	tilesets, err := tilemapJSON.GenTilesets()
	if err != nil {
		log.Fatal(err)
	}

	colliders := tilemapJSON.GenColliders(tilesets)

	return &Game{
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Img:   playerImg,
				X:     50.0,
				Y:     50.0,
				Speed: 2,
			},
			HP: 3,
		},
		enemies: []*entities.Enemy{
			{
				Sprite: &entities.Sprite{
					Img:   skeletonImg,
					X:     100.0,
					Y:     100.0,
					Speed: 1,
				},
				FollowPlayer: true,
			},
			{
				Sprite: &entities.Sprite{
					Img:   skeletonImg,
					X:     150.0,
					Y:     50.0,
					Speed: 1,
				},
				FollowPlayer: false,
			},
		},
		potions: []*entities.Potion{
			{
				Sprite: &entities.Sprite{
					Img: potionImg,
					X:   210.0,
					Y:   100.0,
				},
				AmtHeal: 1,
			},
		},
		tilemapJSON: tilemapJSON,
		tilemapImg:  tilemapImg,
		tilesets:    tilesets,
		colliders:   colliders,
		cam:         NewCamera(0.0, 0.0),
	}
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
