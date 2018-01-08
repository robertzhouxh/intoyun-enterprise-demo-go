package net

import (
	"encoding/json"
	"errors"

	"github.com/parnurzeal/gorequest"
)

//do not use single instance
//var Request *gorequest.SuperAgent = nil
//func init() {
//	Request = gorequest.New()
//}

func New() *gorequest.SuperAgent {
	Request := gorequest.New()
	return Request
}

// func (request *gorequest.SuperAgent) PostJson(url string, data interface{}) ([]byte, gorequest.Response) {
func PostJson(request *gorequest.SuperAgent, url string, data interface{}) ([]byte, gorequest.Response, error) {
	jsonData, err := json.Marshal(data)
	if nil != err {
		return nil, nil, err
	}
	request.Post(url).Send(string(jsonData))
	resp, body, errs := request.End()

	if errs != nil {
		return nil, nil, errors.New("request failed")
	}

	bytes := ([]byte)(body)
	return bytes, resp, nil
}

//func GetJson(request *gorequest.SuperAgent, url string, token string, query map[string]string) ([]byte, gorequest.Response, error) {
func GetJson(request *gorequest.SuperAgent, url string, token string, query string) ([]byte, gorequest.Response, error) {
	request.Get(url).Query(query).Set("X-IntoYun-SrvToken", token)
	resp, body, errs := request.End()
	if errs != nil {
		return nil, nil, errors.New("request failed")
	}

	bytes := ([]byte)(body)
	return bytes, resp, nil
}
