package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"who-owes-me/actual"

	"github.com/go-chi/chi/v5"
	"who-owes-me/db"
	"who-owes-me/internal/envutil"
)

var splitTag = func() string {
	if tag := envutil.Getenv("SPLIT_TAG"); tag != "" {
		return tag
	}
	return "#gsu2026"
}()

func cleanNote(note string) string {
	result := strings.TrimSpace(strings.ReplaceAll(note, splitTag, ""))
	if result == "" {
		return "(no notes)"
	}
	return result
}

var funcMap = template.FuncMap{
	"formatMoney": func(cents int) string {
		if cents < 0 {
			return fmt.Sprintf("-$%.2f", float64(-cents)/100.0)
		}
		return fmt.Sprintf("$%.2f", float64(cents)/100.0)
	},
	"formatDate": func(dateStr string) string {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return dateStr
		}
		return t.Format("Jan 2, 2006")
	},
	"formatAidClassLabel": func(class string) string {
		switch class {
		case "needs_help":
			return "Needs Help"
		case "will_help":
			return "Will Help"
		default:
			return "Regular"
		}
	},
	"formatAidClassColor": func(class string) string {
		switch class {
		case "needs_help":
			return "is-danger"
		case "will_help":
			return "is-success"
		default:
			return "is-light has-text-grey-dark"
		}
	},
}

func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles("templates/base.html", "templates/"+tmplName)
	if err != nil {
		fmt.Printf("Template parsing error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		fmt.Printf("Template execution error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	renderTemplate(w, "error.html", struct {
		Code    int
		Message string
	}{Code: code, Message: message})
}

func getUserDashboardData(user *db.User) (interface{}, error) {
	actClient := actual.NewClient()
	
	deposits, _ := actClient.GetTaggedTransactionsByPayee(user.ActualPayeeID, splitTag)
	if deposits == nil {
		deposits = []actual.Transaction{}
	}

	splits, _ := db.GetSplitsForUser(user.ID)
	if splits == nil {
		splits = []db.ExpenseSplit{}
	}

	taggedTx, _ := actClient.GetTransactionsByTag(splitTag)
	txDateMap := map[string]string{}
	for _, t := range taggedTx {
		txDateMap[t.ID] = t.Date
	}

	balance := 0
	for i := range deposits {
		deposits[i].Notes = cleanNote(deposits[i].Notes)
		balance += deposits[i].Amount
	}
	for _, s := range splits {
		balance -= s.AmountOwed
	}

	return struct {
		User      *db.User
		Deposits  []actual.Transaction
		Splits    []db.ExpenseSplit
		TxDateMap map[string]string
		Balance   int
		SplitTag  string
	}{
		User:      user,
		Deposits:  deposits,
		Splits:    splits,
		TxDateMap: txDateMap,
		Balance:   balance,
		SplitTag:  splitTag,
	}, nil
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	isAdmin := r.Context().Value(isAdminCtxKey).(bool)

	if isAdmin {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	user := r.Context().Value(userCtxKey).(*db.User)
	http.Redirect(w, r, "/users/"+user.OIDCSub, http.StatusFound)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func handleUserDashboardBySub(w http.ResponseWriter, r *http.Request) {
	isAdmin := r.Context().Value(isAdminCtxKey).(bool)
	userVal := r.Context().Value(userCtxKey)
	
	sub := chi.URLParam(r, "sub")

	// Must be admin, or the user requesting their own page.
	if !isAdmin {
		if userVal == nil || userVal.(*db.User).OIDCSub != sub {
			renderError(w, http.StatusForbidden, "You don't have permission to view this page.")
			return
		}
	}

	user, err := db.GetUserBySub(sub)
	if err != nil {
		renderError(w, http.StatusNotFound, "User not found.")
		return
	}

	data, _ := getUserDashboardData(user)
	renderTemplate(w, "user.html", data)
}

type UserWithBalance struct {
	db.User
	Balance int `json:"balance"`
}

type PaymentWithUser struct {
	actual.Transaction
	MappedUserName       string `json:"mapped_user_name"`
	MappedUserID         int    `json:"mapped_user_id"`
	HasMappedUser        bool   `json:"has_mapped_user"`
}

func handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	users, _ := db.GetAllUsers()
	if users == nil {
		users = []db.User{}
	}

	actClient := actual.NewClient()
	apiErrors := []string{}
	
	payees, err := actClient.GetPayees()
	if err != nil {
		apiErrors = append(apiErrors, fmt.Sprintf("Failed to fetch payees: %v", err))
		fmt.Printf("Error fetching payees: %v\n", err)
	}
	if payees == nil {
		payees = []actual.Payee{}
	}
	payeesJSON, _ := json.Marshal(payees)

	// Fetch recent transactions to split, filtering by the specific tag, only expenses (negative amounts)
	allTagged, err := actClient.GetTransactionsByTag(splitTag)
	if err != nil {
		apiErrors = append(apiErrors, fmt.Sprintf("Failed to fetch transactions: %v", err))
		fmt.Printf("Error fetching transactions: %v\n", err)
	}
	var recentTransactions []actual.Transaction
	for _, t := range allTagged {
		if t.Amount < 0 {
			t.Notes = cleanNote(t.Notes)
			recentTransactions = append(recentTransactions, t)
		}
	}
	if recentTransactions == nil {
		recentTransactions = []actual.Transaction{}
	}

	var usersWithBalance []UserWithBalance
	payeeToUser := map[string]db.User{}
	for _, u := range users {
		payeeToUser[u.ActualPayeeID] = u
	}

	seenTx := map[string]bool{}
	var payments []PaymentWithUser
	for _, u := range users {
		balance := 0
		
		// 1. Add all Actual Budget transactions associated with this user's Payee
		deposits, _ := actClient.GetTaggedTransactionsByPayee(u.ActualPayeeID, splitTag)
		for _, d := range deposits {
			if seenTx[d.ID] {
				continue
			}
			seenTx[d.ID] = true
			balance += d.Amount // Positive for deposits/income, negative if refunded

			// Positive amounts are payments (deposits)
			if d.Amount > 0 {
				d.Notes = cleanNote(d.Notes)
				payments = append(payments, PaymentWithUser{
					Transaction:    d,
					MappedUserName: u.Name,
					MappedUserID:   u.ID,
					HasMappedUser:  true,
				})
			}
		}
		
		// 2. Subtract all splits they owe
		splits, _ := db.GetSplitsForUser(u.ID)
		for _, s := range splits {
			balance -= s.AmountOwed
		}
		
		usersWithBalance = append(usersWithBalance, UserWithBalance{
			User:    u,
			Balance: balance,
		})
	}

	// Also collect payee transactions that aren't mapped to any user
	for _, p := range payees {
		if _, ok := payeeToUser[p.ID]; ok {
			continue // already handled above
		}
		orphanTx, _ := actClient.GetTaggedTransactionsByPayee(p.ID, splitTag)
		for _, d := range orphanTx {
			if seenTx[d.ID] {
				continue
			}
			seenTx[d.ID] = true
			if d.Amount > 0 {
				d.Notes = cleanNote(d.Notes)
				payments = append(payments, PaymentWithUser{
					Transaction:   d,
					HasMappedUser: false,
				})
			}
		}
	}

	// Sort: unmapped first, then by date descending (most recent first)
	sort.Slice(payments, func(i, j int) bool {
		if payments[i].HasMappedUser != payments[j].HasMappedUser {
			return !payments[i].HasMappedUser // unmapped first
		}
		return payments[i].Date > payments[j].Date // most recent first
	})

	paymentsJSON, _ := json.Marshal(payments)

	usersJSON, _ := json.Marshal(usersWithBalance)

	allSplits, _ := db.GetAllSplits()
	if allSplits == nil {
		allSplits = []db.ExpenseSplit{}
	}
	splitsJSON, _ := json.Marshal(allSplits)

	payeeNameMap := map[string]string{}
	for _, p := range payees {
		payeeNameMap[p.ID] = p.Name
	}

	transactionsJSON, err := json.Marshal(recentTransactions)
	if err != nil {
		transactionsJSON = []byte("[]")
	}

	splitTxSet := map[string]bool{}
	for _, s := range allSplits {
		splitTxSet[s.ActualTransactionID] = true
	}

	data := struct {
		Users             []UserWithBalance
		UsersJSON         template.JS
		Payees            []actual.Payee
		PayeesJSON        template.JS
		Transactions      []actual.Transaction
		TransactionsJSON  template.JS
		PaymentsJSON      template.JS
		SplitsJSON        template.JS
		Error             string
		APIErrors         []string
		PayeeNames        map[string]string
		SplitTxSet        map[string]bool
		SplitTag          string
	}{
		Users:            usersWithBalance,
		UsersJSON:        template.JS(usersJSON),
		Payees:           payees,
		PayeesJSON:       template.JS(payeesJSON),
		Transactions:     recentTransactions,
		TransactionsJSON: template.JS(transactionsJSON),
		PaymentsJSON:     template.JS(paymentsJSON),
		SplitsJSON:       template.JS(splitsJSON),
		Error:            r.URL.Query().Get("error"),
		APIErrors:        apiErrors,
		PayeeNames:       payeeNameMap,
		SplitTxSet:       splitTxSet,
		SplitTag:         splitTag,
	}

	renderTemplate(w, "admin.html", data)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	oidcSub := r.FormValue("oidc_sub")
	aidClass := r.FormValue("aid_class")
	payeeID := r.FormValue("actual_payee_id")

	err := db.CreateUser(name, oidcSub, aidClass, payeeID)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Duplicate OIDC Subject", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	oidcSub := r.FormValue("oidc_sub")
	aidClass := r.FormValue("aid_class")
	payeeID := r.FormValue("actual_payee_id")

	err = db.UpdateUser(id, name, oidcSub, aidClass, payeeID)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Failed to update user", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func handleGetPayees(w http.ResponseWriter, r *http.Request) {
	actClient := actual.NewClient()
	payees, err := actClient.GetPayees()
	if err != nil {
		http.Error(w, "Error fetching payees", http.StatusInternalServerError)
		return
	}

	tmpl, _ := template.New("").ParseFiles("templates/payee_options.html")
	tmpl.ExecuteTemplate(w, "payee_options.html", payees)
}

func handleCreateSplits(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	txID := r.FormValue("actual_transaction_id")
	if txID == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	// Read dynamic form fields: split_amount_USERID=AMOUNT
	for key, values := range r.Form {
		if len(key) > 13 && key[:13] == "split_amount_" {
			userIDStr := key[13:]
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				continue
			}

			amountStr := values[0]
			amount, err := strconv.Atoi(amountStr)
			if err != nil {
				amount = 0
			}

			// Save to DB (SetSplit handles upsert or delete if <= 0)
			db.SetSplit(txID, userID, amount)
		}
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}
