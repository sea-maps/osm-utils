package valhalla

import "encoding/json"

type Location struct {
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	DateTime      string  `json:"date_time,omitempty"`
	SideOfStreet  string  `json:"side_of_street,omitempty"`
	OriginalIndex int     `json:"original_index,omitempty"`
}

type CostingOptions struct {
	Motorcycle json.RawMessage `json:"motorcycle,omitempty"`
}

type DateTimeOptions struct {
	Type  int    `json:"type"`
	Value string `json:"value"`
}

type RouteRequest struct {
	Locations      []Location      `json:"locations"`
	Costing        string          `json:"costing"`
	CostingOptions CostingOptions  `json:"costing_options"`
	DateTime       DateTimeOptions `json:"date_time,omitempty"`
}

type RouteResponse struct {
	ID   string `json:"id"`
	Trip Trip   `json:"trip"`
}

type Trip struct {
	Locations     []Location `json:"locations"`
	Legs          []Legs     `json:"legs"`
	Summary       Summary    `json:"summary"`
	Units         string     `json:"units"`
	Status        int        `json:"status"`
	StatusMessage string     `json:"status_message"`
	Language      string     `json:"language"`
}

type Legs struct {
	Summary Summary `json:"summary"`
	Shape   string  `json:"shape"`
}

type Maneuvers struct {
	Type                                int      `json:"type"`
	Instruction                         string   `json:"instruction"`
	VerbalSuccinctTransitionInstruction string   `json:"verbal_succinct_transition_instruction"`
	VerbalPreTransitionInstruction      string   `json:"verbal_pre_transition_instruction"`
	VerbalPostTransitionInstruction     string   `json:"verbal_post_transition_instruction"`
	StreetNames                         []string `json:"street_names"`
	Time                                float64  `json:"time"`
	Length                              float64  `json:"length"`
	Cost                                float64  `json:"cost"`
	BeginShapeIndex                     int      `json:"begin_shape_index"`
	EndShapeIndex                       int      `json:"end_shape_index"`
	VerbalMultiCue                      bool     `json:"verbal_multi_cue"`
	TravelMode                          string   `json:"travel_mode"`
	TravelType                          string   `json:"travel_type"`
}

type Summary struct {
	Time               float64 `json:"time"`
	Length             float64 `json:"length"`
	Cost               float64 `json:"cost"`
	HasTimeRestriction bool    `json:"has_time_restriction"`
	HasToll            bool    `json:"has_toll"`
	HasHighway         bool    `json:"has_highway"`
	HasFerry           bool    `json:"has_ferry"`
	MinLat             float64 `json:"min_lat"`
	MinLon             float64 `json:"min_lon"`
	MaxLat             float64 `json:"max_lat"`
	MaxLon             float64 `json:"max_lon"`
}
