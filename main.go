package main

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"strings"

	"html/template"
	"net/http"
)

type record struct {
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
type datum struct {
	Country        string   `json:"country,unknown"`
	StatsByCountry []record `json:"stat_by_country,unknown"`
	URL            []string
	sb             string
	asc            bool
}

func (d *datum) Len() int { return len(d.StatsByCountry) }
func (d *datum) Swap(i, j int) {
	temp := d.StatsByCountry[i]
	d.StatsByCountry[i] = d.StatsByCountry[j]
	d.StatsByCountry[j] = temp
}
func (d *datum) Less(i, j int) bool {
	return d.asc == (strings.Compare(d.StatsByCountry[i].RecordDate, d.StatsByCountry[j].RecordDate) < 0)
}

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

func main() {
	http.HandleFunc("/", index)
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	country := "Israel"
	splitten := strings.Split(r.RequestURI[1:], "/")
	if len(splitten) > 0 {
		country = splitten[0]
	}
	asc := true
	if len(splitten) > 1 {
		asc = splitten[1] != "desc"
	}
	d, err := readData(country, asc)
	if err != nil {
		panic(err)
	}
	tpl.ExecuteTemplate(w, "index.gohtml", d)
}

func readData(country string, asc bool) (*datum, error) {
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
	d := datum{URL: []string{country}, asc: asc}
	json.Unmarshal(body, &d)

	sort.Sort(&d)
	return &d, nil
}
