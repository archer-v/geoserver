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

	return g.parseRespData(response)
}

func (g GeoServer) parseRespData(data []byte) (tasks []GwcTask, err error) {
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
