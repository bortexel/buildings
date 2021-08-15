package litematica

import (
	"io"

	"github.com/Tnze/go-mc/nbt"
)

type Project struct {
	Metadata             Metadata
	MinecraftDataVersion int
	Version              int
	Regions              map[string]Region
}

type Metadata struct {
	Author        string
	Description   string
	EnclosingSize Vec3D
	Name          string
	RegionCount   int
	TimeCreated   int64
	TimeModified  int64
	TotalBlocks   int64
	TotalVolume   int64
}

type Vec3D struct {
	X int `nbt:"x"`
	Y int `nbt:"y"`
	Z int `nbt:"z"`
}

type Region struct {
	BlockStatePalette []CompoundTag
	TileEntities      []CompoundTag
	Position          Vec3D
	Size              Vec3D
	BlockStates       []int64
}

type CompoundTag struct {
	Name string
}

func Load(reader io.Reader) (*Project, error) {
	var project *Project
	_, err := nbt.NewDecoder(reader).Decode(&project)
	return project, err
}
