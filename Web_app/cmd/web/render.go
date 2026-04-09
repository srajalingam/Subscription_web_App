package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type templateData struct {
	StringMap            map[string]string
	IntMap               map[string]int
	FloatMap             map[string]float64
	Data                 any
	CSRFToken            string
	Flash                string
	IsAuthenticated      bool
	Warning              string
	Error                string
	API                  string
	CSSVersion           string
	StripeSecretKey      string
	StripePublishableKey string
}

var functions = template.FuncMap{
	"formatCurrency": formatCurrency,
}

func formatCurrency(amount int) string {
	f := float64(amount) / 100
	return fmt.Sprintf("%.2f", f)
}

//go:embed templates
var templateFS embed.FS

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	td.API = app.config.api
	td.StripeSecretKey = app.config.stripe.secretKey
	td.StripePublishableKey = app.config.stripe.key
	return td
}

func (app *application) renderTemplate(w http.ResponseWriter, r *http.Request, page string, td *templateData, partials ...string) error {
	var t *template.Template
	var err error

	templateToRender := fmt.Sprintf("templates/%s.page.gohtml", page)

	_, templateInMap := app.templateCache[templateToRender]

	if app.config.env == "production" && templateInMap {
		t = app.templateCache[templateToRender]
	} else {
		t, err = app.parseTemplate(partials, page, templateToRender)

		if err != nil {
			app.errorLog.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return err
		}
	}

	if td == nil {
		td = &templateData{}
	}

	td = app.addDefaultData(td, r)

	err = t.Execute(w, td)

	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}

	return nil

}

func (app *application) parseTemplate(partials []string, page string, templateToRender string) (*template.Template, error) {
	var t *template.Template
	var err error

	//build partials slice
	if len(partials) > 0 {
		for i, partial := range partials {
			partials[i] = fmt.Sprintf("templates/%s.partial.gohtml", partial)
		}
	}

	if len(partials) > 0 {
		t, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).ParseFS(templateFS, templateToRender, "templates/base.layout.gohtml", strings.Join(partials, ","), templateToRender)
	} else {
		t, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).ParseFS(templateFS, templateToRender, "templates/base.layout.gohtml", templateToRender)
	}

	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}
	app.templateCache[templateToRender] = t
	return t, err
}
