package geoserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// CRSType geoserver crs response
type CRSType struct {
	Class string `json:"@class,omitempty"`
	Value string `json:"$,omitempty"`
}

// UnmarshalJSON custom deserialization to handle published layers of group
func (u *CRSType) UnmarshalJSON(data []byte) error {
	var raw interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	switch raw := raw.(type) {
	case map[string]interface{}:
		*u = CRSType{Class: raw["@class"].(string), Value: raw["$"].(string)}
	case interface{}:
		*u = CRSType{Class: "string", Value: string(data)}
	}
	return nil
}

// MarshalJSON custom crs serialization
func (u *CRSType) MarshalJSON() ([]byte, error) {
	if IsEmpty(u) {
		x := ""
		return json.Marshal(&x)
	} else if !IsEmpty(u.Class) && u.Class == "string" {
		return json.Marshal(u.Value)
	}
	type crsType struct {
		Class string `json:"@class,omitempty"`
		Value string `json:"$,omitempty"`
	}
	return json.Marshal(&crsType{
		Class: u.Class,
		Value: u.Value,
	})
}

// FeatureTypeService define all geoserver featuretype operations
type FeatureTypeService interface {
	GetFeatureTypes(workspaceName string, datastoreName string) (featureTypes []*Resource, err error)
	GetFeatureType(workspaceName string, datastoreName string, featureTypeName string) (featureType *FeatureType, err error)
	DeleteFeatureType(workspaceName string, datastoreName string, featureTypeName string, recurse bool) (deleted bool, err error)
}

// Entry is geoserver Entry
type Entry struct {
	Key   string `json:"@key,omitempty"`
	Value string `json:"$,omitempty"`
}

// BoundingBox is geoserver Bounding Box for FeatureType
type BoundingBox struct {
	Minx float64 `json:"minx,omitempty"`
	Maxx float64 `json:"maxx,omitempty"`
	Miny float64 `json:"miny,omitempty"`
	Maxy float64 `json:"maxy,omitempty"`
}

// Metadata is the geoserver Metadata
type Metadata struct {
	Entry []*Entry `json:"entry,omitempty"`
}

// Keywords is the geoserver Keywords
type Keywords struct {
	String []string `json:"string,omitempty"`
}

// ResponseSRS is the geoserver ResponseSRS
type ResponseSRS struct {
	String []int `json:"string,omitempty"`
}

// NativeBoundingBox is geoserver NativeBoundingBox for FeatureType
type NativeBoundingBox struct {
	BoundingBox
	Crs *CRSType `json:"crs,omitempty"`
}

// LatLonBoundingBox is geoserver LatLonBoundingBox for FeatureType
type LatLonBoundingBox struct {
	BoundingBox
	Crs *CRSType `json:"crs,omitempty"`
}

// MetadataLink is geoserver metadata link
type MetadataLink struct {
	Type         string `json:"type,omitempty"`
	MetadataType string `json:"metadataType,omitempty"`
	Content      string `json:"content,omitempty"`
}

// MetadataLinks is the geoserver metadata links
type MetadataLinks struct {
	MetadataLink []*MetadataLink `json:"metadataLink,omitempty"`
}

// DataLinks is the geoserver FeatureType Datalinks
type DataLinks struct {
	DataLink []*MetadataLink `json:"org.geoserver.catalog.impl.DataLinkInfoImpl,omitempty"`
}

// Attributes is the geoserver feature type attributes
type Attributes struct {
	Attribute []*Attribute `json:"attribute,omitempty"`
}

// Attribute is geoserver FeatureType Attribute
type Attribute struct {
	Name      string `json:"name,omitempty"`
	MinOccurs int16  `json:"minOccurs,omitempty"`
	MaxOccurs int16  `json:"maxOccurs,omitempty"`
	Nillable  bool   `json:"nillable,omitempty"`
	Binding   string `json:"binding,omitempty"`
	Length    int16  `json:"length,omitempty"`
}

// FeatureType is geoserver FeatureType
type FeatureType struct {
	Name                   string             `json:"name,omitempty"`
	NativeName             string             `json:"nativeName,omitempty"`
	Namespace              *Resource          `json:"namespace,omitempty"`
	Title                  string             `json:"title,omitempty"`
	Abstract               string             `json:"abstract,omitempty"`
	Keywords               *Keywords          `json:"keywords,omitempty"`
	Metadatalinks          *MetadataLinks     `json:"metadatalinks,omitempty"`
	DataLinks              *DataLinks         `json:"dataLinks,omitempty"`
	NativeCRS              *CRSType           `json:"nativeCRS,omitempty"`
	Srs                    string             `json:"srs,omitempty"`
	Enabled                bool               `json:"enabled,omitempty"`
	NativeBoundingBox      *NativeBoundingBox `json:"nativeBoundingBox,omitempty"`
	LatLonBoundingBox      *LatLonBoundingBox `json:"latLonBoundingBox,omitempty"`
	ProjectionPolicy       string             `json:"projectionPolicy,omitempty"`
	Metadata               *Metadata          `json:"metadata,omitempty"`
	Store                  *Resource          `json:"store,omitempty"`
	CqlFilter              string             `json:"cqlFilter,omitempty"`
	MaxFeatures            int32              `json:"maxFeatures,omitempty"`
	NumDecimals            float32            `json:"numDecimals,omitempty"`
	ResponseSRS            *ResponseSRS       `json:"responseSRS,omitempty"`
	CircularArcPresent     bool               `json:"circularArcPresent,omitempty"`
	OverridingServiceSRS   bool               `json:"overridingServiceSRS,omitempty"`
	SkipNumberMatched      bool               `json:"skipNumberMatched,omitempty"`
	LinearizationTolerance bool               `json:"linearizationTolerance,omitempty"`
	Attributes             *Attributes        `json:"attributes,omitempty"`
}

// FeatureTypes holds a list of geoserver styles
type FeatureTypes struct {
	FeatureType []*Resource `json:"featureType,omitempty"`
}

// FeatureTypesResponseBody is the api body
type FeatureTypesResponseBody struct {
	FeatureTypes *FeatureTypes `json:"featureTypes,omitempty"`
}

// FeatureTypesRequestBody is the api body
type FeatureTypesRequestBody struct {
	FeatureType *FeatureType `json:"featureTypes,omitempty"`
}

// GetFeatureTypes return all featureTypes in workspace and datastore if error occurred err will be return and nil for featrueTypes
func (g *GeoServer) GetFeatureTypes(workspaceName string, datastoreName string) (featureTypes []*Resource, err error) {
	if workspaceName != "" {
		workspaceName = fmt.Sprintf("workspaces/%s/", workspaceName)
	}
	if datastoreName != "" {
		datastoreName = fmt.Sprintf("datastores/%s/featuretypes", datastoreName)
	}
	targetURL := g.ParseURL("rest", workspaceName, datastoreName)
	httpRequest := HTTPRequest{
		Method: getMethod,
		Accept: jsonType,
		URL:    targetURL,
		Query:  nil,
	}
	response, responseCode := g.DoRequest(httpRequest)
	if responseCode != statusOk {
		featureTypes = nil
		err = g.GetError(responseCode, response)
		return
	}
	featureTypesResponse := &FeatureTypesResponseBody{FeatureTypes: &FeatureTypes{FeatureType: make([]*Resource, 0, 0)}}
	g.DeSerializeJSON(response, featureTypesResponse)
	featureTypes = featureTypesResponse.FeatureTypes.FeatureType
	return
}

// CreateFeatureType creates featureType in workspace and datastore
// Creating featureType is only allowed to database related datastores (like postgresql, etc)
// if error occurred err will be return and nil for featrueTypes
func (g *GeoServer) CreateFeatureType(workspaceName string, datastoreName string, featureType *FeatureType) (created bool, err error) {
	targetURL := g.ParseURL("rest", "workspaces", workspaceName, "datastores", datastoreName, "featuretypes")

	createFeatureTypeRequest := struct {
		FeatureType *FeatureType `json:"featureType"`
	}{featureType}

	return g.createEntity(targetURL, createFeatureTypeRequest, nil)
}

// DeleteFeatureType Delete FeatureType from geoserver given that workspaceName, datastoreName, featureTypeName
// if featuretype deleted successfully will return true and nil for err
// if error occurred will return false and error for err
func (g *GeoServer) DeleteFeatureType(workspaceName string, datastoreName string, featureTypeName string, recurse bool) (deleted bool, err error) {
	if workspaceName != "" {
		workspaceName = fmt.Sprintf("workspaces/%s/", workspaceName)
	}
	if datastoreName != "" {
		datastoreName = fmt.Sprintf("datastores/%s/", datastoreName)
	}
	targetURL := g.ParseURL("rest", workspaceName, datastoreName, "featuretypes", featureTypeName)
	httpRequest := HTTPRequest{
		Method: deleteMethod,
		Accept: jsonType,
		URL:    targetURL,
		Query:  map[string]string{"recurse": strconv.FormatBool(recurse)},
	}
	response, responseCode := g.DoRequest(httpRequest)
	if responseCode != statusOk {
		deleted = false
		err = g.GetError(responseCode, response)
		return
	}
	deleted = true
	return
}

// GetFeatureType it return geoserver FeatureType and nil err
// if success else nil for fetureType error for err
func (g *GeoServer) GetFeatureType(workspaceName string, datastoreName string, featureTypeName string) (featureType *FeatureType, err error) {
	if workspaceName != "" {
		workspaceName = fmt.Sprintf("workspaces/%s/", workspaceName)
	}
	if datastoreName != "" {
		datastoreName = fmt.Sprintf("datastores/%s/featuretypes", datastoreName)
	}
	targetURL := g.ParseURL("rest", workspaceName, datastoreName, featureTypeName)
	httpRequest := HTTPRequest{
		Method: getMethod,
		Accept: jsonType,
		URL:    targetURL,
		Query:  nil,
	}
	response, responseCode := g.DoRequest(httpRequest)
	if responseCode != statusOk {
		featureType = nil
		err = g.GetError(responseCode, response)
		return
	}
	var featureTypeResponse struct {
		FeatureType *FeatureType `json:"featureType,omitempty"`
	}
	g.DeSerializeJSON(response, &featureTypeResponse)
	featureType = featureTypeResponse.FeatureType
	return
}

// UpdateFeatureType updates geoserver featureType resource, else returns error,
// featureTypeName is a featureType name or empty (featureType.Name value will be used)
// recalculate can be nil, or an array of recalculate options: "nativebbox", "latlonbbox"
// an empty recalculate array cause to avoid all recalculation on large dataset
// see https://docs.geoserver.org/latest/en/api/#1.0.0/featuretypes.yaml
func (g *GeoServer) UpdateFeatureType(workspaceName string, featureType *FeatureType, featureTypeName string, recalculate []string) (modified bool, err error) {

	items := strings.Split(featureType.Store.Name, ":")
	if len(items) != 2 {
		return false, errors.New("internal error during featureType update, can't build store name")
	}

	if featureTypeName == "" {
		featureTypeName = featureType.Name
	}
	targetURL := g.ParseURL("rest", "workspaces", workspaceName, "datastores", items[1], "featuretypes", featureTypeName)

	var query map[string]string

	if recalculate != nil {
		query = make(map[string]string)
		query["recalculate"] = strings.Join(recalculate, ",")
	}

	type FeatureTypeUpdate struct {
		Name              string             `json:"name,omitempty"`
		Title             string             `json:"title,omitempty"`
		Description       string             `json:"description,omitempty"`
		Abstract          string             `json:"abstract,omitempty"`
		Keywords          *Keywords          `json:"keywords,omitempty"`
		Enabled           bool               `json:"enabled,omitempty"`
		Srs               string             `json:"srs,omitempty"`
		NativeBoundingBox *NativeBoundingBox `json:"nativeBoundingBox,omitempty"`
		LatLonBoundingBox *LatLonBoundingBox `json:"latLonBoundingBox,omitempty"`
	}

	type featureTypeUpdateRequestBody struct {
		FeatureType FeatureTypeUpdate `json:"featureType,omitempty"`
	}

	data := featureTypeUpdateRequestBody{FeatureType: FeatureTypeUpdate{
		Name:              featureType.Name,
		Title:             featureType.Title,
		Abstract:          featureType.Abstract,
		Keywords:          featureType.Keywords,
		Enabled:           featureType.Enabled,
		NativeBoundingBox: featureType.NativeBoundingBox,
		LatLonBoundingBox: featureType.LatLonBoundingBox,
		Srs:               featureType.Srs,
	}}

	serializedLayer, _ := g.SerializeStruct(data)
	httpRequest := HTTPRequest{
		Method:   putMethod,
		Accept:   jsonType,
		Data:     bytes.NewBuffer(serializedLayer),
		DataType: jsonType,
		URL:      targetURL,
		Query:    query,
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
