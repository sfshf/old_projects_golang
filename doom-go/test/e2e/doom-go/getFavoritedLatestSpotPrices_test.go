package doom_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/doom-go/api/response"
	. "github.com/nextsurfer/doom-go/internal/model"
	"go.mongodb.org/mongo-driver/bson"
)

func TestGetFavoritedLatestSpotPrices(t *testing.T) {
	ctx := context.Background()
	// mock data
	ts := time.Now().UnixMilli()
	mockFavorite1 := FavoriteToken{
		CreatedAt: ts,
		UpdatedAt: ts,
		Symbol:    "btc",
		UserID:    _testAccount.ID,
	}
	coll := _mongoDB.Collection(CollectionName_FavoriteToken)
	result, err := coll.InsertOne(ctx, &mockFavorite1)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if _, err := coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: result.InsertedID}}); err != nil {
			t.Error(err)
			return
		}
	}()
	mockFavorite2 := FavoriteToken{
		CreatedAt: ts,
		UpdatedAt: ts,
		Symbol:    "eth",
		UserID:    _testAccount.ID,
	}
	result2, err := coll.InsertOne(ctx, &mockFavorite2)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if _, err := coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: result2.InsertedID}}); err != nil {
			t.Error(err)
			return
		}
	}()
	mockFavorite3 := FavoriteToken{
		CreatedAt: ts,
		UpdatedAt: ts,
		Symbol:    "weth",
		UserID:    _testAccount.ID,
	}
	result3, err := coll.InsertOne(ctx, &mockFavorite3)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if _, err := coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: result3.InsertedID}}); err != nil {
			t.Error(err)
			return
		}
	}()
	reqData := struct {
		Symbols  []string `json:"symbols"`
		BaseCoin string   `json:"baseCoin"`
	}{
		Symbols:  []string{"btc", "weth", "eth"},
		BaseCoin: "USDT",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []string `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getFavoritedLatestSpotPrices/v1", &reqData, _testCookie, &respData, nil)
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
	if len(respData.Data.List) <= 0 {
		t.Error("not prospective response data")
		return
	}
}
