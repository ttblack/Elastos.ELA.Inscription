package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/rs/cors"
	"github.com/ttblack/Elastos.ELA.Inscription/store"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

// an instance of the multiplexer
var rpcMethod map[string]func(Params) map[string]interface{}

var datadb *store.LevelDBStorage

const (
	// JSON-RPC protocol error codes.
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
	//-32000 to -32099	Server error, waiting for defining

	// IOTimeout is the maximum duration for JSON-RPC reading or writing
	// timeout.
	IOTimeout = 60 * time.Second

	// MaxRPCRead is the maximum buffer size for reading request.
	MaxRPCRead = 1024 * 1024 * 8
)

func StartRPCServer(httpPort int, db *store.LevelDBStorage) {
	datadb = db
	rpcMethod = make(map[string]func(Params) map[string]interface{})

	rpcMethod["inscriptions"] = getInscriptions
	rpcMethod["getTick"] = getTick
	rpcMethod["getInscribeTxByHeight"] = getInscribeTxByHeight
	rpcMethod["getCrossBtcTxsByHeight"] = getCrossBtcTxsByHeight

	var handler http.Handler
	rpcServeMux := http.NewServeMux()
	c := cors.New(cors.Options{})
	handler = c.Handler(rpcServeMux)

	server := http.Server{
		Handler:      handler,
		ReadTimeout:  IOTimeout,
		WriteTimeout: IOTimeout,
	}
	rpcServeMux.HandleFunc("/", Handle)
	l, err := net.Listen("tcp4", ":"+strconv.Itoa(httpPort))
	if err != nil {
		fmt.Println("Create listener error: ", err.Error())
		return
	}
	err = server.Serve(l)
	if err != nil {
		fmt.Println("ListenAndServe error: ", err.Error())
		os.Exit(0)
	}
}

func Handle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handle handle hanle", r.Proto)
	isClientAllowed := clientAllowed(r)
	if !isClientAllowed {
		RPCError(w, http.StatusForbidden, "Client ip is not allowed")
		return
	}

	// JSON RPC commands should be POSTs
	if r.Method != "POST" {
		RPCError(w, http.StatusMethodNotAllowed, "JSON-RPC protocol only allows POST method")
		return
	}
	contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if contentType != "application/json" && contentType != "text/plain" {
		RPCError(w, http.StatusUnsupportedMediaType, "JSON-RPC need content type to be application/json or text/plain")
		return
	}

	isCheckAuthOk := checkAuth(r)
	if !isCheckAuthOk {
		RPCError(w, http.StatusUnauthorized, "Client authenticate failed")
		return
	}

	//read the body of the request
	body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, MaxRPCRead))
	if err != nil {
		RPCError(w, http.StatusBadRequest, "JSON-RPC request reading error:"+err.Error())
		return
	}

	var requestArray []Request
	var request Request

	err = json.Unmarshal(body, &request)
	if err != nil {
		errArray := json.Unmarshal(body, &requestArray)
		if errArray != nil {
			RPCError(w, http.StatusBadRequest, "JSON-RPC request parsing error:"+err.Error())
			return
		}
	}
	var data []byte
	if len(requestArray) == 0 {
		response := getResponse(request)
		data, _ = json.Marshal(response)
	} else {
		var responseArray []Response
		for _, req := range requestArray {
			response := getResponse(req)
			responseArray = append(responseArray, response)
		}

		data, _ = json.Marshal(responseArray)
	}

	w.Header().Set("Content-type", "application/json")
	w.Write(data)
}

func getResponse(request Request) Response {
	var resp Response
	requestMethod := request.Method
	if len(requestMethod) == 0 {
		resp = Response{
			JSONRPC: "2.0",
			Result:  nil,
			ID:      request.ID,
			Error: map[string]interface{}{
				"id":      request.ID,
				"code":    InvalidRequest,
				"message": "JSON-RPC need a method",
			},
		}
		return resp
	}
	method, ok := rpcMethod[requestMethod]
	if !ok {
		resp = Response{
			JSONRPC: "2.0",
			Result:  nil,
			ID:      request.ID,
			Error: map[string]interface{}{
				"id":      request.ID,
				"code":    MethodNotFound,
				"message": "JSON-RPC method " + requestMethod + " not found",
			},
		}
		return resp
	}

	requestParams := request.Params
	// Json rpc 1.0 support positional parameters while json rpc 2.0 support named parameters.
	// positional parameters: { "requestParams":[1, 2, 3....] }
	// named parameters: { "requestParams":{ "a":1, "b":2, "c":3 } }
	// Here we support both of them.
	var params Params
	switch requestParams := requestParams.(type) {
	case nil:
	case []interface{}:
		params = convertParams(requestMethod, requestParams)
	case map[string]interface{}:
		params = Params(requestParams)
	default:
		resp = Response{
			JSONRPC: "2.0",
			Result:  nil,
			ID:      request.ID,
			Error: map[string]interface{}{
				"id":      request.ID,
				"code":    InvalidRequest,
				"message": "params format error, must be an array or a map",
			},
		}
		return resp
	}

	response := method(params)
	resp = Response{
		JSONRPC: "2.0",
		Result:  response["Data"],
		ID:      request.ID,
		Error:   response["Result"],
	}
	return resp
}

func convertParams(method string, params []interface{}) Params {
	switch method {
	case "getTick":
		return FromArray(params, "tick")
	case "getInscribeTxByHeight":
		return FromArray(params, "height")
	case "getCrossBtcTxsByHeight":
		return FromArray(params, "height")
	default:
		return Params{}
	}
}

func RPCError(w http.ResponseWriter, httpStatus int, message string) {
	data, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"result":  nil,
		"error": map[string]interface{}{
			"message": message,
		},
	})
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(httpStatus)
	w.Write(data)
}

func clientAllowed(r *http.Request) bool {
	//this ipAbbr  may be  ::1 when request is localhost
	ipAbbr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Println("RemoteAddr clientAllowed SplitHostPort failure %s \n", r.RemoteAddr)
		return false

	}
	//after ParseIP ::1 chg to 0:0:0:0:0:0:0:1 the true ip
	remoteIp := net.ParseIP(ipAbbr)

	if remoteIp == nil {
		fmt.Println("clientAllowed ParseIP ipAbbr %s failure  \n", ipAbbr)
		return false
	}

	if remoteIp.IsLoopback() {
		return true
	}

	//for _, cfgIp := range config.Parameters.RpcConfiguration.WhiteIPList {
	//	//WhiteIPList have 0.0.0.0  allow all ip in
	//	if cfgIp == "0.0.0.0" {
	//		return true
	//	}
	//	if cfgIp == remoteIp.String() {
	//		return true
	//	}
	//}
	return false
}

func checkAuth(r *http.Request) bool {
	//if (config.Parameters.RpcConfiguration.User == config.Parameters.RpcConfiguration.Pass) &&
	//	(len(config.Parameters.RpcConfiguration.User) == 0) {
	//	return true
	//}
	//authHeader := r.Header["Authorization"]
	//if len(authHeader) <= 0 {
	//	return false
	//}

	//authSha256 := sha256.Sum256([]byte(authHeader[0]))
	//
	//login := config.Parameters.RpcConfiguration.User + ":" + config.Parameters.RpcConfiguration.Pass
	//auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(login))
	//cfgAuthSha256 := sha256.Sum256([]byte(auth))
	//
	//resultCmp := subtle.ConstantTimeCompare(authSha256[:], cfgAuthSha256[:])
	//if resultCmp == 1 {
	//	return true
	//}

	// Request's auth doesn't match  user
	//return false

	return true
}
