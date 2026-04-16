package main

import (
	"fmt"
	"net/http"
	"strconv"
	"web_app/internal/cards"

	"github.com/go-chi/chi/v5"
)

// Home display
func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

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

	card := cards.Card{
		Secret: app.config.stripe.secretKey,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	data := make(map[string]interface{})
	data["cardHolder"] = cardHolder
	data["email"] = email
	data["pi"] = paymentIntent
	data["pm"] = paymentMethod
	data["pa"] = paymentAmount
	data["pc"] = paymentCurrency
	data["last_four"] = lastFour
	data["expiry_month"] = expiryMonth
	data["expiry_year"] = expiryYear
	fmt.Println(pi)
	if pi.Review != nil && pi.Review.Charge != nil {
		data["bank_return_code"] = pi.Review.Charge.ID
	} else {
		data["bank_return_code"] = ""
	}

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
	// widget := models.Widget{
	// 	ID:             1,
	// 	Name:           "Test Widget",
	// 	Description:    "This is a test widget",
	// 	InventortLevel: 10,
	// 	Price:          1999,
	// }
	id := chi.URLParam(r, "id")

	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidgets(widgetID)

	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Unable to get widget", http.StatusInternalServerError)
		return
	}
	data := make(map[string]interface{})
	data["widget"] = widget
	td.Data = data
	if err := app.renderTemplate(w, r, "buy-once", td, "stripe-js"); err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
