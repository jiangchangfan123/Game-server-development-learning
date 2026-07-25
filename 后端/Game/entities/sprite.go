package entities

import "github.com/hajimehoshi/ebiten/v2"

// 精灵图
type Sprite struct {
	Img   *ebiten.Image
	X     float64
	Y     float64
	Speed float64
	Dx    float64 // 本帧 X 轴位移量，用于碰撞方向判断
	Dy    float64 // 本帧 Y 轴位移量，用于碰撞方向判断
}
