package geocoder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/j4y_funabashi/inari-micropub/pkg/app"
	log "github.com/sirupsen/logrus"
)

type Geocoder struct {
	baseURL string
	logger  *log.Logger
	apiKey  string
}

func New(apiKey, baseURL string, logger *log.Logger) Geocoder {
	return Geocoder{
		baseURL: baseURL,
		logger:  logger,
		apiKey:  apiKey,
	}
}

type geocodeResults struct {
	Response []geocodeResult `json:"results"`
}
type geocodeResult struct {
	AddressComponents []addressComponent `json:"address_components"`
	Geometry          geometry
}

func (res geocodeResult) getLocality() string {
	locality := res.GetComponent("locality")
	if locality == "" {
		locality = res.GetComponent("sublocality")
	}

	if locality == "" {
		locality = res.GetComponent("postal_town")
	}

	return locality
}

func (res geocodeResult) GetComponent(key string) string {
	for _, component := range res.AddressComponents {
		if sliceContains(component.Types, key) {
			return component.LongName
		}
	}
	return ""
}

func sliceContains(slice []string, value string) bool {
	for _, v := range slice {
		if strings.ToLower(v) == strings.ToLower(value) {
			return true
		}
	}
	return false
}

type addressComponent struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}
type geometry struct {
	Location geometryLocation `json:"location"`
}
type geometryLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (geocoder Geocoder) Lookup(address string) []app.Location {
	locList := []app.Location{}

	// build url
	apiBaseURL, err := url.Parse(geocoder.baseURL)
	if err != nil {
		geocoder.logger.WithError(err).Error("failed to parse url")
		return locList
	}
	q := apiBaseURL.Query()
	q.Add("key", geocoder.apiKey)
	q.Add("address", address)
	apiBaseURL.RawQuery = q.Encode()
	geocoder.logger.WithField("url", apiBaseURL).Info("venue search")

	// call url
	resp, err := http.Get(apiBaseURL.String())
	if err != nil {
		geocoder.logger.WithError(err).Error("failed to GET")
		return locList
	}

	// parse response
	geocodeRes := geocodeResults{}
	buf := bytes.Buffer{}
	buf.ReadFrom(resp.Body)
	err = json.Unmarshal(buf.Bytes(), &geocodeRes)
	if err != nil {
		geocoder.logger.WithError(err).Error("failed to unmarshal geocode response")
		return locList
	}

	for _, result := range geocodeRes.Response {
		locList = append(locList, app.Location{
			Lat:      result.Geometry.Location.Lat,
			Lng:      result.Geometry.Location.Lng,
			Locality: result.getLocality(),
			Region:   result.GetComponent("administrative_area_level_2"),
			Country:  result.GetComponent("country"),
		})
	}

	geocoder.logger.
		WithField("locList", locList).Info("response")
	return locList
}

func (geocoder Geocoder) LookupLatLng(lat, lng float64) []app.Location {
	locList := []app.Location{}

	// build url
	apiBaseURL, err := url.Parse(geocoder.baseURL)
	if err != nil {
		geocoder.logger.WithError(err).Error("failed to parse url")
		return locList
	}
	q := apiBaseURL.Query()
	q.Add("key", geocoder.apiKey)
	q.Add("latlng", fmt.Sprintf("%g,%g", lat, lng))
	apiBaseURL.RawQuery = q.Encode()
	geocoder.logger.WithField("url", apiBaseURL).Info("venue search")

	// call url
	resp, err := http.Get(apiBaseURL.String())
	if err != nil {
		geocoder.logger.WithError(err).Error("failed to GET")
		return locList
	}

	// parse response
	geocodeRes := geocodeResults{}
	buf := bytes.Buffer{}
	buf.ReadFrom(resp.Body)
	err = json.Unmarshal(buf.Bytes(), &geocodeRes)
	if err != nil {
		geocoder.logger.WithError(err).Error("failed to unmarshal geocode response")
		return locList
	}

	for _, result := range geocodeRes.Response {
		locList = append(locList, app.Location{
			Lat:      result.Geometry.Location.Lat,
			Lng:      result.Geometry.Location.Lng,
			Locality: result.getLocality(),
			Region:   result.GetComponent("administrative_area_level_2"),
			Country:  result.GetComponent("country"),
		})
	}

	geocoder.logger.
		WithField("locList", locList).Info("response")
	return locList
}
