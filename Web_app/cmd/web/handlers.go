package main

import (
	"net/http"
	"web_app/internal/models"
)

func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	//app.infoLog.Println("Virtual Terminal")

	// td := &templateData{
	// 	StringMap: map[string]string{
	// 		"StripePublishableKey": app.config.stripe.key,
	// 	},
	// }
	td := &templateData{}
	if err := app.renderTemplate(w, r, "terminal", td, "stripe-js"); err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//read posted data
	cardHolder := r.Form.Get("cardHolderName")
	email := r.Form.Get("cardHolderEmail")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")

	data := make(map[string]interface{})
	data["cardHolder"] = cardHolder
	data["email"] = email
	data["pi"] = paymentIntent
	data["pm"] = paymentMethod
	data["pa"] = paymentAmount
	data["pc"] = paymentCurrency

	if err := app.renderTemplate(w, r, "succeeded", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}

func (app *application) ChargeOnce(w http.ResponseWriter, r *http.Request) {
	// td := &templateData{
	// 	StringMap: map[string]string{
	// 		"StripePublishableKey": app.config.stripe.key,
	// 	},
	// }
	td := &templateData{}
	widget := models.Widget{
		ID:             1,
		Name:           "Test Widget",
		Description:    "This is a test widget",
		InventortLevel: 10,
		Price:          1999,
	}
	data := make(map[string]interface{})
	data["widget"] = widget
	td.Data = data
	if err := app.renderTemplate(w, r, "buy-once", td, "stripe-js"); err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
