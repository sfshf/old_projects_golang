package word_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/word/api/response"
	. "github.com/nextsurfer/word/internal/pkg/model"
)

func TestFavoritedDefinitions(t *testing.T) {
	// mock data
	mockFavorite := FavoriteDefinition{
		DefinitionID: 199999999,
		UserID:       _testAccount.ID,
	}
	if err := _wordGormDB.Create(&mockFavorite).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _wordGormDB.Delete(&FavoriteDefinition{ID: mockFavorite.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Definitions []int64 `json:"definitions"`
		} `json:"data,omitempty"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/word/user/definition/favorites/v1", nil, _testCookie, &respData, nil)
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
	if respData.Data.Definitions[0] != mockFavorite.DefinitionID {
		t.Error("not prospective response data")
		return
	}
}
