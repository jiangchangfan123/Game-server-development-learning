package main

import (
	"encoding/json"
	"image"
	"os"
	"path"
)

type TileMapLayerJson struct {
	Data   []int  `json:"data"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Name   string `json:"name"`
}

type TileMapJson struct {
	Layer    []TileMapLayerJson `json:"layers"`
	Tilesets []map[string]any   `json:"tilesets"`
}

func NewTileMapJSON(filepath string) (*TileMapJson, error) {

	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var tilemapJson TileMapJson
	err = json.Unmarshal(content, &tilemapJson)
	if err != nil {
		return nil, err
	}

	return &tilemapJson, nil
}

func (t *TileMapJson) GenTilesets() ([]Tileset, error) {
	tilesets := make([]Tileset, 0)

	for _, tilesetData := range t.Tilesets {
		// convert map relative path to project relative path
		tilesetPath := path.Join("assets/maps/", tilesetData["source"].(string))
		tileset, err := NewTileset(tilesetPath, int(tilesetData["firstgid"].(float64)))
		if err != nil {
			return nil, err
		}

		tilesets = append(tilesets, tileset)
	}

	return tilesets, nil
}

// GenColliders 遍历地图所有图层，生成所有可碰撞瓦片的世界坐标矩形
func (t *TileMapJson) GenColliders(tilesets []Tileset) []image.Rectangle {
	colliders := make([]image.Rectangle, 0)

	for _, layer := range t.Layer {
		for index, id := range layer.Data {
			if id == 0 {
				continue
			}

			// 按 gid 从高到低查找该 tile 属于哪个图集
			for i := len(tilesets) - 1; i >= 0; i-- {
				rect, collidable := tilesets[i].Rect(id)
				if collidable {
					// 计算瓦片在世界坐标中的像素位置
					x := (index % layer.Width) * 16
					y := (index / layer.Width) * 16
					// 将 tile 的局部矩形偏移到世界坐标位置
					colliders = append(colliders, rect.Add(image.Pt(x, y)))
					break
				}
				if rect != (image.Rectangle{}) {
					break // 图集认领了这个 ID 但不可碰撞，跳过
				}
			}
		}
	}

	return colliders
}
