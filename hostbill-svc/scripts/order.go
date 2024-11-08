package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type TenantInfo struct {
	CompanyName             string `json:"CompanyName"`
	DomainName              string `json:"DomainName"`
	Country                 string `json:"Country"`
	State                   string `json:"State"`
	PostalCode              string `json:"PostalCode"`
	IsMultiCloudProductType bool   `json:"IsMultiCloudProducticType"`
}

type Data struct {
	Name          string     `json:"Name"`
	CrmIdentifier string     `json:"CrmIdentifier"`
	TenantInfo    TenantInfo `json:"TenantInfo"`
}

func main() {
	if len(os.Args) != 6 {
		fmt.Printf("Usage: %s <name=value> <crm_identifier=value> <country=value> <state=value> <postal_code=value>\n", os.Args[0])
		os.Exit(1)
	}

	args := make([]string, 5)
	for i, arg := range os.Args[1:] {
		pair := strings.SplitN(arg, "=", 2)
		if len(pair) != 2 {
			fmt.Println("Each argument must be of the form key=value")
			os.Exit(1)
		}
		args[i] = pair[1]
	}

	data := Data{
		Name:          args[0],
		CrmIdentifier: args[1],
		TenantInfo: TenantInfo{
			CompanyName:             args[0],
			DomainName:              args[0],
			Country:                 args[2],
			State:                   args[3],
			PostalCode:              args[4],
			IsMultiCloudProductType: false,
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("error marshalling JSON:", err)
		os.Exit(1)
	}

	fmt.Println("sending the following JSON data:", string(jsonData))

	url := "http://localhost:4000/api/v1/zerto"
	response, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		fmt.Println("error sending request:", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	fmt.Println("response status from server:", response.Status)
}
