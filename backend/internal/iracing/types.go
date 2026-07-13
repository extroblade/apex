package iracing

import "time"

// License is one category's license/rating for a member.
type License struct {
	CategoryID   int     `json:"category_id"`
	Category     string  `json:"category"`
	LicenseLevel int     `json:"license_level"`
	SafetyRating float64 `json:"safety_rating"`
	IRating      int     `json:"irating"`
	GroupName    string  `json:"group_name"`
}

// Member is a driver's profile with per-category licenses.
type Member struct {
	CustID      int       `json:"cust_id"`
	DisplayName string    `json:"display_name"`
	Licenses    []License `json:"licenses"`
}

// CareerStat is aggregate career performance in one category.
type CareerStat struct {
	Category      string  `json:"category"`
	CategoryID    int     `json:"category_id"`
	Starts        int     `json:"starts"`
	Wins          int     `json:"wins"`
	Top5          int     `json:"top5"`
	Poles         int     `json:"poles"`
	AvgStartPos   float64 `json:"avg_start_position"`
	AvgFinishPos  float64 `json:"avg_finish_position"`
	Laps          int     `json:"laps"`
	LapsLed       int     `json:"laps_led"`
	AvgIncidents  float64 `json:"avg_incidents"`
	WinPercentage float64 `json:"win_percentage"`
}

// Track is the minimal track reference embedded in a race.
type Track struct {
	TrackID   int    `json:"track_id"`
	TrackName string `json:"track_name"`
}

// RecentRace is one finished race from a member's recent history.
type RecentRace struct {
	SubsessionID     int       `json:"subsession_id"`
	SeriesID         int       `json:"series_id"`
	SeriesName       string    `json:"series_name"`
	SessionStartTime time.Time `json:"session_start_time"`
	CarID            int       `json:"car_id"`
	Track            Track     `json:"track"`
	StartPosition    int       `json:"start_position"`
	FinishPosition   int       `json:"finish_position"`
	Incidents        int       `json:"incidents"`
	OldiRating       int       `json:"oldi_rating"`
	NewiRating       int       `json:"newi_rating"`
	LapsComplete     int       `json:"laps_complete"`
	CategoryID       int       `json:"category_id"`
}

// DriverSearchResult is a hit from the driver lookup endpoint.
type DriverSearchResult struct {
	CustID      int    `json:"cust_id"`
	DisplayName string `json:"display_name"`
}
