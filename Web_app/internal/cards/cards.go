package cards

import (
	"github.com/stripe/stripe-go/v85"
	paymentintent "github.com/stripe/stripe-go/v85/paymentintent"
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
