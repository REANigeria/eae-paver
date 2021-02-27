package main

import (
	"encoding/json"
	"fmt"
	"github.com/energyaccessexplorer/gdal"
	"strconv"
)

type dataset_info struct {
	Fields       []string   `json:"fields"`
	FeatureCount int        `json:"featurecount"`
	Bounds       [4]float64 `json:"bounds"`
}

func info_featurecount(in filename) int {
	src := gdal.OpenDataSource(in, 0).LayerByIndex(0)
	cs, _ := src.FeatureCount(true)

	return cs
}

func info_bounds(in filename) gdal.Geometry {
	t := gdal.CreateSpatialReference("")
	t.FromEPSG(default_epsg)

	src := gdal.OpenDataSource(in, 0)

	layer := src.LayerByIndex(0)

	env, err := layer.Extent(true)
	if err != nil {
		panic(err)
	}

	// text := fmt.Sprintf(
	// 	"MULTIPOINT (%.32f %.32f, %.32f %.32f)",
	// 	env.MinX(), env.MaxX(), env.MinY(), env.MaxY(),
	// )

	text := fmt.Sprintf(
		`POLYGON ((%.32f %.32f), (%.32f %.32f), (%.32f %.32f), (%.32f %.32f), (%.32f %.32f))`,
		env.MaxX(), env.MinY(), env.MaxX(), env.MaxY(), env.MinX(), env.MaxY(), env.MinX(), env.MinY(),
		env.MaxX(), env.MinY())

	// TODO: this looks like we are going around uselessly...
	//
	layer.ResetReading()
	geom := layer.NextFeature().Geometry()

	v, ok := geom.SpatialReference().AttrValue("AUTHORITY", 1)
	if !ok {
		panic(ok)
	}

	s := gdal.CreateSpatialReference("")
	i, err := strconv.Atoi(v)
	if err != nil {
		panic(err)
	}

	s.FromEPSG(i)

	g, err := gdal.CreateFromWKT(text, s)
	if err != nil {
		panic(err)
	}

	return g
}

func info_fields(in filename) []string {
	fdef := gdal.
		OpenDataSource(in, 0).
		LayerByIndex(0).
		Definition()

	c := fdef.FieldCount()
	a := make([]string, c)

	for i := 0; i < c; i++ {
		a[i] = fdef.FieldDefinition(i).Name()
	}

	return a
}

func info(in filename) string {
	e := info_bounds(in).Envelope()

	i := dataset_info{
		info_fields(in),
		info_featurecount(in),
		[4]float64{e.MinX(), e.MinY(), e.MaxX(), e.MaxY()},
	}

	j, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
	}

	return string(j)
}