package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	inet "intoyun-enterprise-demo-go/libs/network"
	"io"
	"strconv"
	"strings"

	"net"
	"net/http"
	"time"

	log "github.com/thinkboy/log4go"
)

const (
	HOST     = "https://enterprise.intoyun.com/v1"
	ROOT_DIR = "/assets/"
)

type Product struct {
	ProductId   string `json:"productId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AccessMode  string `json:"accessMode"`
	SlaveMode   bool   `json:"slaveMode"`
}

func InitHTTP() (err error) {
	var network, addr string
	for i := 0; i < len(Conf.HTTPAddrs); i++ {
		// register router
		httpServeMux := http.NewServeMux()
		// Notice that because our static directory is set as the root of the FileSystem,
		// we need to strip off the /static/ prefix from the request path before searching the FileSystem for the given file.
		httpServeMux.Handle("/", http.FileServer(http.Dir("."+ROOT_DIR)))
		httpServeMux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("."+ROOT_DIR))))
		httpServeMux.HandleFunc("/v1/token", TokenHandler)
		httpServeMux.HandleFunc("/v1/product", ProductHandler)
		httpServeMux.HandleFunc("/v1/product/", ProductItemHandler)
		httpServeMux.HandleFunc("/v1/device", DeviceHandler)
		httpServeMux.HandleFunc("/v1/control", ControlHandler)
		httpServeMux.HandleFunc("/v1/sensordata", SensordataHandler)

		log.Info("start http server listenning ===> %s", Conf.HTTPAddrs[i])

		// start http server
		if network, addr, err = inet.ParseNetwork(Conf.HTTPAddrs[i]); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		go httpListen(httpServeMux, network, addr)
	}
	return
}

func httpListen(mux *http.ServeMux, network, addr string) {
	httpServer := &http.Server{Handler: mux, ReadTimeout: Conf.HTTPReadTimeout, WriteTimeout: Conf.HTTPWriteTimeout}
	httpServer.SetKeepAlivesEnabled(true)
	l, err := net.Listen(network, addr)
	if err != nil {
		log.Error("net.Listen(\"%s\", \"%s\") error(%v)", network, addr, err)
		panic(err)
	}
	if err := httpServer.Serve(l); err != nil {
		log.Error("server.Serve() error(%v)", err)
		panic(err)
	}
}

// retWrite marshal the result and write to client(get).
func retWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, start time.Time) {
	data, err := json.Marshal(res)
	if err != nil {
		log.Error("json.Marshal(\"%v\") error(%v)", res, err)
		return
	}
	dataStr := string(data)
	if _, err := w.Write([]byte(dataStr)); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", dataStr, err)
	}
	log.Info("req: \"%s\", get: res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}

// retPWrite marshal the result and write to client(post).
func retPWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, body *string, start time.Time) {
	data, err := json.Marshal(res)
	if err != nil {
		log.Error("json.Marshal(\"%v\") error(%v)", res, err)
		return
	}
	dataStr := string(data)
	if _, err := w.Write([]byte(dataStr)); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", dataStr, err)
	}
	log.Info("req: \"%s\", post: \"%s\", res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), *body, dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}

//------------------------http handlers------------------------------
// curl -X POST http://localhost:8080/v1/token
func TokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		body string
		res  = map[string]interface{}{}
	)
	defer retPWrite(w, r, res, &body, time.Now())

	timeStamp := time.Now().UnixNano() / 1e6
	//timeStampStr := strconv.Itoa(timeStamp)
	timeStampStr := strconv.FormatInt(timeStamp, 10)

	h := md5.New()
	io.WriteString(h, timeStampStr+Conf.AppSecret)
	signature := hex.EncodeToString(h.Sum(nil))

	payload := map[string]string{
		"appId":     Conf.AppId,
		"timestamp": timeStampStr,
		"signature": signature,
	}

	ret, resp, err := inet.PostJson(inet.New(), HOST+"/token", payload)
	if err != nil {
		log.Error("inet.PostJson err: %v", err)
	}
	log.Debug("inet.New().PostJson:\nresp===> %+v\n", resp)
	//ret, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(ret, &res)
	log.Debug("inet.New().PostJson:\ntokeninfo===> %+v\n", res)

	return
}

func ProductHandler(w http.ResponseWriter, r *http.Request) {
	// must treat the result (array) as the item slice
	prds := []Product{} // <<--- An array of your struct
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		res []byte
	)
	token := r.Header.Get("X-IntoYun-SrvToken")
	if token == "" {
		http.Error(w, "Invalid X-IntoYun-SrvToken", 403)
		return
	}

	res, resp, err := inet.GetJson(inet.New(), HOST+"/product", token, r.URL.RawQuery)
	if err != nil {
		log.Error("inet.GetJson err: %v", err)
	}
	if resp.StatusCode != 200 {
		http.Error(w, "invalid", resp.StatusCode)
		return
	}

	//log.Debug("GetJson Product ===> %s\n, len: %s", res, len(res))
	err = json.Unmarshal([]byte(res), &prds)
	if err != nil {
		log.Error("json.Unmarshal(\"%s\") error(%v)", res, err)
	}

	//log.Debug("Product array is: ===> %#v\n", prds)
	// --------------------------------------------------------------------
	// you can process the product here, eg, store the product into your db
	// --------------------------------------------------------------------
	if _, err := w.Write(res); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", res, err)
	}

	return
}

func ProductItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
	//		res = map[string]interface{}{}
	)
	//defer retWrite(w, r, res, time.Now())
	token := r.Header.Get("X-IntoYun-SrvToken")
	if token == "" {
		http.Error(w, "Invalid X-IntoYun-SrvToken", 403)
		return
	}

	items := strings.Split(r.URL.Path, "/")
	if len(items) != 4 {
		http.Error(w, "Bab Request", 400)
	}

	ret, resp, err := inet.GetJson(inet.New(), HOST+"/product/"+items[3], token, r.URL.RawQuery)
	if err != nil {
		log.Error("inet.GetJson err: %v", err)
	}
	if resp.StatusCode != 200 {
		http.Error(w, "invalid", resp.StatusCode)
		return
	}
	if _, err := w.Write(ret); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", ret, err)
	}
	//json.Unmarshal(ret, &res)
	return
}

func DeviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		res []byte
	)
	token := r.Header.Get("X-IntoYun-SrvToken")
	if token == "" {
		http.Error(w, "Invalid X-IntoYun-SrvToken", 403)
		return
	}

	res, resp, err := inet.GetJson(inet.New(), HOST+"/device", token, r.URL.RawQuery)
	if err != nil {
		log.Error("inet.GetJson err: %v", err)
	}
	if resp.StatusCode != 200 {
		http.Error(w, "invalid", resp.StatusCode)
		return
	}
	log.Debug("inet.New().GetJson:\ndevice===> %+v\n", res)
	if _, err := w.Write(res); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", res, err)
	}

	return
}

// curl -X POST --header "X-IntoYun-SrvToken:6a08820093e15327cfe5ee13cd31d74f" --header "Content-Type: application/json" -d '[{"dpId": 1, "type": "enum", "value": 0}]' http://localhost:8080/v1/control?productId=VhyODqsghXMMq214&deviceId=b
func ControlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		body string
		res  = map[string]interface{}{}
	)
	defer retPWrite(w, r, res, &body, time.Now())
	token := r.Header.Get("X-IntoYun-SrvToken")
	if token == "" {
		http.Error(w, "Invalid X-IntoYun-SrvToken", 403)
		return
	}

	ret, resp, err := inet.GetJson(inet.New(), HOST+"/control", token, r.URL.RawQuery)
	if err != nil {
		log.Error("inet.GetJson err: %v", err)
	}
	if resp.StatusCode != 200 {
		http.Error(w, "invalid", resp.StatusCode)
		return
	}
	json.Unmarshal(ret, &res)
	log.Debug("inet.New().GetJson:\ncontrol===> %+v\n", res)

	return
}

func SensordataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		res []byte
	)
	token := r.Header.Get("X-IntoYun-SrvToken")
	if token == "" {
		http.Error(w, "Invalid X-IntoYun-SrvToken", 403)
		return
	}

	res, resp, err := inet.GetJson(inet.New(), HOST+"/sensordata", token, r.URL.RawQuery)
	if err != nil {
		log.Error("inet.GetJson err: %v", err)
	}
	if resp.StatusCode != 200 {
		http.Error(w, "invalid", resp.StatusCode)
		return
	}

	log.Debug("inet.New().GetJson:\nsensordata===> %+v\n", res)

	if _, err := w.Write(res); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", res, err)
	}
	return
}
