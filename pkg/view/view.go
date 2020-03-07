// Package view parses app responses into template friendly data structures
package view

import (
	"net/url"
	"strings"
	"time"

	"github.com/j4y_funabashi/inari-micropub/pkg/app"
)

var humanDateLayout = "Mon Jan 02, 2006"

type Presenter struct{}

type Location struct {
	Name     string  `json:"name"`
	Location string  `json:"location"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Locality string  `json:"locality"`
	Region   string  `json:"region"`
	Country  string  `json:"country"`
}

type LocationSearchView struct {
	Query     string
	Locations []Location
}

type Media struct {
	URL         string
	Lat         float64
	Lng         float64
	DateTime    string
	IsPublished bool
}

type ProgressLink struct {
	Name  string
	URL   string
	Value int
	Total int
}

type MediaGalleryView struct {
	Media      [][]Media
	Years      []ProgressLink
	Months     []ProgressLink
	Days       []ProgressLink
	CurrentDay ProgressLink
}

type MediaDetailView struct {
	Media Media
}

func NewPresenter() Presenter {
	return Presenter{}
}

func (pres Presenter) ParseLocationSearch(query string, locations []app.Location) LocationSearchView {
	out := LocationSearchView{
		Query: query,
	}
	for _, m := range locations {
		out.Locations = append(out.Locations, parseLocation(m))
	}
	return out
}

func parseLocation(m app.Location) Location {
	loc := []string{}
	if m.Name != "" {
		loc = append(loc, m.Name)
	}
	if m.Locality != "" {
		loc = append(loc, m.Locality)
	}
	if m.Region != "" {
		loc = append(loc, m.Region)
	}
	if m.Country != "" {
		loc = append(loc, m.Country)
	}
	md := Location{
		Name:     m.Name,
		Lat:      m.Lat,
		Lng:      m.Lng,
		Locality: m.Locality,
		Location: strings.Join(loc[:], ", "),
		Region:   m.Region,
		Country:  m.Country,
	}
	return md
}

type ComposerView struct {
	Media              []Media
	Location           Location
	SuggestedLocations []Location
	Published          string
	HumanDate          string
}

func (pres Presenter) ParseComposer(sess app.SessionData) ComposerView {
	out := ComposerView{}
	for _, m := range sess.Media {
		out.Media = append(out.Media, parseMedia(m))
	}
	out.Location = parseLocation(sess.Location)
	if sess.Published != nil {
		out.Published = sess.Published.Format(time.RFC3339)
		out.HumanDate = sess.Published.Format(humanDateLayout)
	}
	for _, loc := range sess.SuggestedLocations {
		out.SuggestedLocations = append(out.SuggestedLocations, parseLocation(loc))
	}
	return out
}

func (pres Presenter) ParseMediaDetail(mediaRes app.ShowMediaDetailResponse) MediaDetailView {
	return MediaDetailView{
		Media: parseMedia(mediaRes.Media),
	}
}

func (pres Presenter) ParseMediaGallery(mediaRes app.ShowMediaResponse) MediaGalleryView {
	vm := MediaGalleryView{
		Media:      parseMediaColumns(mediaRes),
		Years:      parseYears(mediaRes),
		Months:     parseMonths(mediaRes),
		Days:       parseDays(mediaRes),
		CurrentDay: parseCurrentDay(mediaRes),
	}
	return vm
}

func parseCurrentDay(mediaRes app.ShowMediaResponse) ProgressLink {

	now, _ := time.Parse(
		"2006-1-2",
		mediaRes.CurrentYear.Year+"-"+mediaRes.CurrentMonth.Month+"-"+mediaRes.CurrentDay.Day,
	)
	day := ProgressLink{
		Name:  now.Format("Monday, 2 Jan 2006"),
		Value: mediaRes.CurrentDay.PublishedCount,
		Total: mediaRes.CurrentDay.Count,
	}

	return day
}

func parseYears(mediaRes app.ShowMediaResponse) []ProgressLink {
	years := []ProgressLink{}
	for _, yr := range mediaRes.Years {
		urlParams := url.Values{}
		urlParams.Add("year", yr.Year)
		y := ProgressLink{
			Name:  yr.Year,
			Value: yr.PublishedCount,
			Total: yr.Count,
			URL:   "?" + urlParams.Encode(),
		}
		years = append(years, y)
	}

	return years
}

func parseDays(mediaRes app.ShowMediaResponse) []ProgressLink {
	days := []ProgressLink{}
	for _, item := range mediaRes.Days {
		now, _ := time.Parse("2006-1-2", mediaRes.CurrentYear.Year+"-"+mediaRes.CurrentMonth.Month+"-"+item.Day)
		urlParams := url.Values{}
		urlParams.Add("month", mediaRes.CurrentMonth.Month)
		urlParams.Add("year", mediaRes.CurrentYear.Year)
		urlParams.Add("day", item.Day)
		y := ProgressLink{
			Name:  now.Format("Mon 2"),
			Value: item.PublishedCount,
			Total: item.Count,
			URL:   "?" + urlParams.Encode(),
		}
		days = append(days, y)
	}

	return days
}

func parseMonths(mediaRes app.ShowMediaResponse) []ProgressLink {
	months := []ProgressLink{}
	for _, item := range mediaRes.Months {
		now, _ := time.Parse("1", item.Month)
		urlParams := url.Values{}
		urlParams.Add("month", item.Month)
		urlParams.Add("year", mediaRes.CurrentYear.Year)
		y := ProgressLink{
			Name:  now.Format("January"),
			Value: item.PublishedCount,
			Total: item.Count,
			URL:   "?" + urlParams.Encode(),
		}
		months = append(months, y)
	}

	return months
}

func parseMedia(m app.Media) Media {
	md := Media{
		URL:         m.URL,
		Lat:         m.Lat,
		Lng:         m.Lng,
		DateTime:    m.DateTime.Format(humanDateLayout),
		IsPublished: m.IsPublished,
	}
	return md
}

func parseMediaColumns(mediaRes app.ShowMediaResponse) [][]Media {

	columnCount := 4
	out := [][]Media{}

	i := 1
	column := []Media{}
	for _, m := range mediaRes.Media {
		md := parseMedia(m)
		column = append(column, md)
		if i%columnCount == 0 {
			out = append(out, column)
			column = []Media{}
		}
		i++
	}
	if len(column) > 0 {
		out = append(out, column)
	}

	return out
}
