package main

import (
	"fmt"
	"net/http"
)

//CreateAndSendInvoice

func (app *application) CreateAndSendInvoice(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Create and send invoice endpoint")
}
