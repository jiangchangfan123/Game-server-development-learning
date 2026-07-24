package entities

// 敌人
type Enemy struct {
	*Sprite
	FollowPlayer bool
}
