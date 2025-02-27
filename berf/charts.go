package berf

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strings"
	"text/template"
	"time"

	cors "github.com/AdhityaRamadhanus/fasthttpcors"
	util2 "github.com/bingoohuang/ngg/berf/pkg/util"
	"github.com/bingoohuang/ngg/berf/plugins"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/templates"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/valyala/fasthttp"
)

const (
	assetsPath = "/echarts/statics/"
)

const (
	ViewTpl = `
$(function () {
{{ .SetInterval }}(views_sync, {{ .Interval }}); });
let views = {{.ViewsMap}};
{{.ViewSyncJS}}
`
	PageTpl = `
{{- define "page" }}
<!DOCTYPE html>
<html>
	{{- template "header" . }}
<body>
<p align="center">🚀 <a href="https://github.com/bingoohuang/ngg/berf"><b>Berf</b></a> %s</p>
<style> .box { justify-content:center; display:flex; flex-wrap:wrap } </style>
<div class="box"> {{- range .Charts }} {{ template "base" . }} {{- end }} </div>
</body>
</html>
{{ end }}
`
)

//go:embed res/views_sync.js
var viewSyncJs string

func (c *Views) genViewTemplate(viewChartsMap map[string]string) string {
	tpl, err := template.New("view").Parse(ViewTpl)
	if err != nil {
		panic("failed to parse template " + err.Error())
	}

	viewsMap := "{"
	for k, v := range viewChartsMap {
		viewsMap += k + ": goecharts_" + v + ","
	}
	viewsMap += "noop: null}"

	d := struct {
		ViewsMap    string
		SetInterval string
		ViewSyncJS  string
		Interval    int
	}{
		Interval:    int(time.Second.Milliseconds()),
		ViewsMap:    viewsMap,
		SetInterval: "setInterval",
		ViewSyncJS:  viewSyncJs,
	}

	if c.dryPlots {
		d.Interval = 1
		d.SetInterval = "setTimeout"
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, d); err != nil {
		panic("failed to execute template " + err.Error())
	}

	return buf.String()
}

type Views struct {
	viewChartsMap map[string]string
	size          util2.WidthHeight
	dryPlots      bool
	num           int
}

func NewViews(size string, dryPlots bool) *Views {
	return &Views{
		viewChartsMap: make(map[string]string),
		size:          util2.ParseWidthHeight(size, 500, 300),
		dryPlots:      dryPlots,
	}
}

func New[T any](t T) *T {
	return &t
}

func (c *Views) newBasicView(route string) *charts.Line {
	g := charts.NewLine()
	g.SetGlobalOptions(charts.WithTooltipOpts(opts.Tooltip{Show: New(true), Trigger: "axis"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Time"}),
		charts.WithInitializationOpts(opts.Initialization{Width: c.size.WidthPx(), Height: c.size.HeightPx()}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", XAxisIndex: []int{0}}),
	)
	g.SetXAxis([]string{}).SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: New(true)}))

	c.viewChartsMap[route] = g.ChartID
	if len(c.viewChartsMap) == c.num {
		g.AddJSFuncs(c.genViewTemplate(c.viewChartsMap))
	}
	return g
}

var titleCn = map[string]string{
	"tps":               "吞吐量",
	"latency":           "延时",
	"latencypercentile": "百分位延时",
	"concurrent":        "并发",
	"procstat":          "进程",
	"mem":               "内存",
	"netstat":           "网络",
	"disk":              "磁盘",
	"diskio":            "磁盘IO",
	"net":               "网卡",
	"system":            "系统",
}

func title(name string) string {
	if t := titleCn[strings.ToLower(name)]; t != "" {
		return t
	}

	return ss.ToCamel(name)
}

func (c *Views) newView(name, unit string, series plugins.Series) components.Charter {
	selected := map[string]bool{}
	for _, p := range series.Series {
		selected[p] = len(series.Series) == 1 || ss.AnyOf(p, series.Selected...)
	}

	g := c.newBasicView(name)
	var axisLabel *opts.AxisLabel
	if unit != "" {
		axisLabel = &opts.AxisLabel{Formatter: types.FuncStr("{value} " + unit)}
	}
	g.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title(name)}),
		charts.WithYAxisOpts(opts.YAxis{Scale: New(true), AxisLabel: axisLabel}),
		charts.WithLegendOpts(opts.Legend{Show: New(true), Selected: selected}),
	)

	for _, p := range series.Series {
		g.AddSeries(p, []opts.LineData{})
	}
	return g
}

func (c *Views) newHardwareViews(charts *Charts) (charters []components.Charter) {
	for _, name := range charts.hardwaresNames {
		input := charts.hardwares[name]
		charters = append(charters, c.newView(name, "", input.Series()))
	}

	return
}

func (c *Views) newLatencyView() components.Charter {
	return c.newView("latency", "ms", plugins.Series{
		Series: []string{"Min", "Mean", "StdDev", "Max"}, Selected: []string{"Mean"},
	})
}

func (c *Views) newLatencyPercentileView() components.Charter {
	return c.newView("latencyPercentile", "ms", plugins.Series{
		Series: []string{"P50", "P75", "P90", "P95", "P99", "P99.9", "P99.99"}, Selected: []string{"P50", "P90", "P99"},
	})
}

func (c *Views) newConcurrentView() components.Charter {
	return c.newView("concurrent", "", plugins.Series{Series: []string{"Concurrent"}})
}

func (c *Views) newTPSView() components.Charter {
	series := []string{"TPS", "TPS-0"}
	return c.newView("tps", "", plugins.Series{Series: series, Selected: series})
}

type Charts struct {
	chartsData func() *ChartsReport
	config     *Config

	hardwares map[string]plugins.Input

	hardwaresNames []string
}

func NewCharts(chartsData func() *ChartsReport, config *Config) *Charts {
	templates.PageTpl = fmt.Sprintf(PageTpl, config.Desc)
	c := &Charts{chartsData: chartsData, config: config}
	c.initHardwareCollectors()
	return c
}

func (c *Charts) initHardwareCollectors() {
	var keys []string
	for k, _ := range plugins.Inputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	c.hardwaresNames = keys
	c.hardwares = map[string]plugins.Input{}
	for _, name := range c.hardwaresNames {
		inputFn := plugins.Inputs[name]
		input := inputFn()
		if init, ok := input.(plugins.Initializer); ok {
			if err := init.Init(); err != nil {
				log.Printf("plugin %s init failed: %v", name, err)
			}
		}

		c.hardwares[name] = input
	}
}

//go:embed res/echarts.min.js res/jquery.min.js
var assetsFS embed.FS

func (c *Charts) Handler(ctx *fasthttp.RequestCtx) {
	switch path := string(ctx.Path()); {
	case path == "/data/":
		ctx.SetContentType(`application/json; charset=utf-8`)
		_, _ = ctx.Write(c.handleData())
	case path == "/":
		ctx.SetContentType("text/html")
		size := ctx.QueryArgs().Peek("size")
		views := ctx.QueryArgs().Peek("views")
		c.renderCharts(ctx, string(size), string(views))
	case strings.HasPrefix(path, assetsPath):
		name := "res/" + path[len(assetsPath):]
		if f, err := assetsFS.Open(name); err != nil {
			ctx.Error(err.Error(), 404)
		} else {
			ctx.SetBodyStream(f, -1)
		}
	default:
		ctx.Error("NotFound", 404)
	}
}

func (c *Charts) mergeHardwareMetrics(s []byte) []byte {
	for _, name := range c.hardwaresNames {
		if d, err := c.hardwares[name].Gather(); err != nil {
			log.Printf("E! failed to gather %s error: %v", name, err)
		} else {
			if s, err = jj.SetBytes(s, "values."+name, d); err != nil {
				log.Printf("E! failed to set %s error: %v", name, err)
			}
		}
	}

	return s
}

func (c *Charts) handleData() []byte {
	if c.config.IsDryPlots() {
		if d := c.config.PlotsHandle.ReadAll(); len(d) > 0 {
			return d
		}
		return []byte("[]")
	}

	rd := c.chartsData()
	plots := createMetrics(rd, c.config.IsNop())
	plots = c.mergeHardwareMetrics(plots)

	return []byte(("[" + string(plots) + "]"))
}

func (c *Charts) renderCharts(w io.Writer, size, viewsArg string) {
	v := NewViews(size, c.config.IsDryPlots())
	var fns []func() components.Charter

	if !c.config.IsNop() && !Demo {
		if views := util2.NewFeatures(viewsArg); len(views) == 0 {
			fns = append(fns, v.newLatencyView, v.newTPSView, v.newLatencyPercentileView)
			if !c.config.Incr.IsEmpty() || c.config.IsDryPlots() {
				fns = append(fns, v.newConcurrentView)
			}
		} else {
			if views.HasAny("latency", "l") {
				fns = append(fns, v.newLatencyView)
			}
			if views.HasAny("tps", "r") {
				fns = append(fns, v.newTPSView)
			}
			if views.HasAny("latencypercentile", "lp") {
				fns = append(fns, v.newLatencyPercentileView)
			}
			if views.HasAny("concurrent", "c") {
				fns = append(fns, v.newConcurrentView)
			}
		}
	}

	v.num = len(fns) + len(plugins.Inputs)
	p := components.NewPage()
	p.PageTitle = "berf"
	p.AssetsHost = assetsPath
	p.Assets.JSAssets.Add("jquery.min.js")

	for _, vf := range fns {
		p.AddCharts(vf())
	}
	p.AddCharts(v.newHardwareViews(c)...)
	_ = p.Render(w)
}

func (c *Charts) Serve(ln net.Listener, port int) {
	server := fasthttp.Server{
		Handler: cors.DefaultHandler().CorsMiddleware(c.Handler),
	}

	if c.config.IsDryPlots() {
		log.Printf("Running in dry mode for %s", c.config.PlotsFile)
		go ss.OpenInBrowser(fmt.Sprintf("http://127.0.0.1:%d", port))
		ss.ExitIfErr(server.Serve(ln))
		return
	}

	go func() {
		time.Sleep(3 * time.Second) // 3s之后再弹出，避免运行时间过短，程序已经退出
		go ss.OpenInBrowser(fmt.Sprintf("http://127.0.0.1:%d", port))
		ss.ExitIfErr(server.Serve(ln))
	}()
}
