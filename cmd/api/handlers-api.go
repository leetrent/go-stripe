package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/leetrent/go-stripe/internal/cards"
	"github.com/leetrent/go-stripe/internal/encryption"
	"github.com/leetrent/go-stripe/internal/models"
	"github.com/leetrent/go-stripe/internal/urlsigner"
	"github.com/stripe/stripe-go/v72"
	"golang.org/x/crypto/bcrypt"
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

		txn := models.Transaction{
			Amount:              amount,
			Currency:            "usd",
			LastFour:            data.LastFour,
			ExpiryMonth:         data.ExpiryMonth,
			ExpiryYear:          data.ExpiryYear,
			TransactionStatusID: 2,
			PaymentInent:        subscription.ID,
			PaymentMethod:       data.PaymentMethod,
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
	// Generate Token
	///////////////////////////////////////////////////
	token, err := models.GenerateToken(user.ID, 24*time.Hour, models.ScopeAuthentication)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	///////////////////////////////////////////////////
	// Save Genereate Token to Database
	///////////////////////////////////////////////////
	err = app.DB.InsertToken(token, user)
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

func (app *application) authenticateToken(r *http.Request) (*models.User, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return nil, errors.New("no authorization header received")
	}

	oneSpace := " "
	headerParts := strings.Split(authorizationHeader, oneSpace)
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("no bearer token found for user")
	}

	token := headerParts[1]
	if len(token) != 26 {
		return nil, errors.New("bearer token is an incorrect length")
	}

	user, err := app.DB.GetUserForToken(token)
	if err != nil {
		return nil, errors.New("not matching token found for user")
	}

	return user, nil

}

func (app *application) CheckAuthentication(w http.ResponseWriter, r *http.Request) {
	user, err := app.authenticateToken(r)
	if err != nil {
		app.invalidCredentials(w)
		return
	}

	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = false
	payload.Message = fmt.Sprintf("%s has been authenticated", user.Email)

	app.writeJSON(w, http.StatusOK, payload)
}

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
		ExpiryMonth     int    `json:"expiry_month"`
		ExpiryYear      int    `json:"expiry_year"`
		LastFour        string `json:"last_four"`
	}

	err := app.readJSON(w, r, &txnData)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	logSnippet := "[api][handlers][VirtualTerminalPaymentSucceeded] =>"
	app.infoLog.Printf("%s (PaymentAmount): %d", logSnippet, txnData.PaymentAmount)

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(txnData.PaymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	pm, err := card.GetPaymentMethod(txnData.PaymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
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
		PaymentInent:        txnData.PaymentIntent,
		PaymentMethod:       txnData.PaymentMethod,
		BankReturnCode:      pi.Charges.Data[0].ID,
		TransactionStatusID: 2,
	}

	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	app.infoLog.Printf("%s (txnId): %d", logSnippet, txnID)
	fmt.Printf("%s (txnId): %d", logSnippet, txnID)

	app.writeJSON(w, http.StatusOK, txn)
}

func (app *application) SendPasswordResetEmail(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	////////////////////////////////////
	// Verify that email address exists
	////////////////////////////////////
	_, err = app.DB.GetUserByEmail(payload.Email)
	if err != nil {
		var resp struct {
			Error   bool   `json:"error"`
			Message string `json:"message"`
		}
		resp.Error = true
		resp.Message = "No matching email address was found in our system."
		app.writeJSON(w, http.StatusAccepted, resp)
		return
	}

	link := fmt.Sprintf("%s/reset-password?email=%s", app.config.frontendUrl, payload.Email)
	sign := urlsigner.Signer{
		Secret: []byte(app.config.secretKey),
	}

	signedLink := sign.GenerateTokenFromString(link)

	var data struct {
		Link string
	}

	data.Link = signedLink

	err = app.SendMail("info@widgets.com", payload.Email, "Password Reset Request", "password-reset", data)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	app.writeJSON(w, http.StatusCreated, resp)
}

func (app *application) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
	}

	encryptor := encryption.Encryption{
		Key: []byte(app.config.secretKey),
	}

	decryptedEmail, err := encryptor.Decrypt(payload.Email)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
	}

	user, err := app.DB.GetUserByEmail(decryptedEmail)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
	}

	err = app.DB.UpdatePasswordForUser(user, string(newHash))
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "[api][handlers-api] => Password has been changed."

	app.writeJSON(w, http.StatusCreated, resp)
}

func (app *application) AllSales(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		PageSize    int `json:"page_size"`
		CurrentPage int `json:"page"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	fmt.Println("")
	logSnippet := "\n[api][handlers-api][AllSales] =>"
	fmt.Printf("%s (payload.PageSize)....: %d", logSnippet, payload.PageSize)
	fmt.Printf("%s (payload.CurrentPage).: %d", logSnippet, payload.CurrentPage)

	allNonRecurringSales, lastPage, totalRecords, err := app.DB.GetAllOrdersPaginated(false, payload.PageSize, payload.CurrentPage)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	var resp struct {
		CurrentPage  int             `json:"current_page"`
		PageSize     int             `json:"page_size"`
		LastPage     int             `json:"last_page"`
		TotalRecords int             `json:"total_records"`
		Orders       []*models.Order `json:"orders"`
	}

	resp.CurrentPage = payload.CurrentPage
	resp.PageSize = payload.PageSize
	resp.LastPage = lastPage
	resp.TotalRecords = totalRecords
	resp.Orders = allNonRecurringSales

	app.writeJSON(w, http.StatusOK, resp)
}

// func (app *application) AllSubscriptions(w http.ResponseWriter, r *http.Request) {
// 	allRecurringSales, err := app.DB.GetAllOrders(true)
// 	if err != nil {
// 		app.errorLog.Println(err)
// 		app.badRequest(w, r, err)
// 		return
// 	}

// 	app.writeJSON(w, http.StatusOK, allRecurringSales)
// }

func (app *application) AllSubscriptions(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		PageSize    int `json:"page_size"`
		CurrentPage int `json:"page"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	fmt.Println("")
	logSnippet := "\n[api][handlers-api][AllSubscriptions] =>"
	fmt.Printf("%s (payload.PageSize)....: %d", logSnippet, payload.PageSize)
	fmt.Printf("%s (payload.CurrentPage).: %d", logSnippet, payload.CurrentPage)

	allNonRecurringSales, lastPage, totalRecords, err := app.DB.GetAllOrdersPaginated(true, payload.PageSize, payload.CurrentPage)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	var resp struct {
		CurrentPage  int             `json:"current_page"`
		PageSize     int             `json:"page_size"`
		LastPage     int             `json:"last_page"`
		TotalRecords int             `json:"total_records"`
		Orders       []*models.Order `json:"orders"`
	}

	resp.CurrentPage = payload.CurrentPage
	resp.PageSize = payload.PageSize
	resp.LastPage = lastPage
	resp.TotalRecords = totalRecords
	resp.Orders = allNonRecurringSales

	app.writeJSON(w, http.StatusOK, resp)
}

func (app *application) GetSale(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	orderID, err := strconv.Atoi(id)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	order, err := app.DB.GetOrderByID(orderID)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, order)
}

func (app *application) RefundCharge(w http.ResponseWriter, r *http.Request) {
	var chargeToRefund struct {
		ID            int    `json:"id"`
		PaymentIntent string `json:"pi"`
		Amount        int    `json:"amount"`
		Currency      string `json:"currency"`
	}

	err := app.readJSON(w, r, &chargeToRefund)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	//////////////////////////////////////////////////////////
	// TODO: Validate refund amount by performing a DB lookup
	//////////////////////////////////////////////////////////

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: chargeToRefund.Currency,
	}

	err = card.Refund(chargeToRefund.PaymentIntent, chargeToRefund.Amount)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	//////////////////////////////////////////////////////////
	// UPDATE ORDER STATUS IN DATABASE
	//////////////////////////////////////////////////////////
	err = app.DB.UpdateOrderStatus(chargeToRefund.ID, 2)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, errors.New("charge amount was successfully refunded but database could not be updated"))
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "Charge amount has been successfully refunded"

	app.writeJSON(w, http.StatusOK, resp)
}

func (app *application) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	var subToCancel struct {
		ID            int    `json:"id"`
		PaymentIntent string `json:"pi"`
		Currency      string `json:"currency"`
	}

	err := app.readJSON(w, r, &subToCancel)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: subToCancel.Currency,
	}

	/////////////////////////////////////////////////////////////
	// PaymentIntent is holding subcription id for subscriptions
	/////////////////////////////////////////////////////////////
	err = card.CancelSubscription(subToCancel.PaymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	//////////////////////////////////////////////////////////
	// UPDATE ORDER STATUS IN DATABASE
	//////////////////////////////////////////////////////////
	err = app.DB.UpdateOrderStatus(subToCancel.ID, 3)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, errors.New("subscription was successfully cancelled but database could not be updated"))
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "Subscription cancelled"

	app.writeJSON(w, http.StatusOK, resp)
}

func (app *application) AllUsers(w http.ResponseWriter, r *http.Request) {
	allUsers, err := app.DB.GetAllUsers()
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, allUsers)
}

func (app *application) OneUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	user, err := app.DB.GetOneUser(userID)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, user)
}
