package controller

import (
	"bytes"
	"testing"
	"text/template"
)

func TestTemplate(t *testing.T) {
	tmpl, err := template.New("email_activate_tmpl_test").Parse(emailActivateTmplStr)
	if err != nil {
		t.Fatal(err)
	}
	data := &bytes.Buffer{}
	err = tmpl.Execute(data, &emailActivateTmplFiller{
		From:       "<from>",
		To:         "<to>",
		ActiveLink: "<link>",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("data is %s", data.String())
}
