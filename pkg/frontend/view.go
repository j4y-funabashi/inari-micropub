package frontend

import (
	"bytes"
	"html/template"

	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
)

func RenderHomepage(outBuf *bytes.Buffer, postList []mf2.MicroFormatView) error {

	t, err := template.ParseFiles(
		"view/layout.html",
		"view/homepage.html",
	)
	if err != nil {
		return err
	}
	v := struct {
		PageTitle string
		PostList  []mf2.MicroFormatView
	}{
		PageTitle: "jay.funabashi",
		PostList:  postList,
	}
	err = t.ExecuteTemplate(outBuf, "layout", v)
	return err
}
