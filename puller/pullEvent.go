package puller

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ttblack/Elastos.ELA.Inscription/store"
)

func (p *Puller) PullCrossBtcTxLogs(tx *types.Transaction, txIndex uint32, block *types.Block) ([]*store.InscribeTx, error) {
	receipt, err := p.client.TransactionReceipt(tx.Hash())
	if err != nil {
		fmt.Println("TransactionReceipt failed", "error", err)
		return nil, err
	}
	str := "[{\n      \"anonymous\": false,\n      \"inputs\": [\n        {\n          \"indexed\": false,\n          \"internalType\": \"address\",\n          \"name\": \"from\",\n          \"type\": \"address\"\n        },\n        {\n          \"indexed\": false,\n          \"internalType\": \"string\",\n          \"name\": \"to\",\n          \"type\": \"string\"\n        },\n        {\n          \"indexed\": false,\n          \"internalType\": \"string\",\n          \"name\": \"ticker\",\n          \"type\": \"string\"\n        },\n        {\n          \"indexed\": false,\n          \"internalType\": \"uint256\",\n          \"name\": \"amount\",\n          \"type\": \"uint256\"\n        }\n      ],\n      \"name\": \"avascriptions_protocol_CrossASC20Token\",\n      \"type\": \"event\"\n    }]"
	a, err := abi.JSON(strings.NewReader(str))
	if err != nil {
		return nil, err
	}
	type transfer struct {
		From   common.Address `json:"from"`
		To     string         `json:"to"`
		Ticker string         `json:"ticker"`
		Amount *big.Int       `json:"amount"`
	}
	trans := transfer{}
	txs := make([]*store.InscribeTx, 0)
	for _, log := range receipt.Logs {
		if p.crossSignature.Cmp(log.Topics[0]) == 0 {
			err = a.UnpackIntoInterface(&trans, "avascriptions_protocol_CrossASC20Token", log.Data)
			if err != nil {
				fmt.Println("is error log data", log.TxHash.String())
				continue
			}

			inscriptionTx := &store.InscribeTx{
				TxId:      tx.Hash().String(),
				From:      trans.From.String(),
				To:        trans.To,
				Block:     block.NumberU64(),
				Idx:       txIndex,
				BlockTime: block.Time(),
				Input:     "",
			}
			txs = append(txs, inscriptionTx)
		}
	}
	return txs, nil

}
