package doom_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-go/api/response"
)

// func TestGetDappAssets_AaveV3(t *testing.T) {
// 	reqData := struct {
// 		Address string `json:"address"`
// 		Chain   string `json:"chain"`
// 		App     string `json:"app"`
// 	}{
// 		Address: "0xab961d7c42bbcd454a54b342bd191a8f090219e6", // 0xab961d7c42bbcd454a54b342bd191a8f090219e6 0x2a1e6b2c51426111cafb32d1957de84190c0182d 0x3ddfa8ec3052539b6c9549f12cea2c295cff5296
// 		Chain:   "eth",
// 		App:     "aave_v3", // aave_v3 uniswap_v2 uniswap_v3
// 	}
// 	respData := struct {
// 		Code         int32  `json:"code"`
// 		Message      string `json:"message"`
// 		DebugMessage string `json:"debugMessage"`
// 		Data         struct {
// 			Assets []struct {
// 				Name         string `json:"name"`
// 				TokenAddress string `json:"tokenAddress"`
// 				TotalValue   string `json:"totalValue"`
// 				Holdings     []struct {
// 					Token  string `json:"token"`
// 					Amount string `json:"amount"`
// 					Price  string `json:"price"`
// 					Value  string `json:"value"`
// 				} `json:"holdings"`
// 			} `json:"assets"`
// 			Debts []struct {
// 				Name         string `json:"name"`
// 				TokenAddress string `json:"tokenAddress"`
// 				TotalValue   string `json:"totalValue"`
// 				Holdings     []struct {
// 					Token  string `json:"token"`
// 					Amount string `json:"amount"`
// 					Price  string `json:"price"`
// 					Value  string `json:"value"`
// 				} `json:"holdings"`
// 			} `json:"debts"`
// 			TotalValue string `json:"totalValue"`
// 		} `json:"data"`
// 	}{}
// 	resp, err := postJsonRequest(_kongDNS+"/doom/getDappAssets/v1", &reqData, nil, &respData, nil)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	if resp.StatusCode != http.StatusOK {
// 		t.Error("not prospective response code")
// 		return
// 	}
// 	if respData.Code != response.StatusCodeOK {
// 		t.Error("not prospective response data code")
// 		return
// 	}
// 	if len(respData.Data.Assets) <= 0 && len(respData.Data.Debts) <= 0 {
// 		t.Error("not prospective response data")
// 		return
// 	}
// }

func TestGetDappAssets_UniswapV2(t *testing.T) {
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
		App     string `json:"app"`
	}{
		Address: "0x3ddfa8ec3052539b6c9549f12cea2c295cff5296", // 0x2a1e6b2c51426111cafb32d1957de84190c0182d
		Chain:   "eth",
		App:     "uniswap_v2", // aave_v3 uniswap_v2 uniswap_v3
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Assets []struct {
				Name         string `json:"name"`
				TokenAddress string `json:"tokenAddress"`
				TotalValue   string `json:"totalValue"`
				Holdings     []struct {
					Token  string `json:"token"`
					Amount string `json:"amount"`
					Price  string `json:"price"`
					Value  string `json:"value"`
				} `json:"holdings"`
			} `json:"assets"`
			Debts []struct {
				Name         string `json:"name"`
				TokenAddress string `json:"tokenAddress"`
				TotalValue   string `json:"totalValue"`
				Holdings     []struct {
					Token  string `json:"token"`
					Amount string `json:"amount"`
					Price  string `json:"price"`
					Value  string `json:"value"`
				} `json:"holdings"`
			} `json:"debts"`
			TotalValue string `json:"totalValue"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getDappAssets/v1", &reqData, nil, &respData, nil)
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
	// if len(respData.Data.Assets) <= 0 && len(respData.Data.Debts) <= 0 {
	// 	t.Error("not prospective response data")
	// 	return
	// }
}

func TestGetDappAssets_UniswapV3(t *testing.T) {
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
		App     string `json:"app"`
	}{
		Address: "0x3ddfa8ec3052539b6c9549f12cea2c295cff5296", // 0x2a1e6b2c51426111cafb32d1957de84190c0182d
		Chain:   "eth",
		App:     "uniswap_v3", // aave_v3 uniswap_v2 uniswap_v3
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Assets []struct {
				Name         string `json:"name"`
				TokenAddress string `json:"tokenAddress"`
				TotalValue   string `json:"totalValue"`
				Holdings     []struct {
					Token  string `json:"token"`
					Amount string `json:"amount"`
					Price  string `json:"price"`
					Value  string `json:"value"`
				} `json:"holdings"`
			} `json:"assets"`
			Debts []struct {
				Name         string `json:"name"`
				TokenAddress string `json:"tokenAddress"`
				TotalValue   string `json:"totalValue"`
				Holdings     []struct {
					Token  string `json:"token"`
					Amount string `json:"amount"`
					Price  string `json:"price"`
					Value  string `json:"value"`
				} `json:"holdings"`
			} `json:"debts"`
			TotalValue string `json:"totalValue"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getDappAssets/v1", &reqData, nil, &respData, nil)
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
	// if len(respData.Data.Assets) <= 0 && len(respData.Data.Debts) <= 0 {
	// 	t.Error("not prospective response data")
	// 	return
	// }
}

func TestGetDappAssets_EmptyParameter(t *testing.T) {
	reqData := struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
		App     string `json:"app"`
	}{
		Address: "3ddfa8ec3052539b6c9549f12cea2c295cff5296",
		Chain:   "eth",
		App:     "",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Assets []struct {
				Name         string `json:"name"`
				TokenAddress string `json:"tokenAddress"`
				TotalValue   string `json:"totalValue"`
				Holdings     []struct {
					Token  string `json:"token"`
					Amount string `json:"amount"`
					Price  string `json:"price"`
					Value  string `json:"value"`
				} `json:"holdings"`
			} `json:"assets"`
			Debts []struct {
				Name         string `json:"name"`
				TokenAddress string `json:"tokenAddress"`
				TotalValue   string `json:"totalValue"`
				Holdings     []struct {
					Token  string `json:"token"`
					Amount string `json:"amount"`
					Price  string `json:"price"`
					Value  string `json:"value"`
				} `json:"holdings"`
			} `json:"debts"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/doom/getDappAssets/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}
