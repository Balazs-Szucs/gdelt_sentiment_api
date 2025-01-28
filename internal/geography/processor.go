package geography

import (
	"encoding/json"
	"fmt"
	"os"
	"sentiment_dashboard_api/internal/models"
)

type Processor struct {
	countryPolygons []models.CountryPolygon
}

func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) LoadCountryGeoJSON(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var fc models.FeatureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		return fmt.Errorf("unmarshal geojson: %w", err)
	}

	for _, feature := range fc.Features {
		name, _ := feature.Properties["name"].(string)
		if name == "" {
			name = feature.ID
		}

		switch feature.Geometry.Type {
		case "Polygon":
			var coords [][][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				continue
			}
			p.countryPolygons = append(p.countryPolygons, models.CountryPolygon{
				Name:     name,
				Polygons: [][][][2]float64{convertPolygon(coords)},
			})

		case "MultiPolygon":
			var coords [][][][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				continue
			}
			var polys [][][][2]float64
			for _, polyCoords := range coords {
				polys = append(polys, convertPolygon(polyCoords))
			}
			p.countryPolygons = append(p.countryPolygons, models.CountryPolygon{
				Name:     name,
				Polygons: polys,
			})
		}
	}
	return nil
}

func convertPolygon(coords [][][]float64) [][][2]float64 {
	polygon := make([][][2]float64, len(coords))
	for i, ring := range coords {
		polygon[i] = make([][2]float64, len(ring))
		for j, pt := range ring {
			polygon[i][j] = [2]float64{pt[0], pt[1]}
		}
	}
	return polygon
}

func (p *Processor) GetCountry(lat, lng float64) string {
	for _, cp := range p.countryPolygons {
		for _, polygon := range cp.Polygons {
			if isInsidePolygon(lng, lat, polygon) {
				return cp.Name
			}
		}
	}
	return "Other"
}

func isInsidePolygon(lon, lat float64, polygon [][][2]float64) bool {
	for _, ring := range polygon {
		if pointInRing(lon, lat, ring) {
			return true
		}
	}
	return false
}

func pointInRing(lon, lat float64, ring [][2]float64) bool {
	inside := false
	n := len(ring)
	for i := 0; i < n; i++ {
		j := (i + n - 1) % n
		xi, yi := ring[i][0], ring[i][1]
		xj, yj := ring[j][0], ring[j][1]

		intersect := ((yi > lat) != (yj > lat)) &&
			(lon < (xj-xi)*(lat-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
	}
	return inside
}
