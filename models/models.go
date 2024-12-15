package models

// Define a map to hold valid transaction states
var ValidStates = map[string]bool{
	"win":  true,
	"lose": true,
}

// Define a map to hold valid Source-Type values
var ValidSourceTypes = map[string]bool{
	"game":    true,
	"server":  true,
	"payment": true,
}

// Define the Transaction struct to hold transaction data
type Transaction struct {
	State         string `json:"state"`
	Amount        string `json:"amount"`
	TransactionId string `json:"transactionId"`
}
