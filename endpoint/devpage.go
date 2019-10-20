package endpoint

import (
	"github.com/stepan-s/ws-bro/log"
	"html/template"
	"io/ioutil"
	"net/http"
)

type devPageParams struct {
	WsURL      string
	ApiSignURL string
	ApiKey     string
}

func BindDevPage(pattern string, templatePath string, apiKey string) {
	var index, err = ioutil.ReadFile(templatePath)
	if err != nil {
		log.Error("Fail open devpage template: %v", err)
		return
	}
	var indexTemplate = template.Must(template.New("").Parse(string(index)))
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		err := indexTemplate.Execute(w, devPageParams{
			WsURL:      "wss://" + r.Host + "/bro",
			ApiSignURL: "https://" + r.Host + "/api/sign-auth",
			ApiKey:     apiKey,
		})
		if err != nil {
			log.Error("Fail execute devpage template: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
