package utils

import (
	"net/http"
	"text/template"
)

func RenderTemplate(w http.ResponseWriter, data interface{}, tmpl ...string) {
	template, err := template.ParseFiles(tmpl...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = template.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
