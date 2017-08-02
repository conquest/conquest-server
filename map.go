package main

type City struct {
	Major bool `json:"major"`

	X uint `json:"x"`
	Y uint `json:"y"`
}

type Tile struct {
	Index  uint16 `json:"index"`
	Owner  uint8  `json:"owner"`
	Troops uint32 `json:"troops"`

	X      uint `json:"x"`
	Y      uint `json:"y"`
	Width  uint `json:"w"`
	Height uint `json:"h"`

	City   *City  `json:"city,omitempty"`
	Region string `json:"region,omitempty"`
}

type Region struct {
	Color string `json:"color"`
	Tiles []Tile `json:"tiles"`
}

type Map struct {
	Regions map[string]Region `json:"regions"`
	Scale   float32           `json:"scale"`
}

func (m *Map) Initialize() {
	id := uint16(0)
	for key := range m.Regions {
		for i := range m.Regions[key].Tiles {
			t := &m.Regions[key].Tiles[i]
			t.Index = id
			id++
			t.Troops = 1
		}
	}
}

func (m *Map) Update() {
	for key := range m.Regions {
		for i := range m.Regions[key].Tiles {
			t := &m.Regions[key].Tiles[i]
			if t.City != nil {
				if t.City.Major {
					t.Troops += 5
				} else {
					t.Troops += 2
				}
			}
			t.Troops += 2
		}
	}
}

func (m *Map) Reset() {
	for key := range m.Regions {
		for i := range m.Regions[key].Tiles {
			t := &m.Regions[key].Tiles[i]
			t.Owner = 0
			t.Troops = 1
		}
	}

}

func (m *Map) Read(tiles []Tile) {
	for _, tile := range tiles {
		reg := m.Regions[tile.Region]
		for _, t := range reg.Tiles {
			if tile.Index == t.Index {
				t.Owner = tile.Owner
				t.Troops = tile.Troops
				break
			}
		}
	}
}
