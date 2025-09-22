package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nextsurfer/monitor/internal/common/simplehttp"
)

func handleTransaction(address string, transaction Mempool_GetAddressTransactionsMempool_Transaction) string {
	// check from/to the address
	var isFrom bool
	var fromValue int32
	for _, vin := range transaction.Vin { // from
		if vin.Prevout.ScriptpubkeyAddress == address {
			isFrom = true
		}
		fromValue += vin.Prevout.Value
	}
	var message string
	// var value int32
	if isFrom {
		// var changeValue int32
		// for _, vout := range transaction.Vout {
		// 	if vout.ScriptpubkeyAddress == address {
		// 		changeValue += vout.Value
		// 	}
		// }
		// value = fromValue - changeValue
		message = fmt.Sprintf("从账户地址<%s>转出%d聪BTC", address, fromValue-transaction.Fee)
	} else {
		// for _, vout := range transaction.Vout {
		// 	if vout.ScriptpubkeyAddress == address {
		// 		value += vout.Value
		// 	}
		// }
		message = fmt.Sprintf("向账户地址<%s>转入%d聪BTC", address, fromValue-transaction.Fee)
	}
	return message
}

func handleUSBtcTxsMempoolEntry() error {
	address := "bc1qclyfsxuu8vcwq38yygs5zrskwacq8sjlyvk9mx" // 美国政府持有的比特币
	// 1. mempool request
	rawData, err := Mempool_GetAddressTransactionsMempool(address)
	if err != nil {
		return err
	}
	// 3. if length of list gt 0
	var list []Mempool_GetAddressTransactionsMempool_Transaction
	if err := json.Unmarshal([]byte(rawData), &list); err != nil {
		return err
	}
	if len(list) > 0 {
		for _, item := range list {
			// post messages to telegram
			message := `https://www.blockchain.com/explorer/transactions/btc/` + item.TxID + "\n" // transaction link
			message += handleTransaction(address, item) + "\n"
			log.Println(message)
		}
	}
	return nil
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

func Mempool_GetAddressTransactionsMempool(address string) (string, error) {
	url := "https://mempool.space/api/address/" + address + "/txs/mempool"
	resp, err := simplehttp.Get(url, nil, nil)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", simplehttp.ErrResponseStatusCodeNotEqualTo200
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func main() {
	handleUSBtcTxsMempoolEntry()
}
