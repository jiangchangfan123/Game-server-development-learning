package entities

import "github.com/hajimehoshi/ebiten/v2"

// 精灵图
type Sprite struct {
	Img   *ebiten.Image
	X     float64
	Y     float64
	Speed float64
}
