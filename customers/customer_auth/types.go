package customer_auth

const appleIssuer = "https://appleid.apple.com"

type appleKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type appleKeysResponse struct {
	Keys []appleKey `json:"keys"`
}
