package register

import (
	"fmt"
	"net/http"
)

type defaultPage struct {
}

func (lp *defaultPage) Name() string {
	return "Home"
}
func (lp *defaultPage) Handler(ctx PageContext, w http.ResponseWriter, r *http.Request) interface{} {
	fmt.Println(r.URL.Path)
	//return
	// tmpl := template.Must(template.New("default"), nil)
	// if r.Method != http.MethodPost {
	// 	tmpl.Execute(w, nil)
	// 	return
	// }

	// tmpl.Execute(w, struct{ Success bool }{true})
	return nil
}
