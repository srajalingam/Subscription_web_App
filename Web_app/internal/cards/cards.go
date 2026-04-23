package cards

import (
	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/customer"
	paymentintent "github.com/stripe/stripe-go/v85/paymentintent"
	PaymentMethod "github.com/stripe/stripe-go/v85/paymentmethod"
	"github.com/stripe/stripe-go/v85/subscription"
)

type Card struct {
	Secret   string
	Key      string
	Currency string
}

type Transaction struct {
	TransactionStatusId int
	Amount              int
	Currency            string
	LastFour            string
	BankReturnCode      string
}

func (c *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return c.CreatePaymentIntent(currency, amount)
}

func (c *Card) CreatePaymentIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	stripe.Key = c.Secret

	//creat a payment intent

	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(int64(amount)),
		Currency:    stripe.String(currency),
		Description: stripe.String("Payment for XYZ service"),
	}
	//params.AddMetadata("integration_check", "accept_a_payment")

	params.Shipping = &stripe.ShippingDetailsParams{
		Name: stripe.String("John Doe"),
		Address: &stripe.AddressParams{
			Line1:      stripe.String("Street 1"),
			City:       stripe.String("New York"),
			Country:    stripe.String("US"), // 🔥 MUST be non-India
			PostalCode: stripe.String("10001"),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}
	return pi, "", nil
}

// GetPaymentMethod gets the payment method by payment intent id

func (c *Card) GetPaymentMethod(s string) (*stripe.PaymentMethod, error) {
	stripe.Key = c.Secret

	pm, err := PaymentMethod.Get(s, nil)

	if err != nil {
		return nil, err
	}
	return pm, nil
}

// RetrievePaymentIntent gets an existing payment intent

func (c *Card) RetrievePaymentIntent(id string) (*stripe.PaymentIntent, error) {
	stripe.Key = c.Secret

	pi, err := paymentintent.Get(id, nil)

	if err != nil {
		return nil, err
	}
	return pi, nil
}

// subscribe to a plan
func (c *Card) SubscribeToPlan(cust *stripe.Customer, plan, email, last4, cardType string) (*stripe.Subscription, error) {
	stripeCustomerId := cust.ID
	items := []*stripe.SubscriptionItemsParams{
		{
			Price: stripe.String(plan),
		},
	}
	subParams := &stripe.SubscriptionParams{
		Customer: stripe.String(stripeCustomerId),
		Items:    items,
	}

	subParams.AddMetadata("email", email)
	subParams.AddMetadata("last4", last4)
	subParams.AddMetadata("cardType", cardType)

	subscription, err := subscription.New(subParams)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

// create Customer
func (c *Card) CreateCustomer(pm, email string) (*stripe.Customer, error) {
	stripe.Key = c.Secret

	customerParams := &stripe.CustomerParams{
		Email:         stripe.String(email),
		PaymentMethod: stripe.String(pm),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm),
		},
	}
	customer, err := customer.New(customerParams)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

func cardErrorMessage(code stripe.ErrorCode) string {
	var mgs = ""
	switch code {
	case stripe.ErrorCodeCardDeclined:
		mgs = "Your card was declined."
	case stripe.ErrorCodeInsufficientFunds:
		mgs = "Your card has insufficient funds."
	case stripe.ErrorCodeExpiredCard:
		mgs = "Your card has expired."
	default:
		mgs = "An error occurred while processing your card. Please try again."
	}
	return mgs
}
