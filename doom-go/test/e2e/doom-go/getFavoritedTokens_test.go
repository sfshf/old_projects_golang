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

func TestGetFavoritedTokens(t *testing.T) {
	ctx := context.Background()
	ts := time.Now().UnixMilli()
	// mock data
	mockFavorite := FavoriteToken{
		CreatedAt: ts,
		UpdatedAt: ts,
		Symbol:    "doom-e2e-test-TestGetFavoritedTokens",
		UserID:    _testAccount.ID,
	}
	coll := _mongoDB.Collection(CollectionName_FavoriteToken)
	result, err := coll.InsertOne(ctx, &mockFavorite)
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
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []string `json:"list"`
		} `json:"data,omitempty"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getFavoritedTokens/v1", nil, _testCookie, &respData, nil)
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
	if respData.Data.List[0] != mockFavorite.Symbol {
		t.Error("not prospective response data")
		return
	}
}
