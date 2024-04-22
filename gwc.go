package geoserver

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
)

type GwcSeedRequest struct {
	GridsetId   string `xml:"gridSetId"`
	ZoomStart   int    `xml:"zoomStart"`
	ZoomStop    int    `xml:"zoomStop"`
	Format      string `xml:"format"`
	Type        string `xml:"type"`
	ThreadCount int    `xml:"threadCount"`
}

type GwcLayer struct {
	XMLName           xml.Name                 `xml:"GeoServerLayer"`
	ID                string                   `xml:"id"`
	Enabled           bool                     `xml:"enabled"`
	InMemoryCached    bool                     `xml:"inMemoryCached"`
	Name              string                   `xml:"name"`
	MimeFormats       []string                 `xml:"mimeFormats>string"`
	GridSubsets       []GwcLayerGridSubset     `xml:"gridSubsets>gridSubset"`
	MetaWidthHeight   []int                    `xml:"metaWidthHeight>int"`
	ExpireCache       int                      `xml:"expireCache"`
	ExpireClients     int                      `xml:"expireClients"`
	ParameterFilters  GwcLayerParameterFilters `xml:"parameterFilters"`
	Gutter            int                      `xml:"gutter"`
	CacheWarningSkips []string                 `xml:"cacheWarningSkips"`
}

type GwcLayerGridSubset struct {
	GridSetName string                    `xml:"gridSetName"`
	Extent      *GwcLayerGridSubsetExtent `xml:"extent,omitempty"`
}

type GwcLayerGridSubsetExtent struct {
	Coords []float64 `xml:"coords>double"`
}

type GwcLayerParameterFilters struct {
	StyleParameterFilter struct {
		Key          string `xml:"key"`
		DefaultValue string `xml:"defaultValue"`
	} `xml:"styleParameterFilter"`
}

type GwcTaskStatus int

const (
	GwcTaskAborted GwcTaskStatus = -1
	GwcTaskPending GwcTaskStatus = 0
	GwcTaskRunning GwcTaskStatus = 1
	GwcTaskDone    GwcTaskStatus = 2
)

type GwcTask struct {
	Id             int
	Status         GwcTaskStatus
	TilesProcessed int
	TilesTotal     int
	TilesRemaining int
}

// GwcSeedRequest performs GeoWebCache request for seed, reseed or truncate the Layer tiles cache
// returns nil on success or error
func (g *GeoServer) GwcSeedRequest(workspaceName string, layerName string, seedRequest GwcSeedRequest) (err error) {

	if seedRequest.Type != "seed" && seedRequest.Type != "truncate" && seedRequest.Type != "reseed" {
		return errors.New("'Type' field can be seed, reseed, or truncate")
	}

	if seedRequest.GridsetId == "" {
		return errors.New("'GridsetId' field should be one of available gridsets, for example 'EPSG:900913'")
	}

	if seedRequest.Format == "" {
		return errors.New("'Format' field should be one of available tile format, for example 'image/jpeg'")
	}

	if seedRequest.ThreadCount < 1 {
		seedRequest.ThreadCount = 1
	}

	targetURL := g.ParseURL("gwc", "rest", "seed", workspaceName+":"+layerName+".xml")

	sr := struct {
		XMLName xml.Name `xml:"seedRequest"`
		GwcSeedRequest
	}{GwcSeedRequest: seedRequest}

	serializedData, _ := g.SerializeToXML(sr)
	httpRequest := HTTPRequest{
		Method:   postMethod,
		Data:     bytes.NewBuffer(serializedData),
		DataType: xmlType,
		URL:      targetURL,
		Query:    nil,
	}
	response, responseCode := g.DoRequest(httpRequest)
	if responseCode != statusOk {
		g.logger.Error(string(response))
		err = g.GetError(responseCode, response)
		return
	}
	return
}

// GwcTasks returns list of GeoWebCache seeding tasks
// returns
func (g *GeoServer) GwcTasks(workspaceName string, layerName string) (tasks []GwcTask, err error) {
	targetURL := g.ParseURL("gwc", "rest", "seed", workspaceName+":"+layerName+".json")

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

	return g.parseTasksRespData(response)
}

// GetGwcLayer returns GeoWebCache layer caching configuration data
func (g GeoServer) GetGwcLayer(workspaceName string, layerName string) (layer GwcLayer, err error) {

	targetURL := g.ParseURL("gwc", "rest", "layers", workspaceName+":"+layerName)

	httpRequest := HTTPRequest{
		Method: getMethod,
		Accept: xmlType, // use xml instead json, cause json structs differ for GET and PUT request
		URL:    targetURL,
		Query:  nil,
	}

	response, responseCode := g.DoRequest(httpRequest)
	if responseCode != statusOk {
		g.logger.Error(string(response))
		err = g.GetError(responseCode, response)
		return
	}

	err = xml.Unmarshal(response, &layer)
	if err != nil {
		err = fmt.Errorf("wrong answer, error unmarshalling XML: %v\n", err)
		return
	}
	return
}

// UpdateGwcLayer create or update the layer caching configuration for GeoWebcache
func (g GeoServer) UpdateGwcLayer(layer GwcLayer) (err error) {

	targetURL := g.ParseURL("gwc", "rest", "layers", layer.Name)

	serializedData, _ := g.SerializeToXML(layer)
	httpRequest := HTTPRequest{
		Method:   putMethod,
		Data:     bytes.NewBuffer(serializedData),
		DataType: xmlType,
		URL:      targetURL,
		Query:    nil,
	}
	response, responseCode := g.DoRequest(httpRequest)
	if responseCode != statusOk {
		g.logger.Error(string(response))
		err = g.GetError(responseCode, response)
		return
	}
	return
}

func (g GeoServer) parseTasksRespData(data []byte) (tasks []GwcTask, err error) {
	var respData struct {
		Arr [][]int `json:"long-array-array"`
	}

	if err = g.DeSerializeJSON(data, &respData); err != nil {
		return nil, fmt.Errorf("can't parse the gwc response, %v", err)
	}

	for _, v := range respData.Arr {
		if len(v) != 5 {
			err = errors.New("wrong answer, array length != 5")
			return
		}
		tasks = append(tasks, GwcTask{
			Id:             v[3],
			Status:         GwcTaskStatus(v[4]),
			TilesProcessed: v[0],
			TilesTotal:     v[1],
			TilesRemaining: v[2],
		})
	}
	return
}
