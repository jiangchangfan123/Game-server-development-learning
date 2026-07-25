package main

import (
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Tileset 图集接口：所有图集必须实现此接口
// 输入全局 tile ID（来自 Tiled 地图 data 数组），返回对应图片，找不到返回 nil
type Tileset interface {
	Img(id int) *ebiten.Image          // 根据全局 tile ID 返回图片
	Rect(id int) (image.Rectangle, bool) // 返回 tile 的世界坐标矩形和是否可碰撞
}

// UniformTilesetJSON 均匀图集的 JSON 结构
// 对应 Tiled 导出的标准单图图集，例：TilesetFloor.json → {"image":"../../images/TilesetFloor.png"}
type UniformTilesetJSON struct {
	Path string `json:"image"` // 图集大图的相对路径
}

// UniformTileset 均匀图集：一张大图，所有 tile 等距平铺在里面
// 例：TilesetFloor.png（352x417），每 16x16 一个 tile，每行 22 个，共 572 个
type UniformTileset struct {
	img *ebiten.Image // 整张大图的纹理
	gid int           // 该图集在大地图中的起始 ID（Tiled 的 firstgid）
}

// Img 根据全局 tile ID 返回大图中对应的小 tile 图片
func (u *UniformTileset) Img(id int) *ebiten.Image {
	// 1. 全局 ID → 局部 ID（去掉 firstgid 偏移）
	id -= u.gid
	if id < 0 { // ID 不属于本图集（小于 firstgid）
		return nil
	}

	// 2. 计算 tile 在大图里的行列位置（假设每行 22 个 tile）
	srcX := id % 22 // 列号
	srcY := id / 22 // 行号

	// 3. 行列 → 像素坐标（每个 tile 16x16）
	srcX *= 16
	srcY *= 16

	// 4. 边界检查：源坐标不能超出大图范围
	bounds := u.img.Bounds()
	if srcX >= bounds.Dx() || srcY >= bounds.Dy() {
		return nil
	}

	// 5. 从大图中裁剪 16×16 的 tile 图片并返回
	return u.img.SubImage(
		image.Rect(
			srcX, srcY, srcX+16, srcY+16,
		),
	).(*ebiten.Image)
}

// Rect 返回该 tile 的像素尺寸矩形，以及是否可碰撞
// 均匀图集（地面）全部不可碰撞
func (u *UniformTileset) Rect(id int) (image.Rectangle, bool) {
	id -= u.gid
	if id < 0 { // ID 不属于本图集
		return image.Rectangle{}, false
	}
	return image.Rect(0, 0, 16, 16), false // 地面瓦片，不碰撞
}

// TileJSON 动态图集中单个 tile 的 JSON 结构
// 例：{"id":0, "image":"../../images/buildings/building1.png", "imagewidth":64, "imageheight":48}
type TileJSON struct {
	Id     int    `json:"id"`          // tile 在本图集中的局部编号
	Path   string `json:"image"`       // 该 tile 独立图片的路径
	Width  int    `json:"imagewidth"`  // 图片宽度（像素），不同 tile 可能不同
	Height int    `json:"imageheight"` // 图片高度（像素），不同 tile 可能不同
}

// DynTilesetJSON 动态图集的 JSON 结构
// 例：buildings.json → {"tiles": [tile1, tile2, tile3]}
type DynTilesetJSON struct {
	Tiles []*TileJSON `json:"tiles"` // 每个 tile 独立定义一张图片
}

// DynTileset 动态图集：每个 tile 是独立的 PNG 图片，尺寸各不相同
// 适用于不规则物件，如建筑、大型 NPC 等
type DynTileset struct {
	imgs []*ebiten.Image // 按 tile 局部 ID 顺序存储的图片数组
	gid  int             // 该图集在大地图中的起始 ID（Tiled 的 firstgid）
}

// Img 根据全局 tile ID 返回对应的图片
func (d *DynTileset) Img(id int) *ebiten.Image {
	// 1. 全局 ID → 局部索引
	id -= d.gid
	// 2. 越界检查（ID 不属于本图集）
	if id < 0 || id >= len(d.imgs) {
		return nil
	}

	// 3. 直接按索引返回预加载的图片
	return d.imgs[id]
}

// Rect 返回该 tile 的像素尺寸矩形，以及是否可碰撞
// 动态图集（建筑）全部可碰撞，尺寸各不相同
func (d *DynTileset) Rect(id int) (image.Rectangle, bool) {
	id -= d.gid
	if id < 0 || id >= len(d.imgs) { // ID 不属于本图集
		return image.Rectangle{}, false
	}
	bounds := d.imgs[id].Bounds()
	return bounds, true // 建筑瓦片，可碰撞
}

// NewTileset 工厂函数：读取图集 JSON 文件，自动识别类型并加载图片
// path：图集 JSON 文件路径，如 "assets/maps/tilesets/TilesetFloor.json"
// gid：该图集在大地图中的全局起始 ID（Tiled 的 firstgid）
func NewTileset(path string, gid int) (Tileset, error) {
	// 读取图集 JSON 文件内容
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 通过文件名判断图集类型（硬编码规则：包含 "buildings" 的是动态图集）
	if strings.Contains(path, "buildings") {
		// === 动态图集：每个 tile 是独立图片 ===

		// 1. 反序列化 JSON
		var dynTilesetJSON DynTilesetJSON
		err = json.Unmarshal(contents, &dynTilesetJSON)
		if err != nil {
			return nil, err
		}

		// 2. 创建图集对象，设置起始 ID
		dynTileset := DynTileset{}
		dynTileset.gid = gid
		dynTileset.imgs = make([]*ebiten.Image, 0)

		// 3. 遍历每个 tile，加载其独立的图片
		for _, tileJSON := range dynTilesetJSON.Tiles {

			//  路径转换：Tiled 相对路径 → 项目根路径
			//  Tiled JSON 在 assets/maps/tilesets/ 下，图片引用是 ../../images/xxx.png
			//  需要去掉 ../ 前缀，拼接为 assets/images/xxx.png
			tileJSONPath := tileJSON.Path
			tileJSONPath = filepath.Clean(tileJSONPath)            // 清理路径中的 . 和 ..
			tileJSONPath = strings.ReplaceAll(tileJSONPath, "\\", "/") // 统一为正斜杠
			tileJSONPath = strings.TrimPrefix(tileJSONPath, "../")     // 上溯一层
			tileJSONPath = strings.TrimPrefix(tileJSONPath, "../")     // 上溯两层
			tileJSONPath = filepath.Join("assets/", tileJSONPath)      // 拼接项目根路径

			// 加载图片文件
			img, _, err := ebitenutil.NewImageFromFile(tileJSONPath)
			if err != nil {
				return nil, err
			}

			// 按 tile 局部 ID 顺序追加到数组
			dynTileset.imgs = append(dynTileset.imgs, img)
		}

		return &dynTileset, nil
	}

	// === 均匀图集：一张大图包含所有 tile ===

	// 1. 反序列化 JSON（只需 image 字段）
	var uniformTilesetJSON UniformTilesetJSON
	err = json.Unmarshal(contents, &uniformTilesetJSON)
	if err != nil {
		return nil, err
	}

	uniformTileset := UniformTileset{}

	// 2. 路径转换：同动态图集的逻辑
	tileJSONPath := uniformTilesetJSON.Path
	tileJSONPath = filepath.Clean(tileJSONPath)
	tileJSONPath = strings.ReplaceAll(tileJSONPath, "\\", "/")
	tileJSONPath = strings.TrimPrefix(tileJSONPath, "../")
	tileJSONPath = strings.TrimPrefix(tileJSONPath, "../")
	tileJSONPath = filepath.Join("assets/", tileJSONPath)

	// 3. 加载整张大图
	img, _, err := ebitenutil.NewImageFromFile(tileJSONPath)
	if err != nil {
		return nil, err
	}
	uniformTileset.img = img
	uniformTileset.gid = gid // 记录 firstgid，后续 Img() 用它做 ID 偏移

	return &uniformTileset, nil
}
