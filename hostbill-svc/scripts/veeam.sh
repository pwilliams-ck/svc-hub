#!/bin/bash

# Check for the correct number of arguments
if [ "$#" -ne 2 ]; then
	echo "Usage: $0 <organization_name> <quota_gb>"
	exit 1
fi

# Split input based on "=" and get the value (index 1)
organization_name="$(cut -d'=' -f2 <<<"$1")"
quota_gb="$(cut -d'=' -f2 <<<"$2")"

# Validate quota_gb as a number
if ! [[ "$quota_gb" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
	echo "Error: quota_gb must be a valid number."
	exit 1
fi

url="http://localhost:4000/api/v1/veeam"

# Construct JSON data correctly, keeping quota_gb as a string
jsonData="{\"OrganizationName\":\"$organization_name\",\"QuotaGb\":\"$quota_gb\"}"

# Debug: Print JSON to ensure it's correctly formatted
echo "Sending the following JSON data: $jsonData"

# Send POST request with Curl
response=$(curl -k -X POST "$url" \
	-H "Content-Type: application/json" \
	-d "$jsonData")

# Print response
echo "Response from server: $response"
