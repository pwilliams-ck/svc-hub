package services

// ZertoSessionRequest represents the request body for obtaining a zerto session token
type ZertoSessionRequest struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ZertoSessionResponse represents the response body for obtaining a zerto session token
type ZertoSessionResponse struct {
	SessionId string `json:"sessionId"`
}

type ZertoCreateOrgRequest struct {
	Name          string `json:"Name"`
	CrmIdentifier string `json:"CrmIdentifier"`
	TenantInfo    struct {
		CompanyName             string `json:"CompanyName"`
		DomainName              string `json:"DomainName"`
		Country                 string `json:"Country"`
		State                   string `json:"State"`
		PostalCode              string `json:"PostalCode"`
		IsMultiCloudProductType bool   `json:"IsMultiCloudProductType"`
	} `json:"TenantInfo"`
}

// zertoCreateOrgResponse represents the response body for creating a zerto organization
type ZertoCreateOrgResponse struct {
	Message string `json:"message"`
}
