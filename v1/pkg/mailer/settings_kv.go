package mailer

import "strings"

// ConfigFromSMTPSettings builds Config from settings map (group=smtp).
// Keys: smtp.host, smtp.port, smtp.username, smtp.password, smtp.fromName, smtp.encryption.
func ConfigFromSMTPSettings(kv map[string]string) Config {
	return Config{
		Host:       strings.TrimSpace(kv["smtp.host"]),
		Port:       strings.TrimSpace(kv["smtp.port"]),
		Username:   strings.TrimSpace(kv["smtp.username"]),
		Password:   strings.TrimSpace(kv["smtp.password"]),
		FromName:   strings.TrimSpace(kv["smtp.fromName"]),
		Encryption: strings.TrimSpace(kv["smtp.encryption"]),
	}
}
