// Package view parses app responses into template friendly data structures
package view

import (
	"net/url"
	"time"

	"github.com/j4y_funabashi/inari-micropub/pkg/app"
)

type Presenter struct{}

type Media struct {
	URL string
}

type ProgressLink struct {
	Name  string
	URL   string
	Value int
	Total int
}

type MediaGalleryView struct {
	Media  [][]Media
	Years  []ProgressLink
	Months []ProgressLink
	Days   []ProgressLink
}

func NewPresenter() Presenter {
	return Presenter{}
}

func (pres Presenter) ParseMediaGallery(mediaRes app.ShowMediaResponse) MediaGalleryView {
	vm := MediaGalleryView{
		Media:  parseMedia(mediaRes),
		Years:  parseYears(mediaRes),
		Months: parseMonths(mediaRes),
		Days:   parseDays(mediaRes),
	}
	return vm
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

func parseMedia(mediaRes app.ShowMediaResponse) [][]Media {

	columnCount := 3
	out := [][]Media{}

	i := 1
	column := []Media{}
	for _, m := range mediaRes.Media {
		md := Media{
			URL: m.URL,
		}
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
