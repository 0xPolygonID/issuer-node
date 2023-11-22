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
type VerifyGSTINResponse struct {
	Result struct {
		GstnDetailed struct {
			GstinStatus string `json:"gstinStatus"`
		} `json:"gstnDetailed"`
	} `json:"result"`
}

type GSTNRecord struct {
	ApplicationStatus string `json:"applicationStatus"`
	RegistrationName  string `json:"registrationName"`
	MobNum            string `json:"mobNum"`
	RegType           string `json:"regType"`
	EmailId           string `json:"emailId"`
	TINNumber         string `json:"tinNumber"`
	GSTINRefId        string `json:"gstinRefId"`
	GSTIN             string `json:"gstin"`
}

type PrincipalPlaceSplitAddress struct {
	District    []string   `json:"district"`
	State       [][]string `json:"state"`
	City        []string   `json:"city"`
	Pincode     string     `json:"pincode"`
	Country     []string   `json:"country"`
	AddressLine string     `json:"addressLine"`
}

type AdditionalPlaceSplitAddress struct {
	District    []string   `json:"district"`
	State       [][]string `json:"state"`
	City        []string   `json:"city"`
	Pincode     string     `json:"pincode"`
	Country     []string   `json:"country"`
	AddressLine string     `json:"addressLine"`
}

type GSTNDetailed struct {
	ConstitutionOfBusiness      string                      `json:"constitutionOfBusiness"`
	LegalNameOfBusiness         string                      `json:"legalNameOfBusiness"`
	TradeNameOfBusiness         string                      `json:"tradeNameOfBusiness"`
	CentreJurisdiction          string                      `json:"centreJurisdiction"`
	StateJurisdiction           string                      `json:"stateJurisdiction"`
	RegistrationDate            string                      `json:"registrationDate"`
	TaxPayerDate                string                      `json:"taxPayerDate"`
	TaxPayerType                string                      `json:"taxPayerType"`
	GSTINStatus                 string                      `json:"gstinStatus"`
	CancellationDate            string                      `json:"cancellationDate"`
	EInvoicingStatus            string                      `json:"e-invoicingStatus"`
	NatureOfBusinessActivities  []string                    `json:"natureOfBusinessActivities"`
	PrincipalPlaceAddress       string                      `json:"principalPlaceAddress"`
	PrincipalPlaceLatitude      string                      `json:"principalPlaceLatitude"`
	PrincipalPlaceLongitude     string                      `json:"principalPlaceLongitude"`
	PrincipalPlaceBuildingName  string                      `json:"principalPlaceBuildingNameFromGST"`
	PrincipalPlaceBuildingNo    string                      `json:"principalPlaceBuildingNoFromGST"`
	PrincipalPlaceFlatNo        string                      `json:"principalPlaceFlatNo"`
	PrincipalPlaceStreet        string                      `json:"principalPlaceStreet"`
	PrincipalPlaceLocality      string                      `json:"principalPlaceLocality"`
	PrincipalPlaceCity          string                      `json:"principalPlaceCity"`
	PrincipalPlaceDistrict      string                      `json:"principalPlaceDistrict"`
	PrincipalPlaceState         string                      `json:"principalPlaceState"`
	PrincipalPlacePincode       string                      `json:"principalPlacePincode"`
	AdditionalPlaceAddress      string                      `json:"additionalPlaceAddress"`
	AdditionalPlaceLatitude     string                      `json:"additionalPlaceLatitude"`
	AdditionalPlaceLongitude    string                      `json:"additionalPlaceLongitude"`
	AdditionalPlaceBuildingName string                      `json:"additionalPlaceBuildingNameFromGST"`
	AdditionalPlaceBuildingNo   string                      `json:"additionalPlaceBuildingNoFromGST"`
	AdditionalPlaceFlatNo       string                      `json:"additionalPlaceFlatNo"`
	AdditionalPlaceStreet       string                      `json:"additionalPlaceStreet"`
	AdditionalPlaceLocality     string                      `json:"additionalPlaceLocality"`
	AdditionalPlaceCity         string                      `json:"additionalPlaceCity"`
	AdditionalPlaceDistrict     string                      `json:"additionalPlaceDistrict"`
	AdditionalPlaceState        string                      `json:"additionalPlaceState"`
	AdditionalPlacePincode      string                      `json:"additionalPlacePincode"`
	AdditionalAddressArray      []interface{}               `json:"additionalAddressArray"`
	LastUpdatedDate             string                      `json:"lastUpdatedDate"`
	PrincipalPlaceSplitAddress  PrincipalPlaceSplitAddress  `json:"principalPlaceSplitAddress"`
	AdditionalPlaceSplitAddress AdditionalPlaceSplitAddress `json:"additionalPlaceSplitAddress"`
}

type VerifyData struct {
	GSTNDetailed GSTNDetailed `json:"gstnDetailed"`
	GSTNRecords  []GSTNRecord `json:"gstnRecords"`
	GSTIN        string       `json:"gstin"`
}

type VerifyGSTINResponseNew struct {
	Task       string `json:"task"`
	Id         string `json:"id"`
	Essentials struct {
		GSTIN string `json:"gstin"`
	} `json:"essentials"`
	PatronId string     `json:"patronId"`
	Result   VerifyData `json:"result"`
}

//	{
//	    "service": "Identity",
//	    "itemId": "655d9a126dab5b00238d9db4",
//	    "task": "verifyAadhaar",
//	    "essentials": {
//	        "uid": "894365783749"
//	    },
//	    "accessToken": "c3c86o33m3q3qo1z60bd3zvl4bp4yrjg",
//	    "id": "655d9a1e6dab5b00238d9db5",
//	    "response": {
//	        "url": {},
//	        "id": "655d9a1e3d48d00037d61f65",
//	        "result": {
//	            "verified": "true",
//	            "ageBand": "20-30",
//	            "state": "Karnataka",
//	            "mobileNumber": "*******303",
//	            "gender": "MALE"
//	        },
//	        "instance": {}
//	    }
//	}
type VerifyAadhaarResponse struct {
	Service    string `json:"service"`
	ItemID     string `json:"itemId"`
	Task       string `json:"task"`
	Essentials struct {
		UID string `json:"uid"`
	} `json:"essentials"`
	AccessToken string `json:"accessToken"`
	ID          string `json:"id"`
	Response    struct {
		URL    struct{} `json:"url"`
		ID     string   `json:"id"`
		Result struct {
			Verified     string `json:"verified"`
			AgeBand      string `json:"ageBand"`
			State        string `json:"state"`
			MobileNumber string `json:"mobileNumber"`
			Gender       string `json:"gender"`
		} `json:"result"`
		Instance struct{} `json:"instance"`
	} `json:"response"`
}
