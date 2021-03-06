package main

import (
	"encoding/json"
	"fmt"
	"github.com/energyaccessexplorer/gdal"
	"strconv"
)

type raster_config struct {
	Numbertype string `json:"numbertype"`
	Nodata     int    `json:"nodata"`
	Resample   string `json:"resample"`
}

func raster_ids(in filename, gid string, resolution int) (filename, error) {
	src, err := gdal.OpenEx(in, gdal.OFReadOnly, nil, nil, nil)
	if err != nil {
		return "", err
	}
	defer src.Close()

	out := _filename()

	res := strconv.Itoa(resolution)

	opts := []string{
		"-a", gid,
		"-a_srs", "EPSG:3857",
		"-a_nodata", "-1",
		"-tr", res, res,
		"-of", "GTiff",
		"-ot", "Int16",
		"-co", "COMPRESS=DEFLATE",
		"-co", "PREDICTOR=1",
		"-co", "ZLEVEL=9",
	}

	dest, err := gdal.Rasterize(out, src, opts)
	if err != nil {
		return "", err
	}
	dest.Close()

	return out, err
}

func raster_geometry(in filename, dst filename) (filename, error) {
	src, err := gdal.OpenEx(in, gdal.OFReadOnly, nil, nil, nil)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dest, err := gdal.OpenEx(dst, gdal.OFUpdate, nil, nil, nil)
	if err != nil {
		return "", err
	}

	opts := []string{
		"-l", gdal.OpenDataSource(in, 0).LayerByIndex(0).Name(),
		"-burn", "1",
	}

	out, err := gdal.RasterizeOverwrite(dest, src, opts)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// TODO: dest.Close() segfaults... defer o no defer below, no comprende

	return dst, err
}

func raster_proximity(in filename) (filename, error) {
	src, err := gdal.OpenEx(in, gdal.OFReadOnly, nil, nil, nil)
	if err != nil {
		return "", err
	}

	drv, err := gdal.GetDriverByName("GTiff")
	if err != nil {
		return "", err
	}

	out := _filename()

	opts := []string{
		"-ot", "Int16",
		"-nodata", "-1",
		"DISTUNITS=PIXEL",
		"VALUES=1",
		"USE_INPUT_NODATA=YES",
		fmt.Sprintf("MAXDIST=%d", 512),
		"-co", "COMPRESS=DEFLATE",
		"-co", "PREDICTOR=1",
		"-co", "ZLEVEL=9",
	}

	ds := drv.CreateCopy(out, src, 0, []string{}, gdal.DummyProgress, nil)
	err = src.
		RasterBand(1).
		ComputeProximity(ds.RasterBand(1), opts, gdal.DummyProgress, nil)

	ds.Close()

	return out, err
}

func raster_zeros(in filename, resolution int) (filename, error) {
	src, err := gdal.OpenEx(in, gdal.OFReadOnly, nil, nil, nil)
	if err != nil {
		return "", err
	}
	defer src.Close()

	out := _filename()

	res := strconv.Itoa(resolution)

	opts := []string{
		"-burn", "0",
		"-a_nodata", "-1",
		"-a_srs", "EPSG:3857",
		"-tr", res, res,
		"-of", "GTiff",
		"-ot", "Int16",
	}

	dest, err := gdal.Rasterize(out, src, opts)
	if err != nil {
		return "", err
	}
	defer dest.Close()

	return out, err
}

func raster_crop(in filename, base filename, ref filename, conf string, w reporter) (filename, error) {
	w("RASTER CROP")

	var c raster_config
	err := json.Unmarshal([]byte(conf), &c)

	r, err := gdal.OpenEx(base, gdal.OFReadOnly, nil, nil, nil)
	if err != nil {
		return "", err
	}
	defer r.Close()

	src, err := gdal.OpenEx(in, gdal.OFReadOnly, nil, nil, nil)
	if err != nil {
		return "", err
	}
	defer src.Close()

	out := _filename()

	layer := gdal.OpenDataSource(ref, 0).LayerByIndex(0).Name()
	w(" cropping to first layer: %s", layer)

	x := r.RasterXSize()
	y := r.RasterYSize()
	w(" raster size: (%d,%d)", x, y)

	w(" numbertype: %s", c.Numbertype)
	w(" nodata: %s", c.Nodata)
	w(" resampling method: %s", c.Resample)

	opts := []string{
		"-cutline", ref,
		"-crop_to_cutline",
		"-cl", layer,
		"-of", "GTiff",
		"-ts", strconv.Itoa(x), strconv.Itoa(y),
		"-t_srs", "EPSG:3857",
		"-ot", c.Numbertype,
		"-dstnodata", strconv.Itoa(c.Nodata),
		"-r", c.Resample,
		"-co", "COMPRESS=DEFLATE",
		"-co", "PREDICTOR=1",
		"-co", "ZLEVEL=9",
	}

	dest, err := gdal.Warp(out, []gdal.Dataset{src}, opts)
	if err != nil {
		return "", err
	}
	defer dest.Close()

	return out, nil
}
