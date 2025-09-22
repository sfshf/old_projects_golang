package doom_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
	. "github.com/nextsurfer/doom-go/internal/model"
	"go.mongodb.org/mongo-driver/bson"
)

func TestFavoriteToken(t *testing.T) {
	// mock data
	testSymbol := "ETH"
	// send request -- favorite definition
	reqData := struct {
		Symbol string `json:"symbol"`
	}{
		Symbol: testSymbol,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/favoriteToken/v1", &reqData, _testCookie, &respData, nil)
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
	ctx := context.Background()
	// check db data
	var favoriteToken FavoriteToken
	coll := _mongoDB.Collection(CollectionName_FavoriteToken)
	if err := coll.FindOne(ctx, bson.D{{Key: "symbol", Value: testSymbol}, {Key: "userID", Value: _testAccount.ID}}).
		Decode(&favoriteToken); err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if _, err := coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: favoriteToken.ID}}); err != nil {
			t.Error(err)
			return
		}
	}()
	// send request -- unfavorite definition
	reqData = struct {
		Symbol string `json:"symbol"`
	}{
		Symbol: testSymbol,
	}
	respData = struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	resp, err = postJsonRequest(_kongDNS+"/doom/favoriteToken/v1", &reqData, _testCookie, &respData, nil)
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
	// check db data
	var favoriteToken2 FavoriteToken
	if err := coll.FindOne(ctx, bson.D{
		{Key: "symbol", Value: testSymbol},
		{Key: "userID", Value: _testAccount.ID},
		{Key: "deletedAt", Value: bson.D{{Key: "$gt", Value: 0}}}}).
		Decode(&favoriteToken2); err != nil {
		t.Error(err)
		return
	}
}
