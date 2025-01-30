package handlers

import "html/template"

var tmpl = template.Must(template.ParseGlob("templates/*.html"))