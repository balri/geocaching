package api

var CacheTypes = map[int]string{
	2:    "Traditional",
	3:    "Multi",
	4:    "Virtual",
	5:    "Letterbox Hybrid",
	6:    "Event",
	8:    "Unknown",
	11:   "Webcam",
	13:   "CITO",
	137:  "Earthcache",
	453:  "Mega",
	1858: "Wherigo",
	7005: "Giga",
}

var CacheSizes = map[int]string{
	2: "Micro",
	3: "Regular",
	4: "Large",
	6: "Other",
	8: "Small",
}

var CacheAttributes = map[int]string{
	1:  "Dogs",
	2:  "Access or parking fee",
	3:  "Climbing gear",
	4:  "Boat",
	5:  "Scuba gear",
	6:  "Recommended for kids",
	7:  "Takes less than an hour",
	8:  "Scenic view",
	9:  "Significant hike",
	10: "Difficult climbing",
	11: "May require wading",
	12: "May require swimming",
	13: "Available at all times",
	14: "Recommended at night",
	15: "Available during winter",
	16: "Cacti nearby", //retired, https://www.geocaching.com/geocache/GC684
	17: "Poison plants",
	18: "Dangerous Animals",
	19: "Ticks",
	20: "Abandoned mines",
	21: "Cliff / falling rocks",
	22: "Hunting",
	23: "Dangerous area",
	24: "Wheelchair accessible",
	25: "Parking available",
	26: "Public transportation",
	27: "Drinking water nearby",
	28: "Public restrooms nearby",
	29: "Telephone nearby",
	30: "Picnic tables nearby",
	31: "Camping available",
	32: "Bicycles",
	33: "Motorcycles",
	34: "Quads",
	35: "Off-road vehicles",
	36: "Snowmobiles",
	37: "Horses",
	38: "Campfires",
	39: "Thorns",
	40: "Stealth required",
	41: "Stroller accessible",
	42: "Needs maintenance",
	43: "Watch for livestock",
	44: "Flashlight required",
	46: "Truck Driver/RV",
	47: "Field Puzzle",
	48: "UV Light Required",
	49: "Snowshoes",
	50: "Cross Country Skis",
	51: "Special Tool Required",
	52: "Night Cache",
	53: "Park and Grab",
	54: "Abandoned Structure",
	55: "Short hike (less than 1km)",
	56: "Medium hike (1km-10km)",
	57: "Long Hike (+10km)",
	58: "Fuel Nearby",
	59: "Food Nearby",
	60: "Wireless Beacon",
	61: "Partnership cache",
	62: "Seasonal Access",
	63: "Tourist Friendly",
	64: "Tree Climbing",
	65: "Front Yard (Private Residence)",
	66: "Teamwork Required",
	67: "GeoTour",
	69: "Bonus cache",
	70: "Power trail",
	71: "Challenge Cache",
	72: "Geocaching.com solution checker",
}

// Most options are omitted if empty
// Bool values are 1, 0 or omitted to show both
// Found by and not found by use multiple keys for multiple values
// Matrix is comma separated list of 'd-t'
// Difficulty and terrain are comma separated list
// Dates in format YYYY-MM-DD
// Start and end date go together, all other dates are mutually exclusive
var QueryParams struct{
	CacheTypes              []int    `json:"ct",omitempty`
	CacheSizes              []int    `json:"cs",omitempty`
	SearchTerm              string   `json:"st"`
	OperationType           string   `json:"ot"`
	Radius                  int      `json:"r"`
	CacheName               string   `json:"cn",omitempty`
	HideFound               bool     `json:"hf",omitempty`
	NotFoundBy              string   `json:"nfb",omitempty`
	HideOwned               bool     `json:"ho",omitempty`
	FoundBy                 string   `json:"fb",omitempty`
	Matrix                  []string `json:"m",omitempty`
	SortAsc                 bool     `json:"asc"`
	Sort                    string   `json:"sort"`
	ShowDisabled            bool     `json:"sd",omitempty`
	PremiumOnly             bool     `json:"sp",omitempty`
	ShowCorrectedCoordsOnly bool     `json:"cc",omitempty`
	ShowCacheNotesOnly      bool     `json:"pn",omitempty`
	FavouritePoints         int      `json:"fp",omitempty`
	Difficulty              []string `json:"d",omitempty`
	Terrain                 []string `json:"t",omitempty`
	FoundBeforeDate         string   `json:"fbd",omitempty`
	FoundStartDate          string   `json:"fsd",omitempty`
	FoundEndDate            string   `json:"fed",omitempty`
	FoundOnDate             string   `json:"fod",omitempty`
	FoundAfterDate          string   `json:"fad",omitempty`
	PlacedBeforeDate        string   `json:"pbd",omitempty`
	PlacedStartDate         string   `json:"psd",omitempty`
	PlacedEndDate           string   `json:"ped",omitempty`
	PlacedOnDate            string   `json:"pod",omitempty`
	PlacedAfterDate         string   `json:"pad",omitempty`
	Attributes              []int    `json:"att",omitempty`
}
