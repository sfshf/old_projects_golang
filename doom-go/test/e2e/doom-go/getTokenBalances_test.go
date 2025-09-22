package doom_test

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
	. "github.com/nextsurfer/doom-go/internal/model"
	"go.mongodb.org/mongo-driver/bson"
)

func TestGetTokenBalances(t *testing.T) {
	ctx := context.Background()
	testUserAddress := "3DdfA8eC3052539b6C9549F12cEA2C295cfF5296" // 36cc7B13029B5DEe4034745FB4F24034f3F2ffc6
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
	}{
		Address: testUserAddress,
		Chain:   "eth",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			ToBlock string `json:"toBlock"`
			Tokens  []struct {
				Balance string `json:"balance"`
				Address string `json:"address"`
				Name    string `json:"name"`
				Symbol  string `json:"symbol"`
				Price   string `json:"price"`
				Value   string `json:"value"`
			} `json:"tokens"`
			UnknownTokens []struct {
				Balance string `json:"balance"`
				Address string `json:"address"`
			} `json:"unknownTokens"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getTokenBalances/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	// check data
	if len(respData.Data.Tokens) <= 0 {
		t.Error("not prospective response data")
		return
	}
	var userERC20Tokens UserERC20Tokens
	err = _mongoDB.Collection(CollectionName_UserERC20Tokens).FindOne(ctx, bson.D{{Key: "key", Value: "0x" + testUserAddress}}).Decode(&userERC20Tokens)
	if err != nil {
		t.Error(err)
		return
	}
	for _, item := range userERC20Tokens.Value.Tokens {
		if item.Type == TokenTypeERC20 {
			var has bool
			for _, token := range respData.Data.Tokens {
				if strings.EqualFold(token.Address, item.Address) {
					if item.Balance != "" && token.Balance != "" {
						balance1, err := strconv.ParseFloat(token.Balance, 64)
						if err != nil {
							t.Error(err)
							return
						}
						balance2, err := strconv.ParseFloat(item.Balance, 64)
						if err != nil {
							t.Error(err)
							return
						}
						if balance1 == balance2 {
							has = true
						}
					} else if item.Balance == "" && token.Balance == "" {
						has = true
					}
				}
			}
			if !has {
				t.Error("invalid token data")
				return
			}
		}
	}
}
