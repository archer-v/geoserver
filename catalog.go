package geoserver

import (
	"io"
	"net/http"
	"os"
	"time"
)

// Catalog is geoserver interface that define all operations
type Catalog interface {
	WorkspaceService
	DatastoreService
	StyleService
	AboutService
	LayerService
	LayerGroupService
	CoverageStoresService
	FeatureTypeService
	UtilsInterface
}

// GetCatalog return geoserver catalog instance,
// this fuction take geoserverURL('http://localhost:8080/geoserver/') ,
// geoserver username,
// geoserver password
// return geoserver structObj
func GetCatalog(geoserverURL string, username string, password string) (catalog *GeoServer) {
	geoserver := GeoServer{
		ServerURL: geoserverURL,
		Username:  username,
		Password:  password,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				DisableCompression:    true, // gzip compression is disabled, cause GWC has an issue with erroneous responses
				ResponseHeaderTimeout: time.Second * 5,
			},
		},
		logger: GetLogger(),
	}

	if LogFile != nil {
		if LogConsoleQuiet {
			geoserver.logger.Out = LogFile
		} else {
			geoserver.logger.Out = io.MultiWriter(LogFile, os.Stdout)
		}
	}
	return &geoserver
}
