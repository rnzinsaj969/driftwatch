// Package alert provides functionality for sending drift-detection
// notifications to an external webhook endpoint.
//
// Usage:
//
//	sender := alert.New("https://hooks.example.com/drift")
//	err := sender.Send(alert.Payload{
//		FilePath: "/etc/app/config.yaml",
//		Message:  "checksum mismatch detected",
//		Checksum: "d41d8cd98f00b204e9800998ecf8427e",
//	})
//
// The Sender posts a JSON-encoded Payload to the configured URL.
// Any non-2xx HTTP response is treated as an error.
package alert
