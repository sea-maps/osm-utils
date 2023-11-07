package valhalla

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"os"
	"path/filepath"

	"github.com/montanaflynn/stats"
	"github.com/sea-maps/osm-utils/pkg/api/valhalla"
	"github.com/tidwall/gjson"
	polyline "github.com/twpayne/go-polyline"
	"github.com/umahmood/haversine"
	simplify "github.com/yrsh/simplify-go"
)

var (
	// Polyline6Codec is the polyline codec used by Valhalla.
	Polyline6Codec = polyline.Codec{
		Dim:   2,
		Scale: 1e6,
	}
	RefPoint = haversine.Coord{Lat: 10.777064553034846, Lon: 106.69569180740099}
)

// CompareBatchRoutes will read all the json files in the given directory and
// make a request to the Valhalla service for each one. It will then compare the
// results and output a summary of the differences.
func CompareBatchRoutes(ctx context.Context,
	fileSystem fs.FS,
	valhallaURL string,
	costingJSONFile fs.File,
) error {
	costingJSON, err := io.ReadAll(costingJSONFile)
	if err != nil {
		return err
	}

	return fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) != ".json" {
			return nil
		}

		expectedJSON, err := fs.ReadFile(fileSystem, path)
		if err != nil {
			return fmt.Errorf("failed to read expected json %s, : %w", path, err)
		}
		err = assertData(string(expectedJSON), string(costingJSON), valhallaURL)
		if err != nil {
			fmt.Println(path, err)
			return err
		}

		fmt.Println(path, "OK")
		return nil
	})
}

func assertData(expectedJSON string, costingJSON string, valhallaURL string) error {
	d, err := composeAssertTestInputData(expectedJSON)
	if err != nil {
		return fmt.Errorf("assertData failed to composeAssertTestInputData: %w", err)
	}

	vr := &valhalla.RouteRequest{
		Locations: []valhalla.Location{
			{
				Lat: d.OriginLat,
				Lon: d.OriginLon,
			},
			{
				Lat: d.DestinationLat,
				Lon: d.DestinationLon,
			},
		},
		Costing:        d.CostingMethod,
		CostingOptions: valhalla.CostingOptions{},
		DateTime: valhalla.DateTimeOptions{
			Type:  1,
			Value: d.DepartAt,
		},
	}

	if vr.Costing == "motorcycle" {
		vr.CostingOptions.Motorcycle = []byte(costingJSON)
	}

	reqJSON, err := json.Marshal(vr)
	if err != nil {
		return fmt.Errorf("assertData failed to marshal RouteRequest: %w", err)
	}

	req, err := http.NewRequest("POST", valhallaURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		return fmt.Errorf("assertData failed to create http request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("assertData failed to make http request: %w", err)
	}
	defer resp.Body.Close()

	respJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("assertData failed to read http response: %w", err)
	}

	gotPolyline, err := extractSimplifiedPolylineCoordsFromJSON(string(respJSON))
	if err != nil {
		fmt.Println(string(respJSON))
		return fmt.Errorf("assertData failed to extract polyline from response")
	}

	expectedPolyline, err := extractSimplifiedPolylineCoordsFromJSON(expectedJSON)
	if err != nil {
		return fmt.Errorf("assertData failed to extract polyline from expected")
	}

	if err := compareBasicSummary(string(respJSON), expectedJSON); err != nil {
		return fmt.Errorf("assertData failed to compareBasicSummary: %w", err)
	}

	if simplify.CompareSlices(gotPolyline, expectedPolyline) {
		return nil
	}

	// fmt.Println(getStatsDistanceFromRefPoint(gotPolyline))
	// fmt.Println(getStatsDistanceFromRefPoint(expectedPolyline))

	return nil
}

func compareBasicSummary(gotJSON string, expectedJSON string) error {
	gotHasToll := gjson.Get(gotJSON, "trip.summary.has_toll").Bool()
	expectedHasToll := gjson.Get(expectedJSON, "trip.summary.has_toll").Bool()
	if gotHasToll != expectedHasToll {
		return fmt.Errorf("has_toll is different")
	}

	goHasFerry := gjson.Get(gotJSON, "trip.summary.has_ferry").Bool()
	expectedHasFerry := gjson.Get(expectedJSON, "trip.summary.has_ferry").Bool()
	if goHasFerry != expectedHasFerry {
		return fmt.Errorf("has_ferry is different")
	}

	gotHasHighway := gjson.Get(gotJSON, "trip.summary.has_highway").Bool()
	expectedHasHighway := gjson.Get(expectedJSON, "trip.summary.has_highway").Bool()
	if gotHasHighway != expectedHasHighway {
		return fmt.Errorf("has_highway is different")
	}

	gotLength := gjson.Get(gotJSON, "trip.summary.length").Float()
	expectedLength := gjson.Get(expectedJSON, "trip.summary.length").Float()

	if math.Abs(gotLength-expectedLength) > 0.5 {
		fmt.Println("got polyline", gjson.Get(gotJSON, "trip.legs.0.shape").String())
		return fmt.Errorf("length is different, got %f, expected %f", gotLength, expectedLength)
	}

	percentDifferent := (math.Abs(gotLength-expectedLength) / ((expectedLength + gotLength) / 2)) * 100
	if percentDifferent > 5 {
		fmt.Println("got polyline", gjson.Get(gotJSON, "trip.legs.0.shape").String())
		return fmt.Errorf("length is different by more than 5 percent, got %f, expected %f", gotLength, expectedLength)
	}

	return nil
}

func getStatsDistanceFromRefPoint(coords [][]float64) (float64, float64) {
	var sum float64
	dinstances := make([]float64, 0, len(coords))
	for i := 0; i < len(coords); i++ {
		_, k := haversine.Distance(haversine.Coord{Lat: coords[i][0], Lon: coords[i][1]}, RefPoint)
		dinstances = append(dinstances, k)
		sum += k
	}

	sd, _ := stats.StandardDeviation(dinstances)
	return sum / float64(len(coords)), sd
}

type assertTestInputData struct {
	DepartAt       string
	OriginLat      float64
	OriginLon      float64
	DestinationLat float64
	DestinationLon float64
	CostingMethod  string
}

func composeAssertTestInputData(expectJSON string) (*assertTestInputData, error) {
	d := &assertTestInputData{}
	d.DepartAt = gjson.Get(expectJSON, "trip.locations.0.date_time").String()
	d.OriginLat = gjson.Get(expectJSON, "trip.locations.0.lat").Float()
	d.OriginLon = gjson.Get(expectJSON, "trip.locations.0.lon").Float()
	d.DestinationLat = gjson.Get(expectJSON, "trip.locations.1.lat").Float()
	d.DestinationLon = gjson.Get(expectJSON, "trip.locations.1.lon").Float()
	d.CostingMethod = gjson.Get(expectJSON, "trip.legs.0.maneuvers.0.travel_type").String()
	if d.CostingMethod != "motorcycle" {
		d.CostingMethod = "auto"
	}

	return d, nil
}

func extractSimplifiedPolylineCoordsFromJSON(s string) ([][]float64, error) {
	polylineString := gjson.Get(s, "trip.legs.0.shape").String()
	if len(polylineString) == 0 {
		return nil, fmt.Errorf("extractSimplifiedPolylineCoordsFromJSON failed to get polyline string")
	}

	coords, _, err := Polyline6Codec.DecodeCoords([]byte(polylineString))
	if err != nil {
		return nil, fmt.Errorf("extractSimplifiedPolylineCoordsFromJSON failed to decode polyline string: %w", err)
	}

	return simplify.Simplify(coords, 0.0005, true), nil
}

func listFiles(routeDir string) []string {
	fileNames := []string{}
	fileSystem := os.DirFS(routeDir)
	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		return nil
	})

	return fileNames
}
