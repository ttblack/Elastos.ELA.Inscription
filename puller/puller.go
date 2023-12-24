package puller

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ttblack/Elastos.ELA.Inscription/constant"
	"github.com/ttblack/Elastos.ELA.Inscription/store"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ttblack/Elastos.ELA.Inscription/rpc"
)

const crossBtcEvent = "avascriptions_protocol_CrossASC20Token(address,string,string,uint256)"

type Puller struct {
	client         *rpc.Client
	db             *store.LevelDBStorage
	crossSignature common.Hash
}

func NewPuller(rpcURL string, db *store.LevelDBStorage) (*Puller, error) {
	cli, err := rpc.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	crossSignature := crypto.Keccak256Hash([]byte(crossBtcEvent))
	puller := &Puller{
		client:         cli,
		db:             db,
		crossSignature: crossSignature,
	}
	return puller, nil
}

func (p *Puller) LatestBlock() (uint64, error) {
	latestBlock, err := p.client.LatestBlock()
	if err != nil {
		return 0, err
	}
	return latestBlock.Uint64(), nil
}

func (p *Puller) Start(startBlock uint64) (finishedBlock uint64, err error) {
	latestBlock, err := p.client.LatestBlock()
	if err != nil {
		return 0, err
	}
	ctx := context.Background()
	var block *types.Block
	finishedBlock = latestBlock.Uint64()
	for i := startBlock; i <= finishedBlock; i++ {
		block, err = p.client.BlockByNumber(ctx, big.NewInt(0).SetUint64(i))
		if err != nil {
			return finishedBlock - 1, err
		}
		p.processBlockTransactions(block)
	}
	return finishedBlock, nil
}

func (p *Puller) processBlockTransactions(block *types.Block) {
	fmt.Println("processBlockTransactions height", block.NumberU64())
	txs := block.Transactions()
	elements := make([]*store.InscribeTx, 0)
	crossBtcTxs := make([]*store.InscribeTx, 0)
	for i, tx := range txs {
		fmt.Println("index ", i, "tx ", tx.Hash().String(), " input", string(tx.Data()))
		if isAsc20Format(tx.Data()) {
			inscribeTx, err := p.processJsonData(tx, block, uint32(i))
			if err != nil {
				fmt.Println("process Json data failed", "error", err)
			}
			elements = append(elements, inscribeTx)
		} else {
			fmt.Println("is not asc20 format")
		}
		crossTxs, err := p.PullCrossBtcTxLogs(tx, uint32(i), block)
		if err != nil {
			fmt.Println("PullCrossBtcTxLogs failed", "error", err)
		} else {
			crossBtcTxs = append(crossBtcTxs, crossTxs...)
		}
	}
	if len(elements) > 0 {
		err := p.db.AddInscribeTx(block.NumberU64(), elements)
		if err != nil {
			fmt.Println("AddInscribeTx failed", "error", err)
		}
	}

	if len(crossBtcTxs) > 0 {
		err := p.db.AddCrossBtcTx(block.NumberU64(), crossBtcTxs)
		if err != nil {
			fmt.Println("AddCrossBtcTx failed", "error", err)
		}
	}
}

func (p *Puller) processJsonData(tx *types.Transaction, block *types.Block, txIdx uint32) (*store.InscribeTx, error) {
	inscription := new(store.Inscribe)
	content := bytes.TrimSpace(tx.Data())
	content = content[6:]
	if err := inscription.Unmarshal(content); err != nil {
		return nil, err
	}

	var from = ""
	if inscription.Operation != constant.OP_MINT {
		var signer types.Signer = types.HomesteadSigner{}
		if tx.Protected() {
			signer = types.NewEIP155Signer(tx.ChainId())
		}
		pfrom, err := types.Sender(signer, tx)
		if err != nil {
			return nil, err
		}
		from = pfrom.String()
	}
	inscriptionTx := &store.InscribeTx{
		TxId:      tx.Hash().String(),
		From:      from,
		To:        tx.To().String(),
		Block:     block.NumberU64(),
		Idx:       txIdx,
		BlockTime: block.Time(),
		Input:     hex.EncodeToString(tx.Data()),
	}

	if inscription.Operation == constant.OP_DEPLOY {
		err := p.db.AddDeployInscription(inscription)
		if err != nil {
			fmt.Println("AddDeployInscription error ", err)
			os.Exit(1)
			return nil, err
		}
	} else {
		//if inscription.Operation == constant.OP_MINT {
		//
		//} else if inscription.Operation == constant.OP_TRANSFER {
		//
		//} else if inscription.Operation == constant.OP_LIST {
		//
		//}
	}
	return inscriptionTx, nil
}

func isAsc20Format(input []byte) bool {
	if len(input) < 40 {
		return false
	}
	content := bytes.TrimSpace(input)
	if !bytes.HasPrefix(content, []byte("data:,{")) {
		return false
	}
	if !bytes.HasSuffix(content, []byte("}")) {
		return false
	}
	return true
}

func isJson(contentBody []byte) bool {
	if len(contentBody) < 40 {
		return false
	}
	content := bytes.TrimSpace(contentBody)
	if !bytes.HasPrefix(content, []byte("{")) {
		return false
	}
	if !bytes.HasSuffix(content, []byte("}")) {
		return false
	}
	return true
}
