package views

import (
	"html/template"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gorilla/mux"
)

type Template struct {
	once     sync.Once
	filename string
	tmpl     *template.Template
}

func (t *Template) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.tmpl = template.Must(template.ParseFiles(filepath.Join("views", "templates", t.filename)))
	})

	data := map[string]interface{}{
		"Host": r.Host,
	}

	name, ok := mux.Vars(r)["name"]
	if name != "" || ok {
		data["Name"] = name
	}

	t.tmpl.Execute(w, data)
}
