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
	"negate": func(cents int) int {
		return -cents
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

type LedgerRow struct {
	Date           string
	Notes          string
	AmountOwed     int
	IsCredit       bool
	RunningBalance int
}

func getUserDashboardData(user *db.User) (interface{}, error) {
	actClient := actual.NewClient()
	
	splits, _ := db.GetSplitsForUser(user.ID)
	if splits == nil {
		splits = []db.ExpenseSplit{}
	}

	taggedTx, _ := actClient.GetTransactionsByTag(splitTag)
	txMap := map[string]actual.Transaction{}
	for _, t := range taggedTx {
		t.Notes = cleanNote(t.Notes)
		txMap[t.ID] = t
	}

	var rows []LedgerRow
	balance := 0
	for _, s := range splits {
		date := s.ExpenseDate
		notes := s.ExpenseNote
		isCredit := s.AutoCreated
		if tx, ok := txMap[s.ActualTransactionID]; ok {
			isCredit = tx.Amount > 0
			if date == "" {
				date = tx.Date
			}
			if notes == "" {
				notes = tx.Notes
			}
		}
		if date == "" {
			date = "Unknown"
		}
		if notes == "" {
			notes = "Expense Share"
		}

		rows = append(rows, LedgerRow{
			Date:       date,
			Notes:      notes,
			AmountOwed: s.AmountOwed,
			IsCredit:   isCredit,
		})
	}

	// Sort chronologically (oldest first)
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Date < rows[j].Date
	})

	// Calculate running balance
	runningBalance := 0
	for i := range rows {
		if rows[i].IsCredit {
			runningBalance += rows[i].AmountOwed
		} else {
			runningBalance -= rows[i].AmountOwed
		}
		rows[i].RunningBalance = runningBalance
	}
	balance = runningBalance

	// Reverse to show newest first
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}

	return struct {
		User       *db.User
		LedgerRows []LedgerRow
		Balance    int
		SplitTag   string
	}{
		User:       user,
		LedgerRows: rows,
		Balance:    balance,
		SplitTag:   splitTag,
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

	// Fetch all tagged transactions (both deposits and expenses)
	allTagged, err := actClient.GetTransactionsByTag(splitTag)
	if err != nil {
		apiErrors = append(apiErrors, fmt.Sprintf("Failed to fetch transactions: %v", err))
		fmt.Printf("Error fetching transactions: %v\n", err)
	}
	for i := range allTagged {
		allTagged[i].Notes = cleanNote(allTagged[i].Notes)
	}
	if allTagged == nil {
		allTagged = []actual.Transaction{}
	}

	var usersWithBalance []UserWithBalance
	payeeToUser := map[string]db.User{}
	for _, u := range users {
		payeeToUser[u.ActualPayeeID] = u
	}

	allSplits, _ := db.GetAllSplits()
	if allSplits == nil {
		allSplits = []db.ExpenseSplit{}
	}

	splitTxSet := map[string]bool{}
	for _, s := range allSplits {
		splitTxSet[s.ActualTransactionID] = true
	}

	// Auto-split: for any positive transaction without a split whose payee maps to a user,
	// create a split for the full amount
	for _, tx := range allTagged {
		if splitTxSet[tx.ID] {
			continue
		}
		if user, ok := payeeToUser[tx.Payee]; ok {
			if tx.Amount > 0 {
				db.SetAutoSplit(tx.ID, user.ID, tx.Amount, tx.Date, tx.Notes)
				splitTxSet[tx.ID] = true
			}
		}
	}

	// Re-read splits after auto-split
	allSplits, _ = db.GetAllSplits()
	if allSplits == nil {
		allSplits = []db.ExpenseSplit{}
	}

	txMap := map[string]actual.Transaction{}
	for _, t := range allTagged {
		txMap[t.ID] = t
	}

	// Calculate balances from splits
	for _, u := range users {
		balance := 0
		for _, s := range allSplits {
			if s.UserID == u.ID {
				isCredit := s.AutoCreated
				if tx, ok := txMap[s.ActualTransactionID]; ok {
					isCredit = tx.Amount > 0
				}
				if isCredit {
					balance += s.AmountOwed
				} else {
					balance -= s.AmountOwed
				}
			}
		}
		usersWithBalance = append(usersWithBalance, UserWithBalance{
			User:    u,
			Balance: balance,
		})
	}

	usersJSON, _ := json.Marshal(usersWithBalance)

	// Build payee-to-user map for the frontend
	payeeToUserMap := map[string]int{}
	for _, u := range users {
		if u.ActualPayeeID != "" {
			payeeToUserMap[u.ActualPayeeID] = u.ID
		}
	}
	payeeToUserMapJSON, _ := json.Marshal(payeeToUserMap)

	splitsJSON, _ := json.Marshal(allSplits)
	transactionsJSON, _ := json.Marshal(allTagged)

	data := struct {
		Users              []UserWithBalance
		UsersJSON          template.JS
		Payees             []actual.Payee
		PayeesJSON         template.JS
		Transactions       []actual.Transaction
		TransactionsJSON   template.JS
		SplitsJSON         template.JS
		Error              string
		APIErrors          []string
		PayeeToUserMapJSON template.JS
		SplitTxSet         map[string]bool
		SplitTag           string
	}{
		Users:              usersWithBalance,
		UsersJSON:          template.JS(usersJSON),
		Payees:             payees,
		PayeesJSON:         template.JS(payeesJSON),
		Transactions:       allTagged,
		TransactionsJSON:   template.JS(transactionsJSON),
		SplitsJSON:         template.JS(splitsJSON),
		Error:              r.URL.Query().Get("error"),
		APIErrors:          apiErrors,
		PayeeToUserMapJSON: template.JS(payeeToUserMapJSON),
		SplitTxSet:         splitTxSet,
		SplitTag:           splitTag,
	}

	renderTemplate(w, "admin.html", data)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	names := r.Form["name"]
	oidcSubs := r.Form["oidc_sub"]
	aidClasses := r.Form["aid_class"]
	payeeIDs := r.Form["actual_payee_id"]

	for i := range names {
		if strings.TrimSpace(names[i]) == "" || strings.TrimSpace(oidcSubs[i]) == "" {
			continue
		}
		
		var aidClass string
		if i < len(aidClasses) {
			aidClass = aidClasses[i]
		}
		
		var payeeID string
		if i < len(payeeIDs) {
			payeeID = payeeIDs[i]
		}

		err := db.CreateUser(names[i], oidcSubs[i], aidClass, payeeID)
		if err != nil {
			// Log error but continue with other users
			fmt.Printf("Error creating user %s: %v\n", names[i], err)
		}
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

func handleRefreshCache(w http.ResponseWriter, r *http.Request) {
	actual.ClearCache()
	http.Redirect(w, r, "/admin", http.StatusFound)
}

func handleCreateSplits(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	txID := r.FormValue("actual_transaction_id")
	txDate := r.FormValue("actual_transaction_date")
	txNote := r.FormValue("actual_transaction_note")
	if txID == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	// Always clear existing splits first so removed participants are actually deleted,
	// and "Clear Splits" action can just submit an empty list.
	db.ClearSplitsForTx(txID)

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

			db.SetSplit(txID, userID, amount, txDate, txNote)
		}
	}

	// Optional: map a payee to a user (one-user-save popup)
	mapUserIDStr := r.FormValue("map_payee_to_user_id")
	payeeIDToMap := r.FormValue("payee_id_to_map")
	if mapUserIDStr != "" && payeeIDToMap != "" {
		mapUserID, err := strconv.Atoi(mapUserIDStr)
		if err == nil {
			existingUser, err := db.GetUserByID(mapUserID)
			if err == nil {
				db.UpdateUser(existingUser.ID, existingUser.Name, existingUser.OIDCSub, existingUser.AidClass, payeeIDToMap)
			}
		}
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}
