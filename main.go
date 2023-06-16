package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"math"
)

type BusStop struct {
	ID   string  `json:"id"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
	Name string  `json:"name"`
}

type Bus struct {
	BusStops []BusStop `json:"busStops"`
	FullName  string      `json:"fullName"`
	ID        string      `json:"id"`
	Origin    string      `json:"origin"`
	Path      [][]float64 `json:"path"`
	ShortName string      `json:"shortName"`
}

type BusRoute struct {
	Buses []Bus `json:"payload"`
	Status int `json:"status"`
}

type BusLocation struct {
	Bearing      float64 `json:"bearing"`
	CrowdLevel   string  `json:"crowdLevel"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	VehiclePlate string  `json:"vehiclePlate"`
}

type Location struct {
	BusLocations []BusLocation `json:"payload"`
	Status int `json:"status"`
}

var busRoute BusRoute
var mapBusStops = make(map[string]BusStop) 
//id as key and bus stop info as value

//get overall data on all bus routes 
func GetBusLinesData() BusRoute {
	busRoute := BusRoute{}
	obj, err := http.Get("https://test.uwave.sg/busLines")
	if err != nil {
		log.Fatal(err)
	}
	defer obj.Body.Close()
	json.NewDecoder(obj.Body).Decode(&busRoute)
	//to get bus i: bus[i]
	return busRoute
}

//get data on the location of a bus line 
func GetBusLocationData(busLineId string) Location {
	busLocation := Location{}
	obj, err := http.Get("https://test.uwave.sg/busPositions/" + busLineId)
	if err != nil {
		log.Fatal(err)
	}
	defer obj.Body.Close()
	json.NewDecoder(obj.Body).Decode(&busLocation)
	// fmt.Println(busLocation.BusLocations[0].VehiclePlate)
	return busLocation
}

//get bus routes for the specific bus 
func GetBusRoute(busId string) []BusStop { 
	if (len(busRoute.Buses)==0){
		busRoute = GetBusLinesData()
	}
	for _, bus := range busRoute.Buses {
		// fmt.Println(v.ID)
		if(bus.ID == busId) {
			// fmt.Println(bus.BusStops)
			return bus.BusStops
		}
	}
	return nil
}

//get all buses that stops at a specific bus stop
func GetBusAvailable(busStopId string) []Bus { 
	if(len(busRoute.Buses)==0) {
		busRoute = GetBusLinesData()
	}
	var buses = []Bus{}
	for _, bus := range busRoute.Buses {
		for _, z := range bus.BusStops {
			if _, ok := mapBusStops[z.ID]; !ok {
				mapBusStops[z.ID] = z
			}
			if(z.ID==busStopId) {
				buses=append(buses, bus)
				break
			}
		}
	}
	return buses
}

//calculates the amount of time needed for the bus to reach the bus stop 
func GetEstimatedDuration(busStopId, busLineId string) []float64 {
	var currentBusStop BusStop 
	//find bus stop info 
	if _, ok := mapBusStops[busStopId]; ok {
		currentBusStop = mapBusStops[busStopId]
	} else {
		if(len(busRoute.Buses)==0) {
			busRoute = GetBusLinesData()
		}
		for _, bus := range busRoute.Buses {
			for _, z := range bus.BusStops {
				if _, ok := mapBusStops[z.ID]; !ok {
					mapBusStops[z.ID] = z
				}
				if(z.ID==busStopId) {
					currentBusStop = z
					break
				}
			}
		}
	}

	//check if bus stops there
	var busStopsHere []BusStop= GetBusRoute(busLineId) 
	var stopHere bool

	for _, busStop := range busStopsHere {
		if(busStop.ID==busStopId) {
			stopHere = true
		}
	}

	if(!stopHere) {
		return []float64{-1}
	}


	//find current location of bus 
	currentBusLocation := GetBusLocationData(busLineId)
	
	//calculate distance
	const PI float64 = 3.141592653589793
	var timing []float64;
	
	//using haversine formula 
	for _, busLoc := range currentBusLocation.BusLocations {
		radlat1 := float64(PI * currentBusStop.Lat / 180)
		radlat2 := float64(PI * busLoc.Lat / 180)
		
		theta := float64(currentBusStop.Lng - busLoc.Lng)
		radtheta := float64(PI * theta / 180)

		dist := math.Sin(radlat1) * math.Sin(radlat2) + math.Cos(radlat1) * math.Cos(radlat2) * math.Cos(radtheta)
		
		if dist > 1 {
			dist = 1
		}
	
		dist = math.Acos(dist) //reverse cosine 
		dist = dist * 180 / PI
		dist = dist * 60 * 1.1515 * 1.609344

		if(busLoc.CrowdLevel=="crowded" || busLoc.CrowdLevel=="high") {
			timing = append(timing, (dist/17.8)*60)
		} else if(busLoc.CrowdLevel=="low") {
			timing = append(timing, (dist/21.5)*60)
		} else {
			timing = append(timing, (dist/19.65)*60)
		} 
	}
	return timing
}

func main() { 
	fmt.Println("Get Bus Route: ")
	fmt.Println(GetBusRoute("44480"))

	fmt.Println("Get Bus Available: ")
	var check = make([]Bus, 0, 10)
	check = GetBusAvailable("383009") 
	//pinoeer mrt
	//should get back 44481(campus weekend rider brown) and 44480(campus rider green)
	for _, bus := range check { 
		fmt.Println(bus.ID)
	}

	fmt.Println("Get Bus Location Data: ")
	fmt.Println(GetBusLocationData("44479"))

	fmt.Println("Get Estimated Duration: ")
	fmt.Println(GetEstimatedDuration("383009", "44479"))
}