package register

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// PageContext provides functionality that a component or page might need to perform its functions
type PageContext interface {
	GetPageUrl(p Page) *url.URL
	// Context data refers to any additional parameters or modifiers that have changed how the page has been called (i.e: query string in URL)
	GetContextData() map[string][]string
	UnmarshallData(interface{})
}

var _ PageContext = (*pageContext)(nil)

type pageContext struct {
	request *http.Request
}

func newPageContext(r *http.Request) *pageContext {
	return &pageContext{
		request: r,
	}
}

func (pctx *pageContext) GetContextData() map[string][]string {

	return pctx.request.URL.Query()
}
func (pctx *pageContext) UnmarshallData(v interface{}) {
	err := pctx.request.ParseForm()
	if err != nil {
		logrus.Error(err)
		return
	}
	// valMap := make(map[string]interface{})
	// for k, v := range pctx.request.Form() {

	// }
	// first, we need to simplify our structure as required.

}

func (pctx *pageContext) GetPageUrl(p Page) *url.URL {
	info := globalregister.FindPage(p)
	if info == nil {
		panic("failed to resolve required page")
	}
	basePath := fmt.Sprintf("%s://%s", getSchemeFromProto(pctx.request.Proto), pctx.request.Host) + info.path

	u, err := url.Parse(basePath)
	if err != nil {
		panic(err)
	}
	return u
}

func getSchemeFromProto(proto string) string {
	parts := strings.Split(proto, "/")
	switch {
	case strings.EqualFold(parts[0], "http"):
		return "http"
	}
	return "http"
}
