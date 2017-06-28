package admin

import (
	"io/ioutil"
	"bytes"
	"net/http"
	"github.com/openbankit/go-base/hash"
)

// Returns admin action signature's content
func getAdminHelperSignatureBase(bodyString string, timeCreated string) string {
	return "{method: 'post', body: '" + bodyString + "', timestamp: '" + timeCreated + "'}"
}

// Returns content hash of request
func GetContentsHash(request *http.Request, timeCreated string) [32]byte {
	// Read the content
	var bodyBytes []byte
	if request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(request.Body)
		// Restore the io.ReadCloser to its original state
		request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Use the content
	bodyString := string(bodyBytes)

	signatureBase := getAdminHelperSignatureBase(bodyString, timeCreated)
	hashBase := hash.Hash([]byte(signatureBase))

	return hashBase
}
