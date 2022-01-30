package geoserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Range struct {
	Low  []int
	High []int
}

func (r *Range) UnmarshalJSON(data []byte) error {

	var rdata map[string]string

	err := json.Unmarshal(data, &rdata)
	if err != nil {
		return err
	}

	rangeArray := func(d string) (res []int, err error) {
		strArr := strings.Split(d, " ")
		intArr := make([]int, 0, len(strArr))
		for _, v := range strArr {
			iv, err := strconv.Atoi(v)
			if err != nil {
				return res, err
			}
			intArr = append(intArr, iv)
		}
		return intArr, nil
	}

	*r = Range{}
	r.High, err = rangeArray(rdata["high"])
	if err != nil {
		return err
	}
	r.Low, err = rangeArray(rdata["low"])
	if err != nil {
		return err
	}

	return nil
}

type Transform struct {
	ScaleX     float64 `json:"scaleX"`
	ScaleY     float64 `json:"scaleY"`
	ShearX     float64 `json:"shearX"`
	ShearY     float64 `json:"shearY"`
	TranslateX float64 `json:"translateX"`
	TranslateY float64 `json:"translateY"`
}

type Grid struct {
	Dimension int        `json:"@dimension,omitempty,string"`
	Range     *Range     `json:"range,omitempty"`
	Transform *Transform `json:"transform,omitempty"`
}

// Coverage is geoserver Coverage (raster layer) data struct
type Coverage struct {
	Name                 string             `json:"name,omitempty"`
	NativeCoverageName   string             `json:"nativeCoverageName,omitempty"`
	NativeName           string             `json:"nativeName,omitempty"`
	NativeFormat         string             `json:"nativeFormat,omitempty"`
	Namespace            *Resource          `json:"namespace,omitempty"`
	Title                string             `json:"title,omitempty"`
	Description          string             `json:"description,omitempty"`
	Abstract             string             `json:"abstract,omitempty"`
	Keywords             *Keywords          `json:"keywords,omitempty"`
	NativeCRS            *CRSType           `json:"nativeCRS,omitempty"`
	Srs                  string             `json:"srs,omitempty"`
	Enabled              bool               `json:"enabled,omitempty"`
	NativeBoundingBox    *NativeBoundingBox `json:"nativeBoundingBox,omitempty"`
	LatLonBoundingBox    *LatLonBoundingBox `json:"latLonBoundingBox,omitempty"`
	ProjectionPolicy     string             `json:"projectionPolicy,omitempty"`
	Store                *Resource          `json:"store,omitempty"`
	CqlFilter            string             `json:"cqlFilter,omitempty"`
	OverridingServiceSRS bool               `json:"overridingServiceSRS,omitempty"`
	Grid                 *Grid              `json:"grid,omitempty"`
	//Metadata               *Metadata          `json:"metadata,omitempty"`  //need to fix the implementation due to json parse error
	//SupportedFormats       []string			  `json:"supportedFormats,omitempty"`  //need to fix the implementation due to json parse error
}

type publishedCoverageDescr struct {
	Name               string `json:"name,omitempty"`
	NativeCoverageName string `json:"nativeCoverageName,omitempty"`
}

type publishCoverageRequest struct {
	CoverageDescr *publishedCoverageDescr `json:"coverage,omitempty"`
}

// GetCoverages returns all published raster layers (coverages) for workspace as resources,
// err is an error if error occurred else err is nil
func (g *GeoServer) GetCoverages(workspaceName string) (coverages []*Resource, err error) {
	targetURL := g.ParseURL("rest", "workspaces", workspaceName, "coverages")
	httpRequest := HTTPRequest{
		Method: getMethod,
		Accept: jsonType,
		URL:    targetURL,
		Query:  nil,
	}
	response, responseCode := g.DoRequest(httpRequest)
	if responseCode != statusOk {
		g.logger.Error(string(response))
		err = g.GetError(responseCode, response)
		return
	}

	var coveragesResponse struct {
		Coverages struct {
			Coverage []*Resource `json:"coverage,omitempty"`
		} `json:"coverages,omitempty"`
	}

	var coveragesEmptyResponse struct {
		Coverages string
	}

	if err = json.Unmarshal(response, &coveragesResponse); err != nil {
		if err = g.DeSerializeJSON(response, &coveragesEmptyResponse); err != nil {
			return nil, fmt.Errorf("can't parse the coverage data, %v", err)
		} else {
			return []*Resource{}, nil
		}
	}

	return coveragesResponse.Coverages.Coverage, nil
}

// GetStoreCoverages returns a list for all coverages (raster layers) names including unpublished for coverageStore,
// err is an error if error occurred else err is nil
func (g *GeoServer) GetStoreCoverages(workspaceName string, coverageStore string) (coverages []string, err error) {
	targetURL := g.ParseURL("rest", "workspaces", workspaceName, "coveragestores", coverageStore, "coverages")
	httpRequest := HTTPRequest{
		Method: getMethod,
		Accept: jsonType,
		URL:    targetURL,
		Query:  map[string]string{"list": "all"},
	}
	response, responseCode := g.DoRequest(httpRequest)
	if responseCode != statusOk {
		g.logger.Error(string(response))
		err = g.GetError(responseCode, response)
		return
	}

	var coveragesResponse struct {
		List struct {
			CoverageName []string `json:"string,omitempty"`
		} `json:"list,omitempty"`
	}

	if err = g.DeSerializeJSON(response, &coveragesResponse); err != nil {
		return nil, fmt.Errorf("can't parse the coverages data, %v", err)
	}

	return coveragesResponse.List.CoverageName, nil
}

// GetCoverage returns the coverage with name coverageName
// err is an error if error occurred else err is nil
func (g *GeoServer) GetCoverage(workspaceName string, coverageName string) (coverage *Coverage, err error) {
	targetURL := g.ParseURL("rest", "workspaces", workspaceName, "coverages", coverageName)
	httpRequest := HTTPRequest{
		Method: getMethod,
		Accept: jsonType,
		URL:    targetURL,
		Query:  nil,
	}
	response, responseCode := g.DoRequest(httpRequest)
	if responseCode != statusOk {
		g.logger.Error(string(response))
		err = g.GetError(responseCode, response)
		return
	}

	var coverageResponse struct {
		Coverage Coverage
	}

	if err = g.DeSerializeJSON(response, &coverageResponse); err != nil {
		return nil, fmt.Errorf("can't parse the coverage data, %v", err)
	}

	return &coverageResponse.Coverage, nil
}

// DeleteCoverage removes the coverage,
// err is an error if error occurred else err is nil
func (g *GeoServer) DeleteCoverage(workspaceName string, layerName string, recurse bool) (deleted bool, err error) {
	//it's just a wrapper about DeleteLayer function as it does the same in the most use cases
	return g.DeleteLayer(workspaceName, layerName, recurse)
}

//UpdateCoverage updates geoserver coverage (raster layer), else returns error,
func (g *GeoServer) UpdateCoverage(workspaceName string, coverage *Coverage) (modified bool, err error) {

	items := strings.Split(coverage.Store.Name, ":")
	if len(items) != 2 {
		return false, errors.New("internal error during coverage update, can't build store name")
	}
	targetURL := g.ParseURL("rest", "workspaces", workspaceName, "coveragestores", items[1], "coverages", coverage.Name)

	type CoverageUpdate struct {
		Name              string             `json:"name,omitempty"`
		Title             string             `json:"title,omitempty"`
		Description       string             `json:"description,omitempty"`
		Abstract          string             `json:"abstract,omitempty"`
		Keywords          *Keywords          `json:"keywords,omitempty"`
		Enabled           bool               `json:"enabled,omitempty"`
		NativeBoundingBox *NativeBoundingBox `json:"nativeBoundingBox,omitempty"`
		LatLonBoundingBox *LatLonBoundingBox `json:"latLonBoundingBox,omitempty"`
	}

	type coverageUpdateRequestBody struct {
		Coverage CoverageUpdate `json:"coverage,omitempty"`
	}

	data := coverageUpdateRequestBody{Coverage: CoverageUpdate{
		Name:              coverage.Name,
		Title:             coverage.Title,
		Description:       coverage.Description,
		Abstract:          coverage.Abstract,
		Keywords:          coverage.Keywords,
		Enabled:           coverage.Enabled,
		NativeBoundingBox: coverage.NativeBoundingBox,
		LatLonBoundingBox: coverage.LatLonBoundingBox,
	}}

	serializedLayer, _ := g.SerializeStruct(data)
	httpRequest := HTTPRequest{
		Method:   putMethod,
		Accept:   jsonType,
		Data:     bytes.NewBuffer(serializedLayer),
		DataType: jsonType,
		URL:      targetURL,
		Query:    nil,
	}
	response, responseCode := g.DoRequest(httpRequest)
	if responseCode != statusOk {
		g.logger.Error(string(response))
		err = g.GetError(responseCode, response)
		return
	}
	modified = true
	return
}

// PublishCoverage publishes coverage from coverageStore
// coverageName - the name of the layer in the coverageStore (use GetStoreCoverages to get them), publishName - the name it was presented at geoserver
func (g *GeoServer) PublishCoverage(workspaceName string, coverageStoreName string, coverageName string, publishName string) (published bool, err error) {

	if publishName == "" {
		publishName = coverageName
	}

	publishRequest := publishCoverageRequest{
		&publishedCoverageDescr{
			Name:               publishName,
			NativeCoverageName: coverageName,
		},
	}
	return g.publishCoverage(workspaceName, coverageStoreName, publishRequest)
}

// publishCoverage publishes coverage
func (g *GeoServer) publishCoverage(workspaceName string, coverageStoreName string, publishCoverageRequest publishCoverageRequest) (published bool, err error) {

	if workspaceName != "" {
		workspaceName = fmt.Sprintf("workspaces/%s/", workspaceName)
	}
	targetURL := g.ParseURL("rest", workspaceName, "coveragestores", coverageStoreName, "/coverages")

	serializedLayer, _ := g.SerializeStruct(publishCoverageRequest)

	httpRequest := HTTPRequest{
		Method:   postMethod,
		Accept:   jsonType,
		Data:     bytes.NewBuffer(serializedLayer),
		DataType: jsonType,
		URL:      targetURL,
		Query:    nil,
	}
	response, responseCode := g.DoRequest(httpRequest)
	if responseCode != statusCreated {
		g.logger.Error(string(response))
		err = g.GetError(responseCode, response)
		return
	}

	return true, nil
}

//PublishGeoTiffLayer publishes geotiff to geoserver
func (g *GeoServer) PublishGeoTiffLayer(workspaceName string, coverageStoreName string, publishName string, fileName string) (published bool, err error) {
	//it was moved from layers.go because this is the better place for raster layers functions (coverages)
	//I tried to maintain the original behavior for backward compatibilities,
	//but it didn't seem to be working as expected from scratch
	//there were no tests for this function and I couldn't reproduce the working case
	publishRequest := publishCoverageRequest{
		&publishedCoverageDescr{
			Name:               publishName,
			NativeCoverageName: fileName,
		},
	}

	return g.publishCoverage(workspaceName, coverageStoreName, publishRequest)
}
