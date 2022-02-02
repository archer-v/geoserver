package geoserver

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

const (
	gwcTestStoreName = "sfdem_test"
)

func gwcTestPrecondition(t *testing.T) {

	ws := testConfig.Geoserver.Workspace
	_, err := gsCatalog.CreateWorkspace(ws)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		assert.Fail(t, "can't create workspace as a precondition for FeatureTypes test")
	}

	//creating coverageStore if doesn't exist
	coverageStore := CoverageStore{
		Name:        gwcTestStoreName,
		Description: gwcTestStoreName,
		Type:        "GeoTIFF",
		URL:         "file:" + testConfig.TestData.GeoTiff,
		Workspace: &Resource{
			Name: ws,
		},
		Enabled: true,
	}
	_, err = gsCatalog.CreateCoverageStore(ws, coverageStore)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		assert.Fail(t, "can't create coverage store", err.Error())
	}

	_, err = gsCatalog.PublishCoverage(ws, gwcTestStoreName, testConfig.TestData.CoverageName, "")
	assert.Nil(t, err)

}

func gwcTestPostcondition() {
	_, _ = gsCatalog.DeleteWorkspace(testWorkspace, true)
}

func TestParseRespData(t *testing.T) {
	testData := []byte("{\"long-array-array\":[[3296,6624,124,57,2],[3296,6624,1,58,2],[-1,-1,-2,59,2],[3264,6624,180,60,2],[3328,6624,-1,61,2],[-1,-1,-2,62,2]]}")
	tasks, err := GeoServer{}.parseRespData(testData)
	if err != nil {
		t.Fatalf("parseError: %v", err.Error())
	}

	if len(tasks) != 6 {
		t.Fatalf("got array of %v instdead %v", len(tasks), 6)
	}

	task1 := GwcTask{
		Id:             57,
		Status:         GwcTaskDone,
		TilesProcessed: 3296,
		TilesTotal:     6624,
		TilesRemaining: 124,
	}

	if tasks[0] != task1 {
		t.Fatalf("wrong parsed data")
	}
}

func TestGwcTasks(t *testing.T) {

	test_before(t)

	//precondition
	gwcTestPrecondition(t)
	defer func() {
		gwcTestPostcondition()
	}()

	tasks, err := gsCatalog.GwcTasks(testConfig.Geoserver.Workspace, testConfig.TestData.CoverageName)
	assert.Nil(t, err)
	assert.True(t, len(tasks) == 0)
}

func TestGwcSeed(t *testing.T) {

	test_before(t)

	//precondition
	gwcTestPrecondition(t)
	defer func() {
		gwcTestPostcondition()
	}()

	seedRq := GwcSeedRequest{
		GridsetId: "EPSG:900913",
		Format:    "image/jpeg",
		Type:      "seed",
		ZoomStart: 0,
		ZoomStop:  10,
	}
	err := gsCatalog.GwcSeedRequest(testConfig.Geoserver.Workspace, testConfig.TestData.CoverageName, seedRq)
	assert.Nil(t, err)

	tasks, err := gsCatalog.GwcTasks(testConfig.Geoserver.Workspace, testConfig.TestData.CoverageName)
	assert.Nil(t, err)
	assert.True(t, len(tasks) != 0)
}
