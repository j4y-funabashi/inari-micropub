package frontend

import (
	"bytes"
	"html/template"

	"github.com/j4y_funabashi/inari-micropub/pkg/mf2"
)

func RenderHomepage(outBuf *bytes.Buffer, postList []mf2.MicroFormatView, afterKey string) error {

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
		AfterKey  string
	}{
		PageTitle: "jay.funabashi",
		PostList:  postList,
		AfterKey:  afterKey,
	}
	err = t.ExecuteTemplate(outBuf, "layout", v)
	return err
}
