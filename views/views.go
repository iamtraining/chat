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
	Filename string
	tmpl     *template.Template
}

func (t *Template) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.tmpl = template.Must(template.ParseFiles(filepath.Join("views", "templates", t.Filename)))
	})

	data := map[string]interface{}{
		"Host": r.Host,
	}

	room, ok := mux.Vars(r)["room"]
	if room != "" || ok {
		data["Room"] = room
	}
	//fmt.Println(room)
	//fmt.Println(data)

	t.tmpl.Execute(w, data)
}
