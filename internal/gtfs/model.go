package gtfs

type Stop struct {
	ID                 string
	Code               string
	Name               string
	Lat                float64
	Lon                float64
	LocationType       int
	ParentStation      string
	WheelchairBoarding int
}

type StopTime struct {
	TripID        string
	ArrivalTime   int
	DepartureTime int
	StopID        string
	StopSequence  int
}

type Trip struct {
	RouteID      string
	ServiceID    string
	TripID       string
	Headsign     string
	DirectionID  int
	ShapeID      string
	Wheelchair   int
}

type Route struct {
	ID        string
	AgencyID  string
	ShortName string
	LongName  string
	Type      int
}

type Calendar struct {
	ServiceID string
	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
	Saturday  bool
	Sunday    bool
	StartDate string
	EndDate   string
}

type CalendarDate struct {
	ServiceID     string
	Date          string
	ExceptionType int
}

type Transfer struct {
	FromStopID      string
	ToStopID        string
	TransferType    int
	MinTransferTime int
}

type Feed struct {
	Stops         []Stop
	StopTimes     []StopTime
	Trips         []Trip
	Routes        []Route
	Calendars     []Calendar
	CalendarDates []CalendarDate
	Transfers     []Transfer
}
