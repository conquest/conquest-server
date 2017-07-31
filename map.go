package main

type Tile struct {
	Index  uint16 `json:"index"`
	Owner  uint8  `json:"owner"`
	Troops uint32 `json:"troops"`

	X      uint `json:"x"`
	Y      uint `json:"y"`
	Width  uint `json:"w"`
	Height uint `json:"h"`
}

type Region struct {
	Color string `json:"color"`
	Tiles []Tile `json:"tiles"`
}

type Map struct {
	Regions map[string]Region `json:"regions"`
	Scale   float32           `json:"scale"`
}
