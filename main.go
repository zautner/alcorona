package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"

	"html/template"
	"net/http"

	"github.com/chenjiandongx/go-echarts/charts"
)

//////// TYPES /////////
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
	//ret := make([]time.Time, len(d.StatsByCountry))
	ret := make([]string, len(d.StatsByCountry))
	for i, val := range d.StatsByCountry {
		t := strings.Split(val.RecordDate, " ")[0]
		//t, e := time.Parse("2006-01-02 15:04:05.000", val.RecordDate)
		ret[i] = t
		println(val.RecordDate, ret[i])
	}
	return ret
}

func (d *CoronaList) totalCases() []int {
	ret := make([]int, len(d.StatsByCountry))
	for i, val := range d.StatsByCountry {
		var t int
		_, e := fmt.Sscan(strings.ReplaceAll(val.TotalCasesString, ",", ""), &t)
		if e == nil {
			ret[i] = t
		}
		println(t, ret[i])
	}
	return ret
}
func (d *CoronaList) newCases() []int {
	ret := make([]int, len(d.StatsByCountry))
	for i, val := range d.StatsByCountry {
		var t int
		_, e := fmt.Sscan(val.NewCasesString, &t)
		if e == nil {
			ret[i] = t
		}
		println(t, ret[i])
	}
	return ret
}

func (d *CoronaList) newDeaths() []int {
	ret := make([]int, len(d.StatsByCountry))
	for i, val := range d.StatsByCountry {
		var t int
		_, e := fmt.Sscan(val.NewDeathsString, &t)
		if e == nil {
			ret[i] = t
		}
		println(t, ret[i])
	}
	return ret
}
func (d *CoronaList) deaths() []int {
	ret := make([]int, len(d.StatsByCountry))
	for i, val := range d.StatsByCountry {
		var t int
		_, e := fmt.Sscan(val.TotalDeathsString, &t)
		if e == nil {
			ret[i] = t
		}
		println(t, ret[i])
	}
	return ret
}

//////// VARIABLES /////////
var tpl *template.Template

//////// FUNCTIONS /////////

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

func main() {
	http.HandleFunc("/chart/", chart)
	http.HandleFunc("/", index)
	println(1)
	err := http.ListenAndServe(":80", nil)
	if nil != err {
		println("Error!", err.Error())
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	d := getDataJson(r.RequestURI[1:])
	tpl.ExecuteTemplate(w, "index.gohtml", d)
	drawChart(d, w)
}

func chart(w http.ResponseWriter, r *http.Request) {
	d := getDataJson(r.RequestURI[7:])
	drawChart(d, w)
}

func drawChart(d *CoronaList, w http.ResponseWriter) {
	graph := charts.NewLine()
	graph.SetGlobalOptions(charts.TitleOpts{Title: "Corona cases", Subtitle: d.Country})
	graph.AddXAxis(d.timeSeries()).AddYAxis("Total cases", d.totalCases()).AddYAxis("Deaths", d.deaths())
	graphES := charts.NewEffectScatter()
	graphES.AddXAxis(d.timeSeries()).AddYAxis("New cases", d.newCases()).AddYAxis("New deaths", d.newDeaths())
	f, e := os.Create("line-" + strconv.Itoa(rand.Int()) + ".html")
	defer os.Remove(f.Name())
	if e == nil {
		graph.Render(w, f)
		graphES.Render(w, f)
	}
}

func getDataJson(uri string) *CoronaList {
	println(uri)
	country := "Israel"
	splitten := strings.Split(uri, "/")
	if len(splitten) > 0 {
		country = splitten[0]
	}
	asc := len(splitten) <= 1 || splitten[1] != "desc"
	d, err := readData(country, asc)
	if err != nil {
		panic(err)
	}
	return d
}

func readData(country string, asc bool) (*CoronaList, error) {
	url := "https://coronavirus-monitor.p.rapidapi.com/coronavirus/cases_by_particular_country.php?country=" + country
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("x-rapidapi-host", "coronavirus-monitor.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", "051ca7468fmsh2584062d4642570p169ec0jsn5598e50e8382")

	res, err := http.DefaultClient.Do(req)
	if nil != err {
		panic(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	println(string(body)[:1])
	d := CoronaList{URL: []string{country}, asc: asc}
	json.Unmarshal(body, &d)
	statsByCountry := make([]CoronaRecord, 0)
	sort.Sort(&d)
	for i, v := range d.StatsByCountry {
		if i == len(d.StatsByCountry)-1 || strings.Split(v.RecordDate, " ")[0] != strings.Split(d.StatsByCountry[i+1].RecordDate, " ")[0] {
			statsByCountry = append(statsByCountry, v)
		}
	}
	d.StatsByCountry = statsByCountry
	return &d, nil
}
