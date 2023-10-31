package valhalla

import (
	"fmt"
	"log"
	"os"
)

func Example_composeAssertTestInputData() {
	expectedJSON, err := os.ReadFile("examples/valhalla_directions_example.json")
	if err != nil {
		log.Fatal(err)
	}

	d, err := composeAssertTestInputData(string(expectedJSON))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#v", d)
	// Output: &valhalla.assertTestInputData{DepartAt:"2023-10-19T21:26", OriginLat:10.814736, OriginLon:106.71283, DestinationLat:10.85404, DestinationLon:106.661329, TravelType:"motorcycle"}
}
