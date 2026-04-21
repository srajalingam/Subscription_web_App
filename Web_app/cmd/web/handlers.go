package main

import (
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

type TransactionData struct {
	FirstName       string
	LastName        string
	Email           string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

// GetRTransactionData retrieves transaction data from the request form and returns a TransactionData struct
func (app *application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var txnData TransactionData
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}
	firstName := r.Form.Get("firstName")
	lastName := r.Form.Get("lastName")
	email := r.Form.Get("cardHolderEmail")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	amount, _ := strconv.Atoi(paymentAmount)

	card := cards.Card{
		Secret: app.config.stripe.secretKey,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)

	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	txnData = TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		PaymentAmount:   amount,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  pi.ID,
	}
	return txnData, nil
}

func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//read posted data
	widgetId, _ := strconv.Atoi(r.Form.Get("product_id"))
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Printf("Payment succeeded for widget ID: %d", widgetId)

	//create a new cutomer in our database

	customerId, err := app.SaveCustormer(txnData.FirstName, txnData.LastName, txnData.Email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Printf("Customer created with ID: %d", customerId)

	//save the transaction in our database
	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
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
		Amount:        txnData.PaymentAmount,
		CreatedAt:     strconv.FormatInt(time.Now().Unix(), 10),
		UpdatedAt:     strconv.FormatInt(time.Now().Unix(), 10),
	}
	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//should write this data to session and redirect to a new page to display the data from session instead of passing it directly to the template
	app.Session.Put(r.Context(), "reciept", txnData)
	http.Redirect(w, r, "/reciept", http.StatusSeeOther)

}

func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//read posted data
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//save the transaction in our database
	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		TransactionStatusID: 2, // Assuming 2 represents a successful transaction status
	}
	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	//should write this data to session and redirect to a new page to display the data from session instead of passing it directly to the template
	app.Session.Put(r.Context(), "reciept", txnData)
	http.Redirect(w, r, "/virtual-terminal-reciept", http.StatusSeeOther)

}

func (app *application) Reciept(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "reciept").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn
	app.Session.Remove(r.Context(), "reciept")
	if err := app.renderTemplate(w, r, "reciept", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *application) VirtualTerminalReciept(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "reciept").(TransactionData)
	app.infoLog.Printf("Transaction data: %+v", txn)
	data := make(map[string]any)
	data["txn"] = txn
	app.Session.Remove(r.Context(), "reciept")
	if err := app.renderTemplate(w, r, "virtual-terminal-reciept", &templateData{Data: data}); err != nil {
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

func (app *application) BronzePlan(w http.ResponseWriter, r *http.Request) {
	widget, err := app.DB.GetWidgets(2) // Assuming 2 is the plan ID for the Bronze Plan
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Unable to get widget", http.StatusInternalServerError)
		return
	}
	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "bronze-plan", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
