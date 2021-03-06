package main

import (
	"fmt"
	"net/http"
)

type reporter func(string, ...interface{})

func routine_admin_boundaries(r *http.Request, in filename, idfield string, resolution int) (string, error) {
	w := func(s string, x ...interface{}) {
		socketwrite(fmt.Sprintf(s+"\n", x...), r)
	}

	rprj, err := vectors_reproject(in, 3857)
	if err != nil {
		return "", err
	}
	w("%s <- reprojected", rprj)

	ids, err := raster_ids(rprj, idfield, resolution)
	if err != nil {
		return "", err
	}
	w("%s <- *raster ids", ids)

	stripped, err := vectors_strip(rprj, []string{idfield})
	if err != nil {
		return "", err
	}
	w("%s <- stripped", stripped)

	rprjstripped, err := vectors_reproject(in, 4326)
	if err != nil {
		return "", err
	}
	w("%s <- *stripped reprojected", rprjstripped)

	info := info(rprjstripped)

	w("CLEAN UP")

	trash(rprj, stripped)

	if run_server {
		keeps := []filename{ids, rprjstripped}

		for _, f := range keeps {
			w("%s -> S3", f)
			s3put(f)
			trash(f)
		}
	}

	w("DONE")

	jsonstr := fmt.Sprintf(
		`{ "vectors": "%s", "raster": "%s", "info": %s }`,
		_uuid(rprjstripped),
		_uuid(ids),
		info,
	)

	return jsonstr, nil
}

func routine_clip_proximity(r *http.Request, in filename, ref filename, fields []string, resolution int) (string, error) {
	w := func(s string, x ...interface{}) {
		socketwrite(fmt.Sprintf(s+"\n", x...), r)
	}

	stripped, err := vectors_strip(in, fields)
	if err != nil {
		return "", err
	}
	w("%s <- stripped", stripped)

	refprj, err := vectors_reproject(ref, 3857)
	if err != nil {
		return "", err
	}
	w("%s <- reprojected reference", refprj)

	zeros, err := raster_zeros(refprj, resolution)
	if err != nil {
		return "", err
	}
	w("%s <- zeros", zeros)

	clipped, err := vectors_clip(stripped, ref, w)
	if err != nil {
		return "", err
	}
	w("%s <- *clipped", clipped)

	rstr, err := raster_geometry(clipped, zeros)
	if err != nil {
		return "", err
	}
	w("%s <- rasterised <- zeros", rstr) // overwrites zeros

	prox, err := raster_proximity(rstr)
	if err != nil {
		return "", err
	}
	w("%s <- *proximity", prox)

	w("CLEAN UP")
	trash(in, ref, stripped, rstr, refprj)

	if run_server {
		keeps := []filename{clipped, prox}

		for _, f := range keeps {
			w("%s -> S3", f)
			s3put(f)
			trash(f)
		}
	}

	w("DONE")

	jsonstr := fmt.Sprintf(`{ "vectors": "%s", "raster": "%s" }`, _uuid(clipped), _uuid(prox))

	return jsonstr, nil
}

func routine_crop_raster(r *http.Request, in filename, base filename, ref filename, conf string) (string, error) {
	w := func(s string, x ...interface{}) {
		socketwrite(fmt.Sprintf(s+"\n", x...), r)
	}

	cropped, err := raster_crop(in, base, ref, conf, w)
	if err != nil {
		return "", err
	}
	w("%s <- cropped", cropped)

	w("CLEAN UP")

	if run_server {
		keeps := []filename{cropped}

		for _, f := range keeps {
			w("%s -> S3", f)
			s3put(f)
			trash(f)
		}
	}

	w("DONE")

	jsonstr := fmt.Sprintf(`{ "raster": "%s" }`, _uuid(cropped))

	return jsonstr, nil
}
