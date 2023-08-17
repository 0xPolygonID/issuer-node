package domain

// AuthenticationQrCodeResponse defines model for AuthenticationQrCodeResponse.
type AuthenticationQrCodeResponse struct {
	Body struct {
		CallbackUrl string        `json:"callbackUrl"`
		Reason      string        `json:"reason"`
		Scope       []interface{} `json:"scope"`
	} `json:"body"`
	From string `json:"from"`
	Id   string `json:"id"`
	Thid string `json:"thid"`
	Typ  string `json:"typ"`
	Type string `json:"type"`
}
