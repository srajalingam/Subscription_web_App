package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"web_app/internal/cards"
	"web_app/internal/models"

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
	firstName := r.Form.Get("firstName")
	lastName := r.Form.Get("lastName")
	email := r.Form.Get("cardHolderEmail")
	cardHolder := r.Form.Get("cardHolderName")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	widgetId, _ := strconv.Atoi(r.Form.Get("product_id"))

	app.infoLog.Printf("Payment succeeded for widget ID: %d", widgetId)

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

	//create a new cutomer in our database

	customerId, err := app.SaveCustormer(firstName, lastName, email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Printf("Customer created with ID: %d", customerId)

	//save the transaction in our database
	amount, _ := strconv.Atoi(paymentAmount)
	txn := models.Transaction{
		Amount:              amount,
		Currency:            paymentCurrency,
		LastFour:            lastFour,
		ExpiryMonth:         int(expiryMonth),
		ExpiryYear:          int(expiryYear),
		BankReturnCode:      pi.ID,
		TransactionStatusID: 2, // Assuming 2 represents a successful transaction status
	}
	txnId, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//create a new order in our database
	order := models.Order{
		WidgetID:      widgetId,
		TransactionID: txnId,
		CustomerId:    customerId,
		StatusID:      1, // Assuming 1 represents a new order status
		Quantity:      1,
		Amount:        amount,
		CreatedAt:     strconv.FormatInt(time.Now().Unix(), 10),
		UpdatedAt:     strconv.FormatInt(time.Now().Unix(), 10),
	}
	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

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
	data["bank_return_code"] = pi.ID
	data["first_name"] = firstName
	data["last_name"] = lastName

	if err := app.renderTemplate(w, r, "succeeded", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}

// SaveTransaction saves a new transaction to the database and returns the transaction ID
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {
	transactionId, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}
	return transactionId, nil
}

// SaveOrder saves a new order to the database and returns the order ID
func (app *application) SaveOrder(order models.Order) (int, error) {
	orderId, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}
	return orderId, nil
}

// SaveCustomer saves a new customer to the database and returns the customer ID
func (app *application) SaveCustormer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}
	customerId, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}
	return customerId, nil
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
