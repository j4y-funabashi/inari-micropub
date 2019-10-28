package web

import (
	"bytes"
	"net/http"
	"text/template"

	"github.com/j4y_funabashi/inari-micropub/pkg/view"
)

func renderMediaDetail(media view.MediaDetailView, w http.ResponseWriter) error {
	outBuf := new(bytes.Buffer)
	t, err := template.ParseFiles(
		"view/layout.html",
		"view/media_detail.html",
	)
	if err != nil {
		return err
	}
	v := struct {
		PageTitle string
		Model     view.MediaDetailView
	}{
		PageTitle: "Add media to Post",
		Model:     media,
	}
	err = t.ExecuteTemplate(outBuf, "layout", v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(outBuf.Bytes())
	return err
}

func renderComposerForm(viewModel view.ComposerView, w http.ResponseWriter) error {
	outBuf := new(bytes.Buffer)
	t, err := template.ParseFiles(
		"view/layout.html",
		"view/composer.html",
	)
	if err != nil {
		return err
	}
	v := struct {
		PageTitle string
		Model     view.ComposerView
	}{
		PageTitle: "Add a Post",
		Model:     viewModel,
	}
	err = t.ExecuteTemplate(outBuf, "layout", v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(outBuf.Bytes())
	return err
}

func renderLocationSearch(viewModel view.LocationSearchView, w http.ResponseWriter) error {
	outBuf := new(bytes.Buffer)
	t, err := template.ParseFiles(
		"view/layout.html",
		"view/location-search.html",
	)
	if err != nil {
		return err
	}
	v := struct {
		PageTitle string
		Model     view.LocationSearchView
	}{
		PageTitle: "Search location",
		Model:     viewModel,
	}
	err = t.ExecuteTemplate(outBuf, "layout", v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(outBuf.Bytes())
	return err
}

func renderLoginForm(w http.ResponseWriter) error {
	outBuf := new(bytes.Buffer)
	t, err := template.ParseFiles(
		"view/layout.html",
		"view/login.html",
	)
	if err != nil {
		return err
	}
	v := struct {
		PageTitle string
	}{
		PageTitle: "login",
	}
	err = t.ExecuteTemplate(outBuf, "layout", v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(outBuf.Bytes())
	return err
}

func renderMediaGallery(viewModel view.MediaGalleryView, w http.ResponseWriter) error {
	outBuf := new(bytes.Buffer)
	t, err := template.ParseFiles(
		"view/layout.html",
		"view/media_gallery.html",
	)
	if err != nil {
		return err
	}
	v := struct {
		PageTitle string
		Model     view.MediaGalleryView
	}{
		PageTitle: "jay.funabashi",
		Model:     viewModel,
	}
	err = t.ExecuteTemplate(outBuf, "layout", v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(outBuf.Bytes())
	return err
}
