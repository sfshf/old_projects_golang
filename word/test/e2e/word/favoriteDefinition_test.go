package word_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/word/api/response"
	. "github.com/nextsurfer/word/internal/pkg/model"
)

func TestFavoriteDefinition(t *testing.T) {
	// mock data
	testDefinitionID := 199999999
	// send request -- favorite definition
	reqData := struct {
		DefinitionID int64 `json:"definitionID"`
	}{
		DefinitionID: int64(testDefinitionID),
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/word/user/definition/favorite/v1", &reqData, _testCookie, &respData, nil)
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
	var favoriteDefinition FavoriteDefinition
	if err := _wordGormDB.Table(TableNameFavoriteDefinition).
		Where("deleted_at = 0").
		Where("definition_id = ?", testDefinitionID).
		Where("user_id = ?", _testAccount.ID).
		First(&favoriteDefinition).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _wordGormDB.Delete(&FavoriteDefinition{ID: favoriteDefinition.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	// send request -- unfavorite definition
	reqData = struct {
		DefinitionID int64 `json:"definitionID"`
	}{
		DefinitionID: int64(testDefinitionID),
	}
	respData = struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	resp, err = postJsonRequest(_kongDNS+"/word/user/definition/favorite/v1", &reqData, _testCookie, &respData, nil)
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
	var favoriteDefinition2 FavoriteDefinition
	if err := _wordGormDB.Table(TableNameFavoriteDefinition).
		Where("deleted_at > 0").
		Where("definition_id = ?", testDefinitionID).
		Where("user_id = ?", _testAccount.ID).
		First(&favoriteDefinition2).Error; err != nil {
		t.Error(err)
		return
	}
}
