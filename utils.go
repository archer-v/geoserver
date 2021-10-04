package geoserver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"reflect"
)

//HTTPRequest is an http request object
type HTTPRequest struct {
	URL      string
	Accept   string
	Query    map[string]string
	Data     io.Reader
	DataType string
	Method   string
}

type Error struct {
	Message string `json:"message"`
	Url     string `json:"url"`
	Status  int    `json:"status,string"`
	Servlet string `json:"servlet"`
}

func (e *Error) Error() string {
	geoserverErr, ok := statusErrorMapping[e.Status]
	if !ok {
		geoserverErr = fmt.Errorf("Unexpected Error with status code %d", e.Status)
	}
	return fmt.Sprintf("abstract:%s\ndetails:%s\n", geoserverErr, e.Message)
}

//create newError with detail fields
func newError(statusCode int, text []byte) *Error {
	err := Error{
		Status:  statusCode,
		Message: string(text),
	}
	//some errors are returned as JSON
	//for example when assembled REST api uri is incorrect due to empty the resource name or so on
	_ = json.Unmarshal(text, &err)
	return &err
}

//UtilsInterface contians common function used to help you deal with data and geoserver api
type UtilsInterface interface {
	DoRequest(request HTTPRequest) (responseText []byte, statusCode int)
	SerializeStruct(structObj interface{}) ([]byte, error)
	DeSerializeJSON(response []byte, structObj interface{}) (err error)
	ParseURL(urlParts ...string) (parsedURL string)
}

//DoRequest Send request and return result and statusCode
func (g *GeoServer) DoRequest(request HTTPRequest) (responseText []byte, statusCode int) {
	defer func() {
		if r := recover(); r != nil {
			responseText = []byte(fmt.Sprintf("%s", r))
			statusCode = 0
		}
	}()
	var req *http.Request
	switch request.Method {
	case getMethod, deleteMethod:
		req = g.GetGeoserverRequest(request.URL, request.Method, request.Accept, nil, "")
	case postMethod, putMethod:
		req = g.GetGeoserverRequest(request.URL, request.Method, request.Accept, request.Data, request.DataType)
	default:
		panic("unrecognized http request Method")
	}
	if len(request.Query) != 0 {
		q := req.URL.Query()
		for k, v := range request.Query {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	response, responseErr := g.HttpClient.Do(req)
	if responseErr != nil {
		panic(responseErr)
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	g.logger.Infof("url:%s  Status=%s", req.URL, response.Status)
	return body, response.StatusCode
}

//GetError this return the proper error message
func (g *GeoServer) GetError(statusCode int, text []byte) (err error) {
	return newError(statusCode, text)
}

// IsEmpty helper function to check if obj/struct is nil/empty
func IsEmpty(object interface{}) bool {
	if object == nil {
		return true
	} else if object == "" {
		return true
	} else if object == false {
		return true
	}
	if reflect.ValueOf(object).Kind() == reflect.Struct {
		empty := reflect.New(reflect.TypeOf(object)).Elem().Interface()
		if reflect.DeepEqual(object, empty) {
			return true
		}
	}
	return false
}

//SerializeStruct convert struct to json
func (g *GeoServer) SerializeStruct(structObj interface{}) ([]byte, error) {
	serializedStruct, err := json.Marshal(&structObj)
	if err != nil {
		g.logger.Error(err)
		return nil, err
	}
	return serializedStruct, nil
}

//DeSerializeJSON json struct to struct
func (g *GeoServer) DeSerializeJSON(response []byte, structObj interface{}) (err error) {
	err = json.Unmarshal(response, &structObj)
	if err != nil {
		g.logger.Error(err)
		return err
	}
	return nil
}
func (g *GeoServer) getGoGeoserverPackageDir() string {
	dir, err := filepath.Abs("./")
	if err != nil {
		panic(err)
	}
	return dir
}

//ParseURL this function join urlParts with geoserver url
func (g *GeoServer) ParseURL(urlParts ...string) (parsedURL string) {
	defer func() {
		if r := recover(); r != nil {
			parsedURL = ""
		}
	}()
	geoserverURL, err := url.Parse(g.ServerURL)
	if err != nil {
		g.logger.Error(err)
		panic(err)
	}
	urlArr := append([]string{geoserverURL.Path}, urlParts...)
	geoserverURL.Path = path.Join(urlArr...)
	parsedURL = geoserverURL.String()
	return

}
