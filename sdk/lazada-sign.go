package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sort"
	"strings"
)

const (
	SIGN_METHOD_HMAC   = "hmac"
	SIGN_METHOD_SHA256 = "sha256"
	CHARSET_UTF8       = "UTF-8"
)

func signAPIRequest(params map[string]string, body, appSecret, signMethod, apiName string) (string, error) {
	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build query string
	var query strings.Builder
	query.WriteString(apiName)
	for _, key := range keys {
		if value, exists := params[key]; exists && key != "" && value != "" {
			query.WriteString(key + value)
		}
	}

	// Append body if exists
	if body != "" {
		query.WriteString(body)
	}

	// Generate signature
	var signature []byte
	var err error

	if signMethod == SIGN_METHOD_HMAC {
		signature, err = encryptHMACSHA256(query.String(), appSecret)
		if err != nil {
			return "", err
		}
	} else if signMethod == SIGN_METHOD_SHA256 {
		signature, err = encryptHMACSHA256(query.String(), appSecret)
		if err != nil {
			return "", err
		}
	}

	// Convert signature to hex
	return byte2hex(signature), nil
}

func encryptHMACSHA256(data, secret string) ([]byte, error) {
	h := hmac.New(sha256.New, []byte(secret))
	_, err := io.WriteString(h, data)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func byte2hex(bytes []byte) string {
	return strings.ToUpper(hex.EncodeToString(bytes))
}
