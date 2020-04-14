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
<div media py-lg-5>
{{- range .JSAssets.Values }}
    <script src="{{ . }}"></script>
{{- end }}
{{- range .CSSAssets.Values }}
    <link href="{{ . }}" rel="stylesheet">
{{- end }}
{{- template "routers" . }}
{{- template "base" . }}
<style>
    .item {margin: auto; padding-top:30px}
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
	err := http.ListenAndServe(":8080", nil)
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
	if strings.Contains(r.RequestURI, "favicon.ico") {
		http.Redirect(w, r, "https://s3.amazonaws.com/static-assets/default/logo.png", 301)
	} else {
		d := getDataJson(r.RequestURI[1:])
		sw := &strings.Builder{}
		drawChart(d, sw)
		d.Charts = template.HTML(sw.String())
		e := tpl.Lookup("index.gohtml").Execute(w, (d))
		if e != nil {
			log.Error(e.Error())
		}
	}
}

func chart(w http.ResponseWriter, r *http.Request) {
	d := getDataJson(r.RequestURI[7:])
	drawChart(d, w)
}

func drawChart(d *CoronaList, sw io.Writer) {
	totalCases := d.totalCases()
	active := d.active()
	deaths := d.deaths()
	newCases := d.newCases()
	newDeaths := d.newDeaths()
	serious := d.serious()
	stats := d.stats()
	graphTotal := charts.NewLine()
	graphTotal.SetGlobalOptions(charts.TitleOpts{Title: "COVID-19 cases", Subtitle: d.Country},
		charts.LegendOpts{Left: "200px", Top: "5px", TextStyle: charts.TextStyleOpts{FontSize: 12}},
		charts.TooltipOpts{Show: true},
		charts.YAxisOpts{Scale: true, Type: "value"},
	)
	graphTotal.AddXAxis(d.timeSeries()).AddYAxis("Total cases", totalCases.series).AddYAxis("Active cases", active.series,
		charts.RippleEffectOpts{Period: 3, Scale: 6, BrushType: "fill"})
	graphTotal.SetSeriesOptions(
		charts.MLNameTypeItem{Type: "max"},
		charts.LineOpts{Smooth: true},
		charts.MLStyleOpts{Label: charts.LabelTextOpts{Show: true, Formatter: "{a}: {b}"}},
	)
	grapTotalDeaths := charts.NewLine()
	grapTotalDeaths.SetGlobalOptions(charts.TitleOpts{Title: "COVID-19 deaths", Subtitle: d.Country},
		charts.LegendOpts{Left: "200px", Top: "5px", TextStyle: charts.TextStyleOpts{FontSize: 12}},
		charts.TooltipOpts{Show: true},
		charts.YAxisOpts{Scale: true, Type: "value"},
		charts.ColorOpts{"Black"},
	)
	grapTotalDeaths.AddXAxis(d.timeSeries()).AddYAxis("Deaths", deaths.series)
	grapTotalDeaths.SetSeriesOptions(
		charts.MLNameTypeItem{Type: "max"},
		charts.LineOpts{Smooth: true},
		charts.MLStyleOpts{Label: charts.LabelTextOpts{Show: true, Formatter: "{a}: {b}"}},
	)
	graphNewCases := charts.NewLine()
	graphNewCases.SetGlobalOptions(charts.TitleOpts{Title: "COVID-19 New cases", Subtitle: d.Country},
		charts.LegendOpts{Left: "200px", Top: "5px", TextStyle: charts.TextStyleOpts{FontSize: 12}},
		charts.TooltipOpts{Show: true},
		charts.YAxisOpts{Scale: true, Type: "value", SplitArea: charts.SplitAreaOpts{Show: true, AreaStyle: charts.AreaStyleOpts{
			Opacity: 0.75,
		}}},
		charts.XAxisOpts{Scale: true, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true, AreaStyle: charts.AreaStyleOpts{
			Opacity: 0.75,
		}}},
		charts.VisualMapOpts{
			Calculable: false,
			Max:        200,
			Min:        10,
			InRange:    charts.VMInRange{Color: []string{"#50a3ba", "#eac736", "#d94e5d"}}},
	)
	graphNewCases.AddXAxis(d.timeSeries()).
		AddYAxis("New cases", newCases.series, charts.RippleEffectOpts{Period: 3, Scale: 6, BrushType: "fill"})
	graphNewCases.SetSeriesOptions(
		charts.MLNameTypeItem{Type: "max"},
		charts.LineOpts{Smooth: true},
		charts.MLStyleOpts{Label: charts.LabelTextOpts{Show: true, Formatter: "{a}: {b}"}},
	)
	graphNewD := charts.NewEffectScatter()
	graphNewD.SetGlobalOptions(charts.TitleOpts{Title: "COVID-19 New deaths", Subtitle: d.Country},
		charts.LegendOpts{Left: "200px", Top: "5px", TextStyle: charts.TextStyleOpts{FontSize: 12}},
		charts.TooltipOpts{Show: true},
		charts.YAxisOpts{Scale: true, Type: "value", SplitArea: charts.SplitAreaOpts{Show: true, AreaStyle: charts.AreaStyleOpts{
			Opacity: 0.75,
		}}},
		charts.XAxisOpts{Scale: true, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true, AreaStyle: charts.AreaStyleOpts{
			Opacity: 0.75,
		}}},
		charts.ColorOpts{"Black"},
	)
	graphNewD.SetSeriesOptions(
		charts.MLNameTypeItem{Type: "min"},
		charts.MLNameTypeItem{Type: "max"},
		charts.LineOpts{Smooth: true},
		charts.MLStyleOpts{Label: charts.LabelTextOpts{Show: true, Formatter: "{a}: {b}"}},
	)
	graphNewD.AddXAxis(d.timeSeries()).
		AddYAxis("New deaths", newDeaths.series, charts.RippleEffectOpts{Period: 4, Scale: 10, BrushType: "stroke"})
	graphNewDL := charts.NewLine()
	graphNewDL.SetGlobalOptions(charts.TitleOpts{Title: "COVID-19 cases", Subtitle: d.Country},
		charts.LegendOpts{Left: "200px", Top: "5px", TextStyle: charts.TextStyleOpts{FontSize: 12}},
		charts.TooltipOpts{Show: true},
		charts.YAxisOpts{Scale: true, Type: "value"},
	)
	graphNewDL.SetSeriesOptions(
		charts.MLNameTypeItem{Type: "max"},
		charts.MLNameTypeItem{Type: "min"},
		charts.MLNameTypeItem{Type: "average"},
		charts.LineOpts{Smooth: true},
		charts.MLStyleOpts{Label: charts.LabelTextOpts{Show: true, Formatter: "{a}: {b}"}},
	)
	graphNewDL.AddXAxis(d.timeSeries()).
		AddYAxis("", newDeaths.series, charts.RippleEffectOpts{Period: 4, Scale: 10, BrushType: "stroke"})
	graphNewD.Overlap(graphNewDL)
	graphStats := charts.NewEffectScatter()
	graphStats.SetGlobalOptions(charts.TitleOpts{Title: "COVID-19 stats", Subtitle: d.Country},
		charts.LegendOpts{Left: "200px", Top: "5px", TextStyle: charts.TextStyleOpts{FontSize: 12}},
		charts.TooltipOpts{Show: true},
		charts.YAxisOpts{Scale: true, Type: "value", SplitArea: charts.SplitAreaOpts{Show: true, AreaStyle: charts.AreaStyleOpts{
			Opacity: 0.75,
		}}},
		charts.XAxisOpts{Scale: true, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true, AreaStyle: charts.AreaStyleOpts{
			Opacity: 0.75,
		}}},
		charts.ColorOpts{"Black"},
		charts.VisualMapOpts{
			Calculable: false,
			Max:        float32(stats.max) / 10,
			Min:        float32(stats.min) / 10,
			InRange:    charts.VMInRange{Color: []string{"#50a3ba", "#eac736", "#d94e5d"}}},
	)
	graphStats.SetSeriesOptions(
		charts.MLNameTypeItem{Type: "min"},
		charts.MLNameTypeItem{Type: "max"},
		charts.LineOpts{Smooth: true, Stack: "0"},
		charts.MLStyleOpts{Label: charts.LabelTextOpts{Show: true, Formatter: "{a}: {b}"}},
	)
	graphStats.AddXAxis(d.timeSeries()).
		AddYAxis("Per-Million", stats.series, charts.RippleEffectOpts{Period: 10, Scale: 4, BrushType: "fill"})

	graphSerious := charts.NewLine()
	graphSerious.SetGlobalOptions(charts.TitleOpts{Title: "COVID-19 Serious cases", Subtitle: d.Country},
		charts.LegendOpts{Left: "200px", Top: "5px", TextStyle: charts.TextStyleOpts{FontSize: 12}},
		charts.TooltipOpts{Show: true},
		charts.YAxisOpts{Scale: true, Type: "value", SplitArea: charts.SplitAreaOpts{Show: true, AreaStyle: charts.AreaStyleOpts{
			Opacity: 0.75,
		}}},
		charts.XAxisOpts{Type: "category", SplitArea: charts.SplitAreaOpts{Show: true, AreaStyle: charts.AreaStyleOpts{
			Opacity: 0.75,
		}}},

		charts.VisualMapOpts{
			Calculable: false,
			Max:        float32(serious.max) / 10,
			Min:        float32(serious.min) / 10,
			InRange:    charts.VMInRange{Color: []string{"#50a3ba", "#eac736", "#d94e5d"}}},
		charts.ColorOpts{"Orange", "Yellow", "Navy"},
	)

	graphSerious.AddXAxis(d.timeSeries()).AddYAxis("Serious", serious.series,
		charts.LabelTextOpts{Show: true},
	)
	f, e := os.Create("line-" + strconv.Itoa(rand.Int()) + ".html")
	if e == nil {
		defer os.Remove(f.Name())
		graphNewCases.Render(sw, f)
		graphSerious.Render(sw, f)
		graphTotal.Render(sw, f)
		graphStats.Render(sw, f)
		grapTotalDeaths.Render(sw, f)
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
	body, err := ioutil.ReadAll(res.Body)
	if nil != err || len(body) < 1 {
		return readData("Israel", asc)
	}
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
