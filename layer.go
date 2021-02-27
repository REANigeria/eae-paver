package main

import (
	"errors"
	"github.com/energyaccessexplorer/gdal"
	"strconv"
)

func fields(in filename) []string {
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

func clip(in filename, container filename) (filename, error) {
	out := _filename()

	src := gdal.OpenDataSource(in, 0).LayerByIndex(0)
	tar := gdal.OpenDataSource(container, 0).LayerByIndex(0)

	ct, _ := tar.FeatureCount(true)
	if ct > 1 {
		return "", errors.New("Clipping only supports single-featured reference datasets. Got " + strconv.Itoa(ct))
	}

	drv := gdal.OGRDriverByName("GeoJSON")
	ds, _ := drv.Create(out, []string{})

	s := gdal.CreateSpatialReference("")
	s.FromEPSG(default_epsg)

	result := ds.CreateLayer("Layer0", s, src.Type(), []string{})

	err := src.Clip(tar, result, []string{})
	if err != nil {
		return "", err
	}

	ds.Destroy()

	return out, err
}
