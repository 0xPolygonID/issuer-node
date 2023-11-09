package domain

type SinzyLoginResponse struct {
	Id      string `json:"id"`
	Ttl     int    `json:"ttl"`
	Created string `json:"created"`
	UserId  string `json:"userId"`
}

type DigilockerURLResponse struct {
	Id       string `json:"id"`
	PatronId string `json:"patronId"`
	Task     string `json:"task"`
	Result   struct {
		URL       string `json:"url"`
		RequestId string `json:"requestId"`
	} `json:"result"`
}

type DigilockerDocument struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Size        string   `json:"size"`
	Date        string   `json:"date"`
	Parent      string   `json:"parent"`
	Mime        []string `json:"mime"`
	Doctype     string   `json:"doctype"`
	Description string   `json:"description"`
	Issuerid    string   `json:"issuerid"`
	Issuer      string   `json:"issuer"`
	Id          string   `json:"id"`
}

type DigilockerDocumentList struct {
	Files []DigilockerDocument `json:"files"`
}

type VerificationIdentity struct {
	Type            string   `json:"type"`
	Email           string   `json:"email"`
	CallbackUrl     string   `json:"callbackUrl"`
	Images          []string `json:"images"`
	AccessToken     string   `json:"accessToken"`
	AutoRecognition []string `json:"autoRecognition"`
	Verification    []string `json:"verification"`
	ForgeryCheck    []string `json:"forgeryCheck"`
	Id              string   `json:"id"`
	PatronId        string   `json:"patronId"`
}

type VerifyAdharResponse struct {
	Verified     bool   `json:"verified"`
	AgeBand      string `json:"ageBand"`
	State        string `json:"state"`
	MobileNumber string `json:"mobileNumber"`
	Gender       string `json:"gender"`
}

type VerifyPANResponse struct {
	Verified      bool   `json:"verified"`
	Message       string `json:"message"`
	UpstreamName  string `json:"upstreamName"`
	PanStatus     string `json:"panStatus"`
	PanStatusCode string `json:"panStatusCode"`
}
