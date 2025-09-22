package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nextsurfer/doom-go/internal/common/simplehttp"
)

// gecko ------------------------------------------------------------------------------------

type GeckoCoin struct {
	ID        string `json:"id"`
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Platforms struct {
		Ethereum string `json:"ethereum"`
	} `json:"platforms"`
}

func fetchReputableTokensFromCoinGecko(pageSize int) ([]GeckoCoin, error) {
	// fetch from remote
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1", pageSize)
	var respData []GeckoCoin
	resp, err := simplehttp.Get(url, map[string]string{"User-Agent": "curl/8.5.0"}, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	return respData, nil
}

func fetchSpecialFromCoinGecko(ids ...string) ([]GeckoCoin, error) {
	// fetch from remote
	url := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&ids=" + strings.Join(ids, ",")
	var respData []GeckoCoin
	resp, err := simplehttp.Get(url, map[string]string{"User-Agent": "curl/8.5.0"}, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	return respData, nil
}

func fetchImageData(imageUrl string) ([]byte, error) {
	resp, err := simplehttp.Get(imageUrl, map[string]string{"User-Agent": "curl/8.5.0"}, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func reputableTokenImages() {
	// first, fetch reputable coins from coingecko
	reputableTokens, err := fetchReputableTokensFromCoinGecko(200)
	if err != nil {
		log.Println(err)
		return
	}
	specials, err := fetchSpecialFromCoinGecko("weth")
	if err != nil {
		log.Println(err)
		return
	}
	// get special coins
	for _, coin := range specials {
		var has bool
		for _, item := range reputableTokens {
			if item.ID == coin.ID {
				has = true
			}
		}
		if !has {
			reputableTokens = append(reputableTokens, coin)
		}
	}
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	imagefiles := make(map[string]bool, len(reputableTokens)*2)
	for _, token := range reputableTokens {
		// large image
		log.Println("fetch large image:", token.Image)
		largeFile := strings.ToUpper(token.Symbol) + "_large.png"
		if !imagefiles[largeFile] {
			imagefiles[largeFile] = true
			data, err := fetchImageData(token.Image)
			if err != nil {
				log.Printf("fetch large image of %s error: %v\n", token.Symbol, err)
			} else {
				f, err := w.Create(largeFile)
				if err != nil {
					log.Printf("create largeFile %s error: %v\n", largeFile, err)
					return
				}
				if _, err := f.Write(data); err != nil {
					log.Printf("write largeFile %s error: %v\n", largeFile, err)
					return
				}
			}
		}

		// small image
		smallImage := strings.Replace(token.Image, "large", "small", 1)
		log.Println("fetch small image:", smallImage)
		smallFile := strings.ToUpper(token.Symbol) + "_small.png"
		if !imagefiles[smallFile] {
			imagefiles[smallFile] = true
			data, err := fetchImageData(smallImage)
			if err != nil {
				log.Printf("fetch small image of %s error: %v\n", token.Symbol, err)
			} else {
				f, err := w.Create(smallFile)
				if err != nil {
					log.Printf("create smallFile %s error: %v\n", smallFile, err)
					return
				}
				if _, err := f.Write(data); err != nil {
					log.Printf("write smallFile %s error: %v\n", smallFile, err)
					return
				}
			}
		}
	}
	if err := w.Close(); err != nil {
		log.Printf("zip writer close error: %v\n", err)
		return
	}
	if err := os.MkdirAll("tmp", 0750); err != nil {
		log.Printf("os.MkdirAll error: %v\n", err)
		return
	}
	if err := os.WriteFile("tmp/reputable_token_images.zip", buf.Bytes(), 0660); err != nil {
		log.Printf("os.WriteFile error: %v\n", err)
		return
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	start := time.Now()
	reputableTokenImages()
	end := time.Now()
	log.Printf("end running: %v, duration: %s\n", end, end.Sub(start).String())
}
