package frontend

import (
	"bytes"
	"html/template"
)

func RenderHomepage(outBuf *bytes.Buffer) error {

	t, err := template.ParseFiles(
		"view/layout.html",
		"view/homepage.html",
	)
	if err != nil {
		return err
	}
	v := struct {
		PageTitle string
	}{
		PageTitle: "jay.funabashi",
	}
	err = t.ExecuteTemplate(outBuf, "layout", v)
	return err
}
