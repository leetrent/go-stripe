package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/leetrent/go-stripe/internal/cards"
	"github.com/leetrent/go-stripe/internal/models"
	"github.com/stripe/stripe-go/v72"
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Email         string `json:"email"`
	CardBrand     string `json:"card_brand"`
	ExpiryMonth   int    `json:"exp_month"`
	ExpiryYear    int    `json:"exp_year"`
	LastFour      string `json:"last_four"`
	Plan          string `json:"plan"`
	ProductID     string `json:"product_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	ID      int    `json:"id,omitempty"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var payload stripePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	amount, err := strconv.Atoi(payload.Amount)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: payload.Currency,
	}

	okay := true

	pi, msg, err := card.Charge(payload.Currency, amount)
	if err != nil {
		okay = false
		fmt.Println("Charge card error: ", err)
		fmt.Println("Charge card message: ", msg)
	}

	if okay {
		out, err := json.MarshalIndent(pi, "", "   ")
		if err != nil {
			app.errorLog.Println(err)
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

		out, err := json.MarshalIndent(j, "", "   ")
		if err != nil {
			app.errorLog.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	}
}

func (app *application) GetWidgetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	out, err := json.MarshalIndent(widget, "", "   ")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (app *application) CreateCustomerAndSubscribeToPlan(w http.ResponseWriter, r *http.Request) {
	var data stripePayload
	var subscription *stripe.Subscription
	okay := true
	txnMsg := "Transaction was successful!"

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//app.infoLog.Println(data.Email, data.LastFour, data.PaymentMethod, data.Plan)

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: data.Currency,
	}

	stripeCustomer, msg, err := card.CreateCustomer(data.PaymentMethod, data.Email)
	if err != nil {
		app.errorLog.Println(err)
		txnMsg = msg
		okay = false
	}

	if okay {
		subscription, err = card.SubscribeToPlan(stripeCustomer, data.Plan, data.Email, data.LastFour, "")
		if err != nil {
			app.errorLog.Println(err)
			txnMsg = err.Error()
			okay = false
		}

		app.infoLog.Println("[handlers-api][CreateCustomerAndSubscribeToPlan] => (subscriptionID): ", subscription.ID)
	}

	if okay {
		productID, err := strconv.Atoi(data.ProductID)
		if err != nil {
			app.errorLog.Println(err)
			txnMsg = err.Error()
			okay = false
		}

		//////////////////////////////////////////////////////////////////////////////
		// Create New Customer
		//////////////////////////////////////////////////////////////////////////////
		customerID, err := app.SaveCustomer(data.FirstName, data.LastName, data.Email)
		if err != nil {
			app.errorLog.Println(err)
			txnMsg = err.Error()
			okay = false
		}
		app.infoLog.Println("[api][handlers-api][CreateCustomerAndSubscribeToPlan] => (customerID):", customerID)

		//////////////////////////////////////////////////////////////////////////////
		// Create New Transaction
		//////////////////////////////////////////////////////////////////////////////
		amount, err := strconv.Atoi(data.Amount)
		if err != nil {
			app.errorLog.Println(err)
			txnMsg = err.Error()
			okay = false
		}
		// expiryMonth, err := strconv.Atoi(data.ExpiryMonth)
		// if err != nil {
		// 	app.errorLog.Println(err)
		// 	txnMsg = err.Error()
		// 	okay = false
		// }
		// expiryYear, err := strconv.Atoi(data.ExpiryYear)
		// if err != nil {
		// 	app.errorLog.Println(err)
		// 	txnMsg = err.Error()
		// 	okay = false
		// }

		txn := models.Transaction{
			Amount:              amount,
			Currency:            "usd",
			LastFour:            data.LastFour,
			ExpiryMonth:         data.ExpiryMonth,
			ExpiryYear:          data.ExpiryYear,
			TransactionStatusID: 2,
		}

		txnId, err := app.SaveTransaction(txn)
		if err != nil {
			app.errorLog.Println(err)
			txnMsg = err.Error()
			okay = false
		}
		app.infoLog.Println("[api][handlers-api][CreateCustomerAndSubscribeToPlan] => (transactionID):", txnId)

		//////////////////////////////////////////////////////////////////////////////
		// Create New Order
		//////////////////////////////////////////////////////////////////////////////
		order := models.Order{
			WidgetID:      productID,
			TransactionID: txnId,
			CustomerID:    customerID,
			StatusID:      1,
			Quantity:      1,
			Amount:        amount,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		orderID, err := app.SaveOrder(order)
		if err != nil {
			app.errorLog.Println(err)
			txnMsg = err.Error()
			okay = false
		}
		app.infoLog.Println("[api][handlers-api][CreateCustomerAndSubscribeToPlan] => (orderID):", orderID)

	}

	resp :=
		jsonResponse{
			OK:      okay,
			Message: txnMsg,
		}

	out, err := json.MarshalIndent(resp, "", "   ")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)

}

// SaveCustomer saves a customer and returns the primary key
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		app.errorLog.Println(err)
		return 0, err
	}

	return id, nil
}

// SaveTransaction saves a transaction and returns the primary key
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {
	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return 0, err
	}

	return id, nil
}

// SaveOrder saves an order and returns the primary key
func (app *application) SaveOrder(order models.Order) (int, error) {
	id, err := app.DB.InsertOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return 0, err
	}

	return id, nil
}

func (app *application) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	logSnippet := "[api][handers-api][CreateAuthToken] =>"

	fmt.Printf("%s", logSnippet)

	var userInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &userInput)
	if err != nil {
		fmt.Printf("%s application.readJSON returned an error", logSnippet)
		app.badRequest(w, r, err)
		return
	}

	///////////////////////////////////////////////////
	// Retrieve user by email address
	///////////////////////////////////////////////////
	user, err := app.DB.GetUserByEmail(userInput.Email)
	if err != nil {
		fmt.Printf("%s app.DB.GetUserByEmail returned an error", logSnippet)
		app.invalidCredentials(w)
		return
	}

	app.infoLog.Printf("%s User found: %s", logSnippet, user.Email)
	fmt.Printf("%s User found: %s", logSnippet, user.Email)

	///////////////////////////////////////////////////
	// Validate password
	///////////////////////////////////////////////////
	validPassword, err := app.passwordMatches(user.Password, userInput.Password)
	if err != nil {
		app.invalidCredentials(w)
		return
	}

	if !validPassword {
		app.invalidCredentials(w)
		return
	}

	fmt.Printf("%s Password is valid for: %s", logSnippet, user.Email)

	///////////////////////////////////////////////////
	// Genereate Token
	///////////////////////////////////////////////////
	token, err := models.GenerateToken(user.ID, 24*time.Hour, models.ScopeAuthentication)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

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
