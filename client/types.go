package main

import "encoding/json"

type TokenInfo struct {
	SrvToken string `json:"token"`
	ExpireAt int64  `json:"expireAt"`
}

/*
type Product struct {
	ProductId   string `json:"productId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AccessMode  string `json:"accessMode"`
	SlaveMode   bool   `json:"slaveMode"`
}
*/

//{"description":"xxx","direction":1,"dpId":3,"max":300,"min":5,"nameCn":"triger","nameEn":"sonicTriger","resolution":"1","type":"float","unit":"todo"}
type DpInfo struct {
	DpId        int         `json:"dpId"`
	NameEn      string      `json:"nameEn"`
	NameCn      string      `json:"nameCn"`
	Direction   int         `json:"direction,omitempty"`
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Resolution  json.Number `json:"resolution,omitempty"`
	Max         int         `json:"max,omitempty"`
	Min         int         `json:"min,omitempty"`
	Unit        string      `json:"unit,omitempty"`
}

type Product struct {
	ProductId   string   `json:"productId"`
	Name        string   `json:"name"`
	Description string   `json:"description, omitempty"`
	AccessMode  string   `json:"accessMode, omitempty"`
	SlaveMode   bool     `json:"slaveMode, omitempty"`
	Datapoints  []DpInfo `json:"datapoints"`
}

type Dp struct {
	DpId   uint16
	DpType uint16
	DpVal  []byte
}

type RtData struct {
	DevId string `json:"devId"`
	PrdId string `json:"prdId"`
	StoId string `json:"stoId"`
	Data  string `json:"data"`
}
