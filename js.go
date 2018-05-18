// +build js

package main

//go:generate go-bindata -o assets.go www/vendor/hyperapp.min.js www/dist.js www/styles.css www/index.html

import (
	"net/url"

	"github.com/zserge/webview"
)

var uiFrameworkName = ""

func injectHTML(html string) string {
	// body := fmt.Sprintf(`document.body.innerHTML = "%s";`, url.PathEscape(html))
	body := `data:text/html,` + url.PathEscape(html)
	// fmt.Println(body)
	// w.Eval(body)
	return body
}

func loadUIFramework(w webview.WebView) {
	// injectHTML(w, string(MustAsset("js/index.html")))
	// Inject Vue.js
	w.Eval(string(MustAsset("www/vendor/hyperapp.min.js")))
	// Inject app code
	w.Eval(string(MustAsset("www/dist.js")))
}
