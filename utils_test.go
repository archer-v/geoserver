package geoserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializeStruct(t *testing.T) {
	gsCatalog := GetCatalog("http://localhost:8080/geoserver/", "admin", "geoserver")
	resource := Resource{Class: "Test", Href: "http://localhost:8080/geoserver/", Name: "Test1"}
	json, err := gsCatalog.SerializeStruct(&resource)
	assert.NotEmpty(t, json)
	assert.Nil(t, err)
}

func TestDeSerializeJSON(t *testing.T) {
	gsCatalog := GetCatalog("http://localhost:8080/geoserver/", "admin", "geoserver")
	json := []byte(`{"@class":"Test","name":"Test1","href":"http://localhost:8080/geoserver/"}`)
	resource := Resource{}
	err := gsCatalog.DeSerializeJSON(json, &resource)
	assert.NotNil(t, resource)
	assert.NotEmpty(t, resource)
	assert.Nil(t, err)
}