package main

import (
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	PlayerImage *ebiten.Image
	X           float64
	Y           float64
}

/**
只要实现了Update()、Draw(screen *ebiten.Image)、Layout(outsideWidth, outsideHeight int)
这个窗口就被视为Ebat引擎游戏
**/

func (g *Game) Update() error {

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.X += 2
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.X -= 2
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.Y += 2
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.Y -= 2
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(color.RGBA{120, 180, 255, 255}) //将屏幕设置成浅蓝色

	// ebitenutil.DebugPrint(screen, "Hello, World!") //窗口上添加文字

	opts := ebiten.DrawImageOptions{}
	opts.GeoM.Translate(g.X, g.Y)

	screen.DrawImage(
		g.PlayerImage.SubImage(
			image.Rect(0, 0, 16, 16), //截取16x16图像左上角的像素图（子图像）
		).(*ebiten.Image),
		&opts,
	)
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

	if err := ebiten.RunGame(&Game{PlayerImage: playerImg, X: 150, Y: 150}); err != nil {
		log.Fatal(err)
	}
}
