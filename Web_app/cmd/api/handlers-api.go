package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"web_app/internal/cards"
	"web_app/internal/models"
	"web_app/internal/urlsigner"

	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v85"
	"golang.org/x/crypto/bcrypt"
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

	// Generate a  token for the user
	token, err := models.GenerateToken(user.ID, time.Hour*24, "authentication")
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}

	//save to database
	err = app.DB.InsertToken(token, user)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}

	//send response back to the client
	var payload struct {
		Error   bool          `json:"error"`
		Message string        `json:"message"`
		Token   *models.Token `json:"authentication_token"`
	}

	payload.Error = false
	payload.Message = fmt.Sprintf("token for %s created", userInput.Email)
	payload.Token = token

	_ = app.writeJSON(w, http.StatusOK, payload)
}

//authenticateToken

func (app *application) authenticateToken(r *http.Request) (*models.User, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return nil, errors.New("no authorization header provided")
	}
	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("invalid authorization header format")
	}
	tokenString := headerParts[1]
	if len(tokenString) != 26 {
		return nil, errors.New("invalid token length")
	}
	user, err := app.DB.GetUserForToken(tokenString)
	if err != nil {
		return nil, errors.New("invalid token")
	}
	return user, nil
}

//checkAuthentication

func (app *application) checkAuthentication(w http.ResponseWriter, r *http.Request) {
	user, err := app.authenticateToken(r)
	if err != nil {
		app.errorLog.Println(err)
		app.invalidCredentialsResponse(w, r)
		return
	}

	//valid user
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	payload.Error = false
	payload.Message = fmt.Sprintf("Authenticated user: %s", user.Email)
	_ = app.writeJSON(w, http.StatusOK, payload)
}

//VirtualTerminalPaymentSucceeded

func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	var txnData struct {
		PaymentAmount   int    `json:"amount"`
		PaymentCurrency string `json:"currency"`
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		Email           string `json:"email"`
		PaymentIntent   string `json:"payment_intent"`
		PaymentMethod   string `json:"payment_method"`
		BankReturnCode  string `json:"bank_return_code"`
		ExpiryMonth     int    `json:"exp_month"`
		ExpiryYear      int    `json:"exp_year"`
		LastFour        string `json:"last_four"`
	}
	err := app.readJSON(w, r, &txnData)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}

	card := cards.Card{
		Secret: app.config.stripe.secretKey,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(txnData.PaymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}
	pm, err := card.GetPaymentMethod(txnData.PaymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}
	txnData.LastFour = pm.Card.Last4
	txnData.ExpiryMonth = int(pm.Card.ExpMonth)
	txnData.ExpiryYear = int(pm.Card.ExpYear)

	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		PaymentIntent:       txnData.PaymentIntent,
		PaymentMethod:       txnData.PaymentMethod,
		BankReturnCode:      pi.ID,
		TransactionStatusID: 2,
	}
	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}
	app.writeJSON(w, http.StatusOK, txn)
}

// SendPasswordResetEmail
func (app *application) SendPasswordResetEmail(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}
	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}

	//check if the email exists in the database
	_, err = app.DB.GetUserByEmail(payload.Email)
	if err != nil {
		var resp struct {
			Error   bool   `json:"error"`
			Message string `json:"message"`
		}
		resp.Error = true
		resp.Message = "No user found with that email address"
		_ = app.writeJSON(w, http.StatusOK, resp)
		return
	}

	link := fmt.Sprintf("%s/reset-password?email=%s", app.config.frontend, payload.Email)

	signer := urlsigner.Signer{
		Secret: []byte(app.config.secretKey),
	}

	signedLink := signer.GenerateTokenFromString(link)

	var data struct {
		Link string
	}

	data.Link = signedLink

	//send email
	err = app.sendEmail("raja@subscription.com", payload.Email, "Password Reset", "password-reset", data)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}
	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	resp.Error = false
	resp.Message = "Password reset email sent"
	_ = app.writeJSON(w, http.StatusOK, resp)
}

// ResetPassword
func (app *application) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}
	user, err := app.DB.GetUserByEmail(payload.Email)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}
	err = app.DB.UpdateUserPassword(user, string(newHash))
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}
	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	resp.Error = false
	resp.Message = "Password reset successfully"
	_ = app.writeJSON(w, http.StatusOK, resp)

}

// AllSales
func (app *application) AllSales(w http.ResponseWriter, r *http.Request) {
	allSales, err := app.DB.GetAllOrders()
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, allSales)
}
