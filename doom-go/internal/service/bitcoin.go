package service

import (
	"context"
	"net/http"

	doom_api "github.com/nextsurfer/doom-go/api"
	"github.com/nextsurfer/doom-go/api/response"
	"github.com/nextsurfer/doom-go/internal/common/simplehttp"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"go.uber.org/zap"
)

type BitCoinService struct {
	*DoomService
}

func NewBitCoinService(DoomService *DoomService) *BitCoinService {
	return &BitCoinService{
		DoomService: DoomService,
	}
}

type Mempool_GetAddressResponse struct {
	Address      string                                  `json:"address,omitempty"`
	ChainStats   Mempool_GetAddressResponse_ChainStats   `json:"chain_stats,omitempty"`
	MempoolStats Mempool_GetAddressResponse_MempoolStats `json:"mempool_stats,omitempty"`
}

type Mempool_GetAddressResponse_ChainStats struct {
	FundedTxoCount int32 `json:"funded_txo_count,omitempty"`
	FundedTxoSum   int64 `json:"funded_txo_sum,omitempty"`
	SpentTxoCount  int32 `json:"spent_txo_count,omitempty"`
	SpentTxoSum    int64 `json:"spent_txo_sum,omitempty"`
	TxCount        int64 `json:"tx_count,omitempty"`
}

type Mempool_GetAddressResponse_MempoolStats struct {
	FundedTxoCount int32 `json:"funded_txo_count,omitempty"`
	FundedTxoSum   int64 `json:"funded_txo_sum,omitempty"`
	SpentTxoCount  int32 `json:"spent_txo_count,omitempty"`
	SpentTxoSum    int64 `json:"spent_txo_sum,omitempty"`
	TxCount        int32 `json:"tx_count,omitempty"`
}

func Mempool_GetAddress(address string) (*Mempool_GetAddressResponse, error) {
	url := "https://mempool.space/api/address/" + address
	var respData Mempool_GetAddressResponse
	resp, err := simplehttp.Get(url, nil, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	return &respData, nil
}

func (s *BitCoinService) GetAddress(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetAddressRequest) (*doom_api.GetAddressResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	respData, err := Mempool_GetAddress(req.Address)
	if err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
	}
	return &doom_api.GetAddressResponse_Data{
		Address: respData.Address,
		ChainStats: &doom_api.GetAddressResponse_ChainStats{
			FundedTxoCount: respData.ChainStats.FundedTxoCount,
			FundedTxoSum:   respData.ChainStats.FundedTxoSum,
			SpentTxoCount:  respData.ChainStats.SpentTxoCount,
			SpentTxoSum:    respData.ChainStats.SpentTxoSum,
			TxCount:        respData.ChainStats.TxCount,
		},
		MempoolStats: &doom_api.GetAddressResponse_MempoolStats{
			FundedTxoCount: respData.MempoolStats.FundedTxoCount,
			FundedTxoSum:   respData.MempoolStats.FundedTxoSum,
			SpentTxoCount:  respData.MempoolStats.SpentTxoCount,
			SpentTxoSum:    respData.MempoolStats.SpentTxoSum,
			TxCount:        respData.MempoolStats.TxCount,
		},
	}, nil
}

type Mempool_GetAddressTransactions_Transaction struct {
	TxID     string                                 `json:"txid,omitempty"`
	Version  int32                                  `json:"version,omitempty"`
	Locktime int32                                  `json:"locktime,omitempty"`
	Vin      []Mempool_GetAddressTransactions_Vin   `json:"vin,omitempty"`
	Vout     []Mempool_GetAddressTransactions_Vout  `json:"vout,omitempty"`
	Size     int32                                  `json:"size,omitempty"`
	Weight   int32                                  `json:"weight,omitempty"`
	Fee      int32                                  `json:"fee,omitempty"`
	Status   *Mempool_GetAddressTransactions_Status `json:"status,omitempty"`
}

type Mempool_GetAddressTransactions_Vin struct {
	TxID         string                                     `json:"txid,omitempty"`
	Vout         int32                                      `json:"vout,omitempty"`
	Prevout      Mempool_GetAddressTransactions_Vin_Prevout `json:"prevout,omitempty"`
	Scriptsig    string                                     `json:"scriptsig,omitempty"`
	ScriptsigAsm string                                     `json:"scriptsig_asm,omitempty"`
	Witness      []string                                   `json:"witness,omitempty"`
	IsCoinbase   bool                                       `json:"is_coinbase,omitempty"`
	Sequence     int64                                      `json:"sequence,omitempty"`
}

type Mempool_GetAddressTransactions_Vin_Prevout struct {
	Scriptpubkey        string `json:"scriptpubkey,omitempty"`
	ScriptpubkeyAsm     string `json:"scriptpubkey_asm,omitempty"`
	ScriptpubkeyType    string `json:"scriptpubkey_type,omitempty"`
	ScriptpubkeyAddress string `json:"scriptpubkey_address,omitempty"`
	Value               int32  `json:"value,omitempty"`
}

type Mempool_GetAddressTransactions_Vout struct {
	Scriptpubkey        string `json:"scriptpubkey,omitempty"`
	ScriptpubkeyAsm     string `json:"scriptpubkey_asm,omitempty"`
	ScriptpubkeyType    string `json:"scriptpubkey_type,omitempty"`
	ScriptpubkeyAddress string `json:"scriptpubkey_address,omitempty"`
	Value               int32  `json:"value,omitempty"`
}

type Mempool_GetAddressTransactions_Status struct {
	Confirmed   bool   `json:"confirmed,omitempty"`
	BlockHeight int32  `json:"block_height,omitempty"`
	BlockHash   string `json:"block_hash,omitempty"`
	BlockTime   int64  `json:"block_time,omitempty"`
}

func Mempool_GetAddressTransactions(address string) ([]Mempool_GetAddressTransactions_Transaction, error) {
	url := "https://mempool.space/api/address/" + address + "/txs"
	var respData []Mempool_GetAddressTransactions_Transaction
	resp, err := simplehttp.Get(url, nil, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	return respData, nil
}

func (s *BitCoinService) GetAddressTransactions(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetAddressTransactionsRequest) (*doom_api.GetAddressTransactionsResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	respData, err := Mempool_GetAddressTransactions(req.Address)
	if err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
	}
	var list []*doom_api.GetAddressTransactionsResponse_Transaction
	for _, item := range respData {
		aTransaction := &doom_api.GetAddressTransactionsResponse_Transaction{
			TxID:     item.TxID,
			Version:  item.Version,
			Locktime: item.Locktime,
			Size:     item.Size,
			Weight:   item.Weight,
			Fee:      item.Fee,
			Status: &doom_api.GetAddressTransactionsResponse_Status{
				Confirmed:   item.Status.Confirmed,
				BlockHeight: item.Status.BlockHeight,
				BlockHash:   item.Status.BlockHash,
				BlockTime:   item.Status.BlockTime,
			},
		}
		// vin
		var vinList []*doom_api.GetAddressTransactionsResponse_Vin
		for _, vin := range item.Vin {
			vinList = append(vinList, &doom_api.GetAddressTransactionsResponse_Vin{
				TxID: vin.TxID,
				Vout: vin.Vout,
				Prevout: &doom_api.GetAddressTransactionsResponse_Prevout{
					Scriptpubkey:        vin.Prevout.Scriptpubkey,
					ScriptpubkeyAsm:     vin.Prevout.ScriptpubkeyAsm,
					ScriptpubkeyType:    vin.Prevout.ScriptpubkeyType,
					ScriptpubkeyAddress: vin.Prevout.ScriptpubkeyAddress,
					Value:               vin.Prevout.Value,
				},
				Scriptsig:    vin.Scriptsig,
				ScriptsigAsm: vin.ScriptsigAsm,
				Witness:      vin.Witness,
				IsCoinbase:   vin.IsCoinbase,
				Sequence:     vin.Sequence,
			})
		}
		aTransaction.Vin = vinList
		// vout
		var voutList []*doom_api.GetAddressTransactionsResponse_Vout
		for _, vout := range item.Vout {
			voutList = append(voutList, &doom_api.GetAddressTransactionsResponse_Vout{
				Scriptpubkey:        vout.Scriptpubkey,
				ScriptpubkeyAsm:     vout.ScriptpubkeyAsm,
				ScriptpubkeyType:    vout.ScriptpubkeyType,
				ScriptpubkeyAddress: vout.ScriptpubkeyAddress,
				Value:               vout.Value,
			})
		}
		aTransaction.Vout = voutList

		list = append(list, aTransaction)
	}
	return &doom_api.GetAddressTransactionsResponse_Data{List: list}, nil
}

type Mempool_GetAddressTransactionsMempool_Transaction struct {
	TxID     string                                       `json:"txid,omitempty"`
	Version  int32                                        `json:"version,omitempty"`
	Locktime int32                                        `json:"locktime,omitempty"`
	Vin      []Mempool_GetAddressTransactionsMempool_Vin  `json:"vin,omitempty"`
	Vout     []Mempool_GetAddressTransactionsMempool_Vout `json:"vout,omitempty"`
	Size     int32                                        `json:"size,omitempty"`
	Weight   int32                                        `json:"weight,omitempty"`
	Fee      int32                                        `json:"fee,omitempty"`
	Status   Mempool_GetAddressTransactionsMempool_Status `json:"status,omitempty"`
}

type Mempool_GetAddressTransactionsMempool_Vin struct {
	TxID         string                                            `json:"txid,omitempty"`
	Vout         int32                                             `json:"vout,omitempty"`
	Prevout      Mempool_GetAddressTransactionsMempool_Vin_Prevout `json:"prevout,omitempty"`
	Scriptsig    string                                            `json:"scriptsig,omitempty"`
	ScriptsigAsm string                                            `json:"scriptsig_asm,omitempty"`
	Witness      []string                                          `json:"witness,omitempty"`
	IsCoinbase   bool                                              `json:"is_coinbase,omitempty"`
	Sequence     int64                                             `json:"sequence,omitempty"`
}

type Mempool_GetAddressTransactionsMempool_Vin_Prevout struct {
	Scriptpubkey        string `json:"scriptpubkey,omitempty"`
	ScriptpubkeyAsm     string `json:"scriptpubkey_asm,omitempty"`
	ScriptpubkeyType    string `json:"scriptpubkey_type,omitempty"`
	ScriptpubkeyAddress string `json:"scriptpubkey_address,omitempty"`
	Value               int32  `json:"value,omitempty"`
}

type Mempool_GetAddressTransactionsMempool_Vout struct {
	Scriptpubkey        string `json:"scriptpubkey,omitempty"`
	ScriptpubkeyAsm     string `json:"scriptpubkey_asm,omitempty"`
	ScriptpubkeyType    string `json:"scriptpubkey_type,omitempty"`
	ScriptpubkeyAddress string `json:"scriptpubkey_address,omitempty"`
	Value               int32  `json:"value,omitempty"`
}

type Mempool_GetAddressTransactionsMempool_Status struct {
	Confirmed   bool   `json:"confirmed,omitempty"`
	BlockHeight int32  `json:"block_height,omitempty"`
	BlockHash   string `json:"block_hash,omitempty"`
	BlockTime   int64  `json:"block_time,omitempty"`
}

func Mempool_GetAddressTransactionsMempool(address string) ([]Mempool_GetAddressTransactionsMempool_Transaction, error) {
	url := "https://mempool.space/api/address/" + address + "/txs/mempool"
	var respData []Mempool_GetAddressTransactionsMempool_Transaction
	resp, err := simplehttp.Get(url, nil, &respData)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	return respData, nil
}

func (s *BitCoinService) GetAddressTransactionsMempool(ctx context.Context, rpcCtx *rpc.Context, req *doom_api.GetAddressTransactionsMempoolRequest) (*doom_api.GetAddressTransactionsMempoolResponse_Data, *gerror.AppError) {
	logger := s.Logger
	if rpcCtx != nil {
		logger = rpcCtx.Logger
	}
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	respData, err := Mempool_GetAddressTransactionsMempool(req.Address)
	if err != nil {
		logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeBadRequest)
	}
	var list []*doom_api.GetAddressTransactionsMempoolResponse_Transaction
	for _, item := range respData {
		aTransaction := &doom_api.GetAddressTransactionsMempoolResponse_Transaction{
			TxID:     item.TxID,
			Version:  item.Version,
			Locktime: item.Locktime,
			Size:     item.Size,
			Weight:   item.Weight,
			Fee:      item.Fee,
			Status: &doom_api.GetAddressTransactionsMempoolResponse_Status{
				Confirmed:   item.Status.Confirmed,
				BlockHeight: item.Status.BlockHeight,
				BlockHash:   item.Status.BlockHash,
				BlockTime:   item.Status.BlockTime,
			},
		}
		// vin
		var vinList []*doom_api.GetAddressTransactionsMempoolResponse_Vin
		for _, vin := range item.Vin {
			vinList = append(vinList, &doom_api.GetAddressTransactionsMempoolResponse_Vin{
				TxID: vin.TxID,
				Vout: vin.Vout,
				Prevout: &doom_api.GetAddressTransactionsMempoolResponse_Prevout{
					Scriptpubkey:        vin.Prevout.Scriptpubkey,
					ScriptpubkeyAsm:     vin.Prevout.ScriptpubkeyAsm,
					ScriptpubkeyType:    vin.Prevout.ScriptpubkeyType,
					ScriptpubkeyAddress: vin.Prevout.ScriptpubkeyAddress,
					Value:               vin.Prevout.Value,
				},
				Scriptsig:    vin.Scriptsig,
				ScriptsigAsm: vin.ScriptsigAsm,
				Witness:      vin.Witness,
				IsCoinbase:   vin.IsCoinbase,
				Sequence:     vin.Sequence,
			})
		}
		aTransaction.Vin = vinList
		// vout
		var voutList []*doom_api.GetAddressTransactionsMempoolResponse_Vout
		for _, vout := range item.Vout {
			voutList = append(voutList, &doom_api.GetAddressTransactionsMempoolResponse_Vout{
				Scriptpubkey:        vout.Scriptpubkey,
				ScriptpubkeyAsm:     vout.ScriptpubkeyAsm,
				ScriptpubkeyType:    vout.ScriptpubkeyType,
				ScriptpubkeyAddress: vout.ScriptpubkeyAddress,
				Value:               vout.Value,
			})
		}
		aTransaction.Vout = voutList

		list = append(list, aTransaction)
	}
	return &doom_api.GetAddressTransactionsMempoolResponse_Data{List: list}, nil
}
