package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-echarts/go-echarts/templates"
	"github.com/pingcap/log"
	"io"
	"io/ioutil"
	"math/rand"
	"mime"
	"os"
	"sort"
	"strconv"
	"strings"

	"html/template"
	"net/http"

	"github.com/go-echarts/go-echarts/charts"
)

const pageTemplate = `{{- define "page" }}
	{{- template "header" . }}
	<div>
	{{- template "routers" . }}
	{{- range .Charts }}
	{{ template "base" . }}
	<br/>
	{{- end }}
	<style>
	.container {display: flex;justify-content: center;align-items: center; width:"96%"}
	.item {margin: auto;}
	</style>
	</div>
	{{ end }}`
const chartTemplate = `{{- define "chart" }}
<div class=container-fluid>
{{- range .JSAssets.Values }}
    <script src="{{ . }}"></script>
{{- end }}
{{- range .CSSAssets.Values }}
    <link href="{{ . }}" rel="stylesheet">
{{- end }}
{{- template "routers" . }}
{{- template "base" . }}
<style>
    .item {margin: auto;}
</style>
</div>
{{ end }}
`

var (
	HEADER_ONE       = os.Getenv("header_rapidapi")
	HEADER_ONE_VALUE = os.Getenv("header_rapidapi_value")
	HEADER_TWO       = os.Getenv("header_rapidapi_key")
	HEADER_TWO_VALUE = os.Getenv("header_rapidapi_key_value")
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
		_, e := fmt.Sscan(strings.ReplaceAll(val.NewCasesString, ",", ""), &t)
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
		_, e := fmt.Sscan(strings.ReplaceAll(val.NewDeathsString, ",", ""), &t)
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
		_, e := fmt.Sscan(strings.ReplaceAll(val.TotalDeathsString, ",", ""), &t)
		if e == nil {
			ret[i] = t
		}
		println(t, ret[i])
	}
	return ret
}

//////// VARIABLES /////////
var tpl *template.Template
var countries = Countries{
	AffectedCountries: []string{},
}

//////// FUNCTIONS /////////

func init() {
	templates.PageTpl = pageTemplate
	templates.ChartTpl = chartTemplate
	tpl = template.Must(template.ParseGlob("templates/*.go*html"))
	countriesInit()
}

func countriesInit() {
	url := "https://coronavirus-monitor.p.rapidapi.com/coronavirus/affected.php"
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add(HEADER_ONE, "coronavirus-monitor.p.rapidapi.com")
	req.Header.Add(HEADER_TWO, HEADER_TWO_VALUE)

	res, err := http.DefaultClient.Do(req)
	if nil != err {
		panic(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	log.Info(string(body)[:1])

	json.Unmarshal(body, &countries)
	sort.Strings(countries.AffectedCountries)
}

func main() {
	http.HandleFunc("/download/", download)
	http.HandleFunc("/templates/", load)
	http.HandleFunc("/public/", load)
	http.HandleFunc("/assets/", load)
	http.HandleFunc("/chart/", chart)
	http.HandleFunc("/", index)
	err := http.ListenAndServe(":80", nil)
	if nil != err {
		log.Error(err.Error())
	}
}

func download(writer http.ResponseWriter, request *http.Request) {
	fns := strings.Split(filename(request), "/")
	println("file", fns[len(fns)-1])
	d, _ := readData(fns[len(fns)-1], false)
	b, e := json.Marshal(d.StatsByCountry)
	if e != nil {
		http.Error(writer, e.Error(), http.StatusForbidden)
	}
	writer.Header().Set("Content-Type", "application/json")
	ct := http.DetectContentType(b)
	ebt, e := mime.ExtensionsByType(ct)
	if e != nil {
		http.Error(writer, e.Error(), http.StatusForbidden)
	}
	writer.Header().Set("Content-Disposition", "attachment; filename="+d.Country+ebt[0])
	writer.Write(b)
}

func load(writer http.ResponseWriter, request *http.Request) {
	bs, _ := ioutil.ReadFile(filename(request))
	writer.Write(bs)
}

func filename(request *http.Request) string {
	return fmt.Sprintf("./%s", request.RequestURI)
}

func index(w http.ResponseWriter, r *http.Request) {
	d := getDataJson(r.RequestURI[1:])
	sw := &strings.Builder{}
	drawChart(d, sw)
	d.Charts = template.HTML(sw.String())
	e := tpl.Lookup("index.gohtml").Execute(w, (d))
	if e != nil {
		log.Error(e.Error())
	}
}

func chart(w http.ResponseWriter, r *http.Request) {
	d := getDataJson(r.RequestURI[7:])
	drawChart(d, w)
}

func drawChart(d *CoronaList, sw io.Writer) {
	graphTotal := charts.NewLine()
	graphTotal.SetGlobalOptions(charts.TitleOpts{Title: "COVID-19 cases", Subtitle: d.Country},
		charts.LegendOpts{Left: "200px", Top: "5px", TextStyle: charts.TextStyleOpts{FontSize: 12}},
		charts.TooltipOpts{Show: true},
		charts.YAxisOpts{Scale: true, Type: "value"},
	)
	graphTotal.AddXAxis(d.timeSeries()).AddYAxis("Total cases", d.totalCases())
	grapTotalDeaths := charts.NewBar()
	grapTotalDeaths.SetGlobalOptions(charts.TitleOpts{Title: "COVID-19 deaths", Subtitle: d.Country},
		charts.LegendOpts{Left: "200px", Top: "5px", TextStyle: charts.TextStyleOpts{FontSize: 12}},
		charts.TooltipOpts{Show: true},
		charts.YAxisOpts{Scale: true, Type: "value"},
	)
	grapTotalDeaths.AddXAxis(d.timeSeries()).AddYAxis("Deaths", d.deaths())

	graphNewCases := charts.NewEffectScatter()
	graphNewCases.SetGlobalOptions(charts.TitleOpts{Title: "New cases", Subtitle: d.Country},
		charts.LegendOpts{Left: "200px", Top: "5px", TextStyle: charts.TextStyleOpts{FontSize: 12}},
		charts.TooltipOpts{Show: true},
		charts.YAxisOpts{Scale: true, Type: "value"},
	)
	graphNewCases.AddXAxis(d.timeSeries()).
		AddYAxis("New cases", d.newCases(), charts.RippleEffectOpts{Period: 3, Scale: 6, BrushType: "fill"})

	graphNewD := charts.NewEffectScatter()
	graphNewD.SetGlobalOptions(charts.TitleOpts{Title: "New deaths", Subtitle: d.Country},
		charts.LegendOpts{Left: "200px", Top: "5px", TextStyle: charts.TextStyleOpts{FontSize: 12}},
		charts.TooltipOpts{Show: true},
		charts.YAxisOpts{Scale: true, Type: "value"},
	)
	graphNewD.AddXAxis(d.timeSeries()).
		AddYAxis("New deaths", d.newDeaths(), charts.RippleEffectOpts{Period: 4, Scale: 10, BrushType: "stroke"})
	f, e := os.Create("line-" + strconv.Itoa(rand.Int()) + ".html")
	if e == nil {
		defer os.Remove(f.Name())
		graphTotal.Render(sw, f)
		grapTotalDeaths.Render(sw, f)
		graphNewCases.Render(sw, f)
		graphNewD.Render(sw, f)
	} else {
	}
}

func getDataJson(uri string) *CoronaList {
	log.Info(uri)
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

	req.Header.Add(HEADER_ONE, HEADER_ONE_VALUE)
	req.Header.Add(HEADER_TWO, HEADER_TWO_VALUE)

	res, err := http.DefaultClient.Do(req)
	if nil != err {
		panic(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	log.Info(string(body)[:1])
	d := CoronaList{URL: []string{country}, asc: asc, Countries: countries.AffectedCountries}
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
