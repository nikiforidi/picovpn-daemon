package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Daemon struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	CertPEM []byte `json:"certPem"`
	// KeyPem  []byte `json:"keyPem"`
}

func RegisterSelf(daemon Daemon) {
	b, err := json.Marshal(daemon)
	if err != nil {
		log.Println("failed to marshal daemon:", err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, "https://picovpn.ru/api/daemons", bytes.NewBuffer(b))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("X-Daemon-Key %s", os.Getenv("TELEGRAM_BOT_TOKEN")))

	// resp, err := http.Post("https://picovpn.ru/api/daemons", "application/json", bytes.NewBuffer(b))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("failed to send request:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("failed to register daemon, status code: %d\n", resp.StatusCode)
		return
	}
}
