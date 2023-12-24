package rpc

import (
	"fmt"
	"math/big"
)

type ServerErrCode int

const (
	Error   ServerErrCode = -1
	Success ServerErrCode = 0
)

func getInscriptions(param Params) map[string]interface{} {
	inscriptions, err := datadb.GetDeployedInscription()
	fmt.Println("getInscriptions ", inscriptions, "error ", err)
	if err != nil {
		return map[string]interface{}{"Result": "failed", "data": inscriptions, "Error": Error}
	}
	return map[string]interface{}{"Result": "suc", "Data": inscriptions, "Error": Success}
}

func getTick(param Params) map[string]interface{} {
	tick, ok := param.String("tick")
	if !ok {
		return map[string]interface{}{"Result": "params error", "Data": nil, "Error": Error}
	}
	inscribe, err := datadb.GetInscribeInfo(tick)
	if err != nil {
		return map[string]interface{}{"Result": "failed", "Data": inscribe, "Error": Error}
	}
	return map[string]interface{}{"Result": "suc", "Data": inscribe, "Error": Success}
}

func getInscribeTxByHeight(param Params) map[string]interface{} {
	fmt.Println("getInscribeTxByHeight", param)
	number, ok := param.String("height")
	fmt.Println("getInscribeTxByHeight", number, "oik ", ok)
	if !ok {
		return map[string]interface{}{"Result": "no height", "Data": nil, "Error": Error}
	}
	bheight, ok := big.NewInt(0).SetString(number, 10)
	if !ok {
		return map[string]interface{}{"Result": "height is error " + number, "Data": nil, "Error": Error}
	}
	tx, err := datadb.GetInscribeTxs(bheight.Uint64())
	if err != nil {
		return map[string]interface{}{"Result": err.Error(), "Data": tx, "Error": Error}
	}
	return map[string]interface{}{"Result": "suc", "Data": tx, "Error": Success}
}

func getCrossBtcTxsByHeight(param Params) map[string]interface{} {
	fmt.Println("getCrossBtcTxsByHeight", param)
	number, ok := param.String("height")
	if !ok {
		return map[string]interface{}{"Result": "no height", "Data": nil, "Error": Error}
	}
	bheight, ok := big.NewInt(0).SetString(number, 10)
	if !ok {
		return map[string]interface{}{"Result": "height is error " + number, "Data": nil, "Error": Error}
	}
	tx, err := datadb.GetInscribeCrossTxs(bheight.Uint64())
	if err != nil {
		return map[string]interface{}{"Result": err.Error(), "Data": tx, "Error": Error}
	}
	return map[string]interface{}{"Result": "suc", "Data": tx, "Error": Success}
}
