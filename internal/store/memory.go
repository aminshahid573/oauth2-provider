package store

import "sync"

var Clients = map[string]string{
	"demoapp": "http://localhost:9000/callback",
}

var Codes = map[string]string{}

var CodeMux sync.Mutex

func IsValidClient(clientID, redirectURI string) bool {
	expected, ok := Clients[clientID]
	return ok && expected == redirectURI
}

func SaveCode(code, username string) {
	CodeMux.Lock()
	defer CodeMux.Unlock()
	Codes[code] = username
}

func GetUserNameFromCode(code string) (string, bool) {
	CodeMux.Lock()
	defer CodeMux.Unlock()

	username, ok := Codes[code]
	if ok {
		delete(Codes, code) // one-time use
	}
	return username, ok
}
