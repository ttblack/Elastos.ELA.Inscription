package store

import (
	"encoding/json"
	"errors"

	"github.com/ttblack/Elastos.ELA.Inscription/constant"
)

type InscribeTx struct {
	TxId      string
	From      string
	To        string
	Block     uint64
	Idx       uint32
	BlockTime uint64
	Input     string
}

type Inscribe struct {
	Proto     string `json:"p,omitempty"`
	Operation string `json:"op,omitempty"`
	Tick      string `json:"tick,omitempty"` //There is a limit on the number of names, usually 4 letters
	Max       string `json:"max,omitempty"`
	Amount    string `json:"amt,omitempty"`
	Limit     string `json:"lim,omitempty"` // option
	Decimal   string `json:"dec,omitempty"` // option
}

func (body *Inscribe) Unmarshal(contentBody []byte) (err error) {
	var bodyMap = make(map[string]interface{})
	if err := json.Unmarshal(contentBody, &bodyMap); err != nil {
		return err
	}
	if v, ok := bodyMap["p"].(string); ok {
		body.Proto = v
	} else {
		return errors.New("not have Proto")
	}

	if body.Proto != constant.ASC20_P {
		return errors.New("protocol is not asc20")
	}

	if v, ok := bodyMap["op"].(string); ok {
		body.Operation = v
	} else {
		return errors.New("not have operation")
	}

	if v, ok := bodyMap["tick"].(string); ok {
		body.Tick = v
	} else {
		return errors.New("not have tick name")
	}

	if v, ok := bodyMap["max"].(string); ok {
		body.Max = v
	}
	if v, ok := bodyMap["amt"].(string); ok {
		body.Amount = v
	}

	if _, ok := bodyMap["lim"]; !ok {
		body.Limit = body.Max
	} else {
		if v, ok := bodyMap["lim"].(string); ok {
			body.Limit = v
		}
	}

	if _, ok := bodyMap["dec"]; !ok {
		body.Decimal = constant.DEFAULT_DECIMAL_18
	} else {
		if v, ok := bodyMap["dec"].(string); ok {
			body.Decimal = v
		}
	}

	return nil
}
