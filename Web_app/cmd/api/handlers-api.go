package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"web_app/internal/cards"
	"web_app/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v85"
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Email         string `json:"email"`
	CradBrand     string `json:"card_brand"`
	ExpireMonth   int    `json:"exp_month"`
	ExpireYear    int    `json:"exp_year"`
	LastFour      string `json:"last_four"`
	Plan          string `json:"plan"`
	ProductID     string `json:"product_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content any    `json:"content,omitempty"`
	ID      int    `json:"id"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {

	var payload stripePayload

	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Unable to decode json", http.StatusBadRequest)
		return
	}

	fmt.Println(payload)

	amount, err := strconv.Atoi(payload.Amount)

	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Unable to convert amount to integer", http.StatusBadRequest)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secretKey,
		Key:      app.config.stripe.key,
		Currency: payload.Currency,
	}

	okay := true
	pi, msg, err := card.Charge(payload.Currency, amount)
	if okay {
		out, err := json.MarshalIndent(pi, "", "  ")
		if err != nil {
			app.errorLog.Println(err)
			http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	} else {
		j := jsonResponse{
			OK:      false,
			Message: msg,
			Content: "",
		}

		out, err := json.MarshalIndent(j, "", "  ")
		if err != nil {
			app.errorLog.Println(err)
			http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	}
}

func (app *application) GetWidgetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidgets(widgetID)

	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Unable to get widget", http.StatusInternalServerError)
		return
	}
	out, err := json.MarshalIndent(widget, "", "  ")
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (app *application) CreateCustomerAndSubscribeToPlan(w http.ResponseWriter, r *http.Request) {
	var data stripePayload
	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Unable to decode json", http.StatusBadRequest)
		return
	}
	app.infoLog.Println(data)

	card := cards.Card{
		Secret:   app.config.stripe.secretKey,
		Key:      app.config.stripe.key,
		Currency: data.Currency,
	}

	okay := true
	var subscription *stripe.Subscription
	txnMsg := "Transaction successful"

	stripeCustomer, err := card.CreateCustomer(data.PaymentMethod, data.Email)

	if err != nil {
		app.errorLog.Println(err)
		okay = false
		txnMsg = "Unable to create customer"
		return
	}
	if okay {
		subscription, err = card.SubscribeToPlan(stripeCustomer, data.Plan, data.Email, data.LastFour, "")
		if err != nil {
			app.errorLog.Println(err)
			txnMsg = "Unable to subscribe to plan"
			okay = false
		}
		app.infoLog.Printf("Subscription ID: %s", subscription.ID)
	}

	if okay {
		productId, _ := strconv.Atoi(data.ProductID)
		customerID, err := app.SaveCustormer(data.FirstName, data.LastName, data.Email)
		if err != nil {
			txnMsg = "Unable to save customer to database"
			app.errorLog.Println(err)
			return
		}

		// create a new txn
		amount, _ := strconv.Atoi(data.Amount)
		txn := models.Transaction{
			Amount:              amount,
			Currency:            "card",
			LastFour:            data.LastFour,
			ExpiryMonth:         data.ExpireMonth,
			ExpiryYear:          data.ExpireYear,
			TransactionStatusID: 2,
		}
		taxID, err := app.SaveTransaction(txn)
		if err != nil {
			txnMsg = "Unable to save transaction to database"
			app.errorLog.Println(err)
			return
		}

		//create order
		order := models.Order{
			WidgetID:      productId,
			TransactionID: taxID,
			CustomerId:    customerID,
			StatusID:      1,
			Quantity:      1,
			Amount:        amount,
			CreatedAt:     time.Now().Format(time.RFC3339),
			UpdatedAt:     time.Now().Format(time.RFC3339),
		}

		_, err = app.SaveOrder(order)
		if err != nil {
			app.errorLog.Println(err)
			txnMsg = "Unable to save order to database"
			return
		}

	}

	resp := jsonResponse{
		OK:      okay,
		Message: txnMsg,
		Content: "",
	}

	out, err := json.MarshalIndent(resp, "", "  ")

	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
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

// CreateAuthToken handles user authentication and returns a JWT token if the credentials are valid
func (app *application) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &userInput)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}

	// Get the user from the database
	user, err := app.DB.GetUserByEmail(userInput.Email)
	if err != nil {
		app.errorLog.Println(err)
		app.invalidCredentialsResponse(w, r)
		return
	}

	//validate the password
	validPassword, err := app.passwordMatches(user.Password, userInput.Password)
	if err != nil {
		app.errorLog.Println(err)
		app.invalidCredentialsResponse(w, r)
		return
	}
	if !validPassword {
		app.invalidCredentialsResponse(w, r)
		return
	}

	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = false
	payload.Message = "Authentication successful"

	_ = app.writeJSON(w, http.StatusOK, payload)
}
