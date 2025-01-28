package models

import "encoding/json"

type Event struct {
	GlobalEventID  string  `json:"GlobalEventID"`
	Date           string  `json:"Date"`
	SourceActor    Actor   `json:"SourceActor"`
	TargetActor    Actor   `json:"TargetActor"`
	EventCode      string  `json:"EventCode"`
	EventRootCode  string  `json:"EventRootCode"`
	GoldsteinScale float64 `json:"GoldsteinScale"`
	AvgTone        float64 `json:"AvgTone"`
	NumMentions    int     `json:"NumMentions"`
	NumSources     int     `json:"NumSources"`
	NumArticles    int     `json:"NumArticles"`
	SourceURL      string  `json:"SourceURL"`
	Lat            float64 `json:"Lat"`
	Lng            float64 `json:"Lng"`
	Country        string  `json:"Country"`
}

type Actor struct {
	Code        string `json:"Code"`
	Name        string `json:"Name"`
	CountryCode string `json:"CountryCode"`
}

type CountryPolygon struct {
	Name     string
	Polygons [][][][2]float64
}

type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   Geometry               `json:"geometry"`
}

type Geometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}
