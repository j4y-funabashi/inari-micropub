// Package view parses app responses into template friendly data structures
package view

import "github.com/j4y_funabashi/inari-micropub/pkg/app"

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
	Media []Media
	Years []ProgressLink
}

func NewPresenter() Presenter {
	return Presenter{}
}

func (pres Presenter) ParseMediaGallery(mediaRes app.ShowMediaResponse) MediaGalleryView {

	media := []Media{}
	for _, med := range mediaRes.Media {
		m := Media{
			URL: med.URL,
		}
		media = append(media, m)
	}

	years := []ProgressLink{}
	for _, yr := range mediaRes.Years {
		y := ProgressLink{
			Name:  yr.Year,
			Value: yr.PublishedCount,
			Total: yr.Count,
			URL:   "?year=" + yr.Year,
		}
		years = append(years, y)
	}

	vm := MediaGalleryView{
		Media: media,
		Years: years,
	}
	return vm
}
