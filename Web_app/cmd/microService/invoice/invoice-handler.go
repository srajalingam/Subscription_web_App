package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
)

type Order struct {
	ID        int       `json:"id"`
	Quantity  int       `json:"quantity"`
	Amount    int       `json:"amount"`
	Product   string    `json:"product"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

//CreateAndSendInvoice

func (app *application) CreateAndSendInvoice(w http.ResponseWriter, r *http.Request) {
	//recive json
	var order Order
	// err := app.readJSON(w, r, &order)
	// if err != nil {
	// 	app.errorLog.Println(err)
	// 	app.badRequestResponse(w, r, err)
	// 	return
	// }

	order.ID = 123
	order.Quantity = 2
	order.Amount = 4999
	order.Product = "Widget"
	order.FirstName = "John"
	order.LastName = "Doe"
	order.Email = "admin@example.com"
	order.CreatedAt = time.Now()

	//create invoice pdf
	err := app.createInvoicePdf(order)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequestResponse(w, r, err)
		return
	}

	//create attachment
	attachments := []string{
		fmt.Sprintf("./invoices/%d.pdf", order.ID),
	}

	//send email with invoice attached

	err = app.sendEmail(
		"raja@subscription.com",
		order.Email,
		fmt.Sprintf("Invoice for Order #%d", order.ID),
		"invoice",
		attachments,
		nil,
	)

	var response struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	response.Error = false
	response.Message = fmt.Sprintf("Invoice %d.pdf created and sent to %s", order.ID, order.Email)
	err = app.writeJSON(w, http.StatusCreated, response)
}

//createInvoicePdf

func (app *application) createInvoicePdf(order Order) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 13, 10)
	pdf.SetAutoPageBreak(true, 10)

	importer := gofpdi.NewImporter()
	tpl := importer.ImportPage(pdf, "../../../pdf-templates/invoice.pdf", 1, "/MediaBox")

	pdf.AddPage()
	importer.UseImportedTemplate(pdf, tpl, 0, 0, 210, 297)

	pdf.SetY(50)
	pdf.SetX(10)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(97, 8, fmt.Sprintf("Attention: %s %s", order.FirstName, order.LastName), "", 1, "C", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(97, 8, order.CreatedAt.Format("2006-01-02"), "", 1, "C", false, 0, "")

	pdf.SetX(58)
	pdf.SetY(93)
	pdf.CellFormat(155, 8, order.Product, "", 1, "C", false, 0, "")
	pdf.SetX(166)
	pdf.CellFormat(20, 8, fmt.Sprintf("Quantity: %d", order.Quantity), "", 1, "C", false, 0, "")
	pdf.SetX(185)
	pdf.CellFormat(20, 8, fmt.Sprintf("Amount: $%.2f", float64(order.Amount)/100), "", 1, "C", false, 0, "")

	invoicePath := fmt.Sprintf("./invoices/%d.pdf", order.ID)
	err := pdf.OutputFileAndClose(invoicePath)
	if err != nil {
		return err
	}

	return nil
}
