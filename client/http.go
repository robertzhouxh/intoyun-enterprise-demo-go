package main

import (
	"encoding/json"
	"errors"
	"fmt"
	inet "intoyun-enterprise-demo-go/libs/network"
)

func initHttp() error {
	var (
		tokenInfo = new(TokenInfo)
	)

	// step1: 获取第三方服务器对应的token
	ret, err := sendPost("", httpAddr+"/v1/token")
	if err != nil {
		return err
	}
	json.Unmarshal(ret, tokenInfo)
	//fmt.Printf("tokeninfo:%v\n\n", tokenInfo)

	// step2: 获取产品列表
	ret, err = sendGet(tokenInfo.SrvToken, httpAddr+"/v1/product", "")
	if err != nil {
		fmt.Printf("get product list Err: %v\n", err)
		return err
	}
	if err = json.Unmarshal(ret, &prds); err != nil {
		return err
	}
	//fmt.Printf("product list ===>:\n\n%v\n\n", prds)

	// step3: 获取所有的产品对应的数据点信息
	for _, item := range prds {
		ret, err = sendGet(tokenInfo.SrvToken, httpAddr+"/v1/product/"+item.ProductId, "")
		if err != nil {
			fmt.Printf("get product item failed Err: %v\n", err)
			return err
		}
		//fmt.Printf("\n\nproduct item: %s\n\n", ret)
		prd := Product{}
		if err = json.Unmarshal(ret, &prd); err != nil {
			return err
		}
		prdMap[item.ProductId] = prd
	}

	//for key, val := range prdMap {
	//	fmt.Printf("::::::::::::::::::::::::::::::::\nkey: %s\nProduct: %v\n", key, val)
	//}

	return nil
}

func sendPost(token, url string) ([]byte, error) {
	req := inet.New()
	if len(token) != 0 {
		req.Set("X-IntoYun-SrvToken", token)
	}
	//fmt.Printf("inet.PostJson request===>headerVal: %s, url: %s===>\n", token, url)
	ret, resp, err := inet.PostJson(req, url, []byte("{}"))
	if err != nil {
		fmt.Printf("inet.PostJson(%s) failed Err: %v\n", url, err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("bad request")
	}

	//fmt.Printf("inet.PostJson response:\n------------------------\nstatus:%s\nret:%s\n------------------------\n\n", resp.Status, ret)
	return ret, nil
}

func sendGet(token, url, query string) ([]byte, error) {
	fmt.Printf("inet.GetJson request url=%s\n", url)
	req := inet.New()
	ret, resp, err := inet.GetJson(req, url, token, query)
	if err != nil {
		fmt.Printf("inet.PostJson(%s) failed Err: %v\n", url, err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("bad request")
	}
	//fmt.Printf("inet.PostJson response:\n------------------------\nstatus:%s\nret:%s\n------------------------\n\n", resp.Status, ret)
	return ret, nil
}
