package acme

import (
	"bytes"
	"crypto"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// acmeSignedMessage is the ACME signed message payload
type acmeSignedMessage struct {
	Payload         string `json:"payload"`
	ProtectedHeader string `json:"protected"`
	Signature       string `json:"signature"`
}

// ProtectedHeader piece of the ACME payload
type protectedHeader struct {
	Algorithm  string      `json:"alg"`
	JsonWebKey *jsonWebKey `json:"jwk,omitempty"`
	KeyId      string      `json:"kid,omitempty"`
	Nonce      string      `json:"nonce"`
	Url        string      `json:"url"`
}

// jsonWebKey for the ACME protectedHeader
type jsonWebKey struct {
	KeyType        string `json:"kty,omitempty"`
	PublicExponent string `json:"e,omitempty"`   // RSA
	Modulus        string `json:"n,omitempty"`   // RSA
	CurveName      string `json:"crv,omitempty"` // EC
	CurvePointX    string `json:"x,omitempty"`   // EC
	CurvePointY    string `json:"y,omitempty"`   // EC
}

// AccountKey is the necessary account / key information for signed message generation
type AccountKey struct {
	Key crypto.PrivateKey
	Kid string
}

// postToUrlSigned posts the payload to the specified url, using the specified AccountKeyInfo
// and returns the response body (data / bytes) and headers from ACME
func (service *Service) postToUrlSigned(payload any, url string, accountKey AccountKey) (body []byte, headers http.Header, err error) {
	// message is what will ultimately be posted to ACME
	var message acmeSignedMessage

	/// header
	var header protectedHeader

	// alg
	header.Algorithm, err = accountKey.signingAlg()
	if err != nil {
		return nil, nil, err
	}

	// key or kid
	// use kid if available, otherwise use jsonWebKey
	if accountKey.Kid != "" {
		header.JsonWebKey = nil
		header.KeyId = accountKey.Kid
	} else {
		header.JsonWebKey, err = accountKey.jwk()
		header.KeyId = ""
	}

	// nonce
	header.Nonce, err = service.nonceManager.Nonce()
	if err != nil {
		return nil, nil, err
	}

	// url
	header.Url = url

	// encord and insert into message
	message.ProtectedHeader, err = encodeJson(header)
	if err != nil {
		return nil, nil, err
	}
	/// header (end)

	/// payload
	message.Payload, err = encodeJson(payload)
	if err != nil {
		return nil, nil, err
	}

	/// signature
	message.Signature, err = accountKey.Sign(message)
	if err != nil {
		return nil, nil, err
	}

	/// post
	messageJson, err := json.Marshal(message)
	if err != nil {
		return nil, nil, err
	}

	response, err := http.Post(url, "application/jose+json", bytes.NewBuffer(messageJson))
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()

	// read body of response
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove (debugging)
	service.logger.Println(string(bodyBytes))

	// check if the response was an AcmeError
	acmeError, err := unmarshalErrorResponse(bodyBytes)

	// TODO: Retry logic, using scoped nonce
	if acmeError.Type == "urn:ietf:params:acme:error:badNonce" {
		// TODO
		// scoped retry nonce should NOT be saved to manager
	} else {
		// save nonce in manager
		nonce := response.Header.Get("Replay-Nonce")
		service.logger.Println(nonce)
		nonceErr := service.nonceManager.SaveNonce(nonce)
		if nonceErr != nil {
			// no need to error out of routine, just log the save failure
			service.logger.Println(nonceErr)
		}
	}

	// re: acmeError decode
	// if it didn't error, that means an error response WAS decoded
	if err == nil {
		return nil, nil, acmeError.Error()
	}

	return bodyBytes, response.Header, nil
}
