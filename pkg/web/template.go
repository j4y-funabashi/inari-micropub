package web

import (
	"bytes"
	"net/http"
	"text/template"

	"github.com/j4y_funabashi/inari-micropub/pkg/view"
)

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
