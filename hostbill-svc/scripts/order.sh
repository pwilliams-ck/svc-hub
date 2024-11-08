#!/bin/bash

# Validate correct number of args
if [ "$#" -ne 5 ]; then
	echo "Usage: $0 <name> <crm_identifier> <country> <state> <postal_code>"
	exit 1
fi

# Split input based on "=" and get the value (index 1)
# We re-use name for both the name and company name in json_data.
name="$(cut -d'=' -f2 <<<"$1")"
crm_identifier="$(cut -d'=' -f2 <<<"$2")"
country="$(cut -d'=' -f2 <<<"$3")"
state="$(cut -d'=' -f2 <<<"$4")"
postal_code="$(cut -d'=' -f2 <<<"$5")"

# Validate crm_identifier as a number
if ! [[ "$crm_identifier" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
	echo "error: crm_identifier must be a valid number."
	exit 1
fi

url="http://deployment-server.cloudkey.io/api/v1/zerto"

# Construct JSON data correctly, keeping crm_identifier as a string
json_data=$(
	cat <<EOF
{
    "Name": "$name",
    "CrmIdentifier": "$crm_identifier",
    "TenantInfo": {
        "CompanyName": "$name",
        "DomainName": "$name",
        "Country": "$country",
        "State": "$state",
        "PostalCode": "$postal_code",
        "IsMultiCloudProductType": false
    }
}
EOF
)

# Debug: Print JSON to ensure it's correctly formatted
echo "sending the following JSON data: $json_data"

# Send POST request with Curl
response=$(curl -k -X POST "$url" \
	-H "Content-Type: application/json" \
	-d "$json_data")

# Print response
echo "response from server: $response"
