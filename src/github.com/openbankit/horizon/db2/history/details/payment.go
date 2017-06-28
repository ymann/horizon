package details

type Payment struct {
	Asset
	Fee    Fee    `json:"fee"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
}
