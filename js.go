// +build js

package main

//go:generate go-bindata -o assets.go www/dist.js www/styles.css www/index.html

import (
	"net/url"

	"github.com/zserge/webview"
)

func injectHTML(html string) string {
	body := `data:text/html,` + url.PathEscape(html)
	return body
}

func loadUIFramework(w webview.WebView) {
	w.Eval(string(MustAsset("www/dist.js")))
}
