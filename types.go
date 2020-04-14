package main

import (
	"fmt"
	"html/template"
	"math"
	"strings"
)

//////// TYPES /////////
type Countries struct {
	AffectedCountries []string `json:"affected_countries"`
}

/*CoronaRecord holds one report per country */
type CoronaRecord struct {
	IDString                   string `json:"id,unknown"`
	CountryName                string `json:"country_name,unknown"`
	TotalCasesString           string `json:"total_cases,unknown"`
	NewCasesString             string `json:"new_cases,unknown"`
	ActiveCasesString          string `json:"active_cases,unknown"`
	TotalDeathsString          string `json:"total_deaths,unknown"`
	NewDeathsString            string `json:"new_deaths,unknown"`
	TotalRecoveredString       string `json:"total_recovered,unknown"`
	SeriousCriticalString      string `json:"serious_critical,unknown"`
	Region                     string `json:"region,unknown"`
	TotalCasesPerMillionString string `json:"total_cases_per1m,unknown"`
	RecordDate                 string `json:"record_date,unknown"`
}

/* Full per-country data */
type CoronaList struct {
	Country        string         `json:"country,unknown"`
	StatsByCountry []CoronaRecord `json:"stat_by_country,unknown"`
	URL            []string
	sb             string
	asc            bool
	Countries      []string
	Charts         template.HTML
}

func (d *CoronaList) Len() int { return len(d.StatsByCountry) }
func (d *CoronaList) Swap(i, j int) {
	temp := d.StatsByCountry[i]
	d.StatsByCountry[i] = d.StatsByCountry[j]
	d.StatsByCountry[j] = temp
}
func (d *CoronaList) Less(i, j int) bool {
	return d.asc == (strings.Compare(d.StatsByCountry[i].RecordDate, d.StatsByCountry[j].RecordDate) < 0)
}

func (d *CoronaList) timeSeries() []string {
	//ret := make([]time.Time, len(Countries.StatsByCountry))
	ret := make([]string, len(d.StatsByCountry))
	for i, val := range d.StatsByCountry {
		t := strings.Split(val.RecordDate, " ")[0]
		//t, e := time.Parse("2006-01-02 15:04:05.000", val.RecordDate)
		ret[i] = t
	}
	return ret
}

func (d *CoronaList) totalCases() (gd GraphData) {
	gd = d.initHelper()
	for i, val := range d.StatsByCountry {
		t, e := helperScan(val.TotalCasesString)
		if e == nil {
			gd.max, gd.min = helperBL(gd.series, i, t, gd.max, gd.min)
		}
	}
	return
}

func (d *CoronaList) newCases() (gd GraphData) {
	gd = d.initHelper()
	for i, val := range d.StatsByCountry {
		t, e := helperScan(val.NewCasesString)
		if e == nil {
			gd.max, gd.min = helperBL(gd.series, i, t, gd.max, gd.min)
		}
	}
	return
}

func (d *CoronaList) active() (gd GraphData) {
	gd = d.initHelper()
	for i, val := range d.StatsByCountry {
		t, e := helperScan(val.ActiveCasesString)
		if e == nil {
			gd.max, gd.min = helperBL(gd.series, i, t, gd.max, gd.min)
		}
	}
	return
}

func (d *CoronaList) newDeaths() (gd GraphData) {
	gd = d.initHelper()
	for i, val := range d.StatsByCountry {
		t, e := helperScan(val.NewDeathsString)
		if e == nil {
			gd.max, gd.min = helperBL(gd.series, i, t, gd.max, gd.min)
		}
	}
	return
}

func (d *CoronaList) deaths() (gd GraphData) {
	gd = d.initHelper()
	for i, val := range d.StatsByCountry {
		t, e := helperScan(val.TotalDeathsString)
		if e == nil {
			gd.max, gd.min = helperBL(gd.series, i, t, gd.max, gd.min)
		}
	}
	return
}

func (d *CoronaList) serious() (gd GraphData) {
	gd = d.initHelper()
	for i, val := range d.StatsByCountry {
		t, e := helperScan(val.SeriousCriticalString)
		if e == nil {
			gd.max, gd.min = helperBL(gd.series, i, t, gd.max, gd.min)
		}
	}
	return
}

func (d *CoronaList) stats() (gd GraphData) {
	gd = d.initHelper()
	for i, val := range d.StatsByCountry {
		t, e := helperScan(val.TotalCasesPerMillionString)
		if e == nil {
			gd.max, gd.min = helperBL(gd.series, i, t, gd.max, gd.min)
		}
	}
	return
}

/**
initHelper internal method
*/
type GraphData struct {
	series []int
	max    int
	min    int
}

func (d *CoronaList) initHelper() GraphData {
	ret := make([]int, len(d.StatsByCountry))
	max := math.MinInt64
	min := math.MaxInt64
	return GraphData{ret, max, min}
}

///////////////////////////////////////////////////////////////////////
func helperScan(val string) (int, error) {
	var t int
	_, e := fmt.Sscan(strings.ReplaceAll(val, ",", ""), &t)
	return t, e
}

func helperBL(ret []int, i int, t int, max int, min int) (int, int) {
	ret[i] = t
	if max < t {
		max = t
	}
	if min > t {
		min = t
	}
	return max, min
}
