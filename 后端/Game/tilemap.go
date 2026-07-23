package main

import (
	"encoding/json"
	"os"
)

type TileMapLayerJson struct {
	Data   []int `json:"data"`
	Width  int   `json:"width"`
	Height int   `json:"height"`
}

type TileMapJson struct {
	Layer []TileMapLayerJson `json:"layers"`
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
