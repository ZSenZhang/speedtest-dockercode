package spdb

import "time"

type JSONPoint struct {
	Type  string    `json:"type"`
	Coord []float64 `json:"coordinates"`
}

type OoklaInfo struct {
	CountryCode string `json:"cc"`
	Sponsor     string `json:"sponsor"`
	Https       int    `json:"https_functional"`
	Url         string `json:"url"`
}

type MlabInfo struct {
	Url     string   `json:"url"`
	Version []string `json:"version"`
}

type SpeedServer struct {
	Type        string      `json:"type"`
	Id          string      `json:"id"`
	Identifier  string      `json:"identifier"`
	Location    JSONPoint   `json:"location"`
	Country     string      `json:"country"`
	Host        string      `json:"host"`
	City        string      `json:"city"`
	IPv4        string      `json:"ipv4"`
	IPv6        string      `json:"ipv6"`
	Asnv4       string      `json:"asn"`
	Enabled     bool        `json:"enabled"`
	LastUpdated time.Time   `json:"lastupdated"`
	Additional  interface{} `json:"additional"`
}

type VMDataStatus struct {
	Mon        string `json:"mon"`
	BdrmapFile string `json:"bdrmapfile"`
	TraceFile  string `json:"trfile"`
}

//struct for storing interdomain link
type Link struct {
	Region  string `json:"region"`
	Linkkey string `json:"linkkey"`
	NearIP  string `json:"nearip"`
	FarIP   string `json:"farip"`
	FarAS   string `json:"faras"`
	Covered bool   `json:"covered"`
}
