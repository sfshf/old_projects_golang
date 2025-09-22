package main

import (
	"log"
	"net/http"

	"github.com/nextsurfer/monitor/internal/common/simplehttp"
)

func main() {
	message := `https://www.blockchain.com/explorer/transactions/btc/5f248c87957fb2ebdd0c2d51f4ec23ec3f7871d0bd48bef481c73621316d7124`
	resp, err := simplehttp.PostJsonRequest(
		`https://api.telegram.org/bot1678806156:AAE8cWdlygrGCHWmHElQHNJ0ZjOv1IRQGeg/sendMessage`,
		map[string]string{
			"Accept":     "*/*",
			"Host":       "api.telegram.org",
			"User-Agent": `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.2 Safari/605.1.15`,
		},
		struct {
			ChatID string `json:"chat_id"`
			Text   string `json:"text"`
		}{
			ChatID: "1417969737",
			Text:   message,
		},
		nil,
		nil,
	)
	if err != nil {
		log.Fatalln(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalln(simplehttp.ErrResponseStatusCodeNotEqualTo200)
	}
}
