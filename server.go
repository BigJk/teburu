package teburu

import (
	"fmt"
	"github.com/coocood/freecache"
	cache "github.com/gitsight/go-echo-cache"
	"github.com/iancoleman/strcase"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Server represents the teburu server
type Server struct {
	*echo.Echo

	sheetService *sheets.Service
}

// NewServer creates a new server
func NewServer(sheetService *sheets.Service) *Server {
	return &Server{
		Echo:         echo.New(),
		sheetService: sheetService,
	}
}

func buildColumnsFilter(query string) (map[string]struct{}, error) {
	columns := strings.Split(query, ",")
	columnsFilter := map[string]struct{}{}
	if len(columns) > 0 {
		for _, f := range columns {
			if f == "" {
				continue
			}

			unescapedName, err := url.PathUnescape(f)
			if err != nil {
				return nil, err
			}

			columnsFilter[unescapedName] = struct{}{}
		}
	}
	return columnsFilter, nil
}

func buildCaseFn(format string) func(string) string {
	switch format {
	case "camel":
		return strcase.ToLowerCamel
	case "snake":
		return strcase.ToSnake
	case "kebab":
		return strcase.ToKebab
	case "screaming_snake":
		return strcase.ToScreamingSnake
	case "plain":
		return func(s string) string { return s }
	default:
		return strcase.ToLowerCamel
	}
}

func fetchHeaders(resp *sheets.Spreadsheet) []string {
	var sheetHeaders []string
	for _, cell := range resp.Sheets[0].Data[0].RowData[0].Values {
		if cell.EffectiveValue == nil {
			break
		}
		sheetHeaders = append(sheetHeaders, *cell.EffectiveValue.StringValue)
	}
	return sheetHeaders
}

// Start starts the server
func (s *Server) Start(bind string) error {
	s.Echo.HideBanner = true
	s.Echo.HidePort = true

	api := s.Group("/api")

	api.GET("/sheet/:id/:sheet", func(c echo.Context) error {
		sheet, err := url.PathUnescape(c.Param("sheet"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		resp, err := s.sheetService.Spreadsheets.Get(c.Param("id")).Ranges(fmt.Sprintf("%s!A1:Z", sheet)).Fields(sheetFields...).Do()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		if len(resp.Sheets) == 0 {
			return c.JSON(http.StatusBadRequest, "no data found")
		}

		// Build case function
		caseFn := buildCaseFn(c.QueryParam("case"))

		// Get format function
		format := c.QueryParam("format")
		if format == "" {
			format = "simple"
		}

		// Build columns filter
		columnsFilter, err := buildColumnsFilter(c.QueryParam("columns"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		// Get headers
		sheetHeaders := fetchHeaders(resp)

		// Get data and map to headers
		var data []map[string]interface{}
		for _, row := range resp.Sheets[0].Data[0].RowData[1:] {
			if row.Values == nil {
				break
			}

			rowData := map[string]interface{}{}
			for i, cell := range row.Values {
				if i >= len(sheetHeaders) {
					break
				}
				if len(columnsFilter) > 0 {
					_, okCased := columnsFilter[caseFn(sheetHeaders[i])]
					_, okRaw := columnsFilter[sheetHeaders[i]]
					if !okCased && !okRaw {
						continue
					}
				}
				rowData[caseFn(sheetHeaders[i])] = CollapseCell(cell.EffectiveValue, cell.Hyperlink, CellType(format))
			}

			data = append(data, rowData)
		}

		if c.QueryParam("pretty") == "true" {
			return c.JSONPretty(http.StatusOK, data, "  ")
		}
		return c.JSON(http.StatusOK, data)
	})

	api.GET("/sheet/:id/:sheet/:eid", func(c echo.Context) error {
		sheet, err := url.PathUnescape(c.Param("sheet"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		eid := c.Param("eid")
		resp, err := s.sheetService.Spreadsheets.Get(c.Param("id")).Ranges(
			fmt.Sprintf("%s!A1:Z1", sheet),
			fmt.Sprintf("%s!A%s:Z%s", sheet, eid, eid),
		).Fields(sheetFields...).Do()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		if len(resp.Sheets) == 0 {
			return c.JSON(http.StatusBadRequest, "no data found")
		}

		// Build case function
		caseFn := buildCaseFn(c.QueryParam("case"))

		// Get format function
		format := c.QueryParam("format")
		if format == "" {
			format = "simple"
		}

		// Build columns filter
		columnsFilter, err := buildColumnsFilter(c.QueryParam("columns"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		// Get headers
		sheetHeaders := fetchHeaders(resp)

		row := resp.Sheets[0].Data[1].RowData[0]
		if row.Values == nil {
			return c.JSON(http.StatusBadRequest, "no data found")
		}

		rowData := map[string]interface{}{}
		for i, cell := range row.Values {
			if i >= len(sheetHeaders) {
				break
			}
			if len(columnsFilter) > 0 {
				_, okCased := columnsFilter[caseFn(sheetHeaders[i])]
				_, okRaw := columnsFilter[sheetHeaders[i]]
				if !okCased && !okRaw {
					continue
				}
			}
			rowData[caseFn(sheetHeaders[i])] = CollapseCell(cell.EffectiveValue, cell.Hyperlink, CellType(format))
		}

		if c.QueryParam("pretty") == "true" {
			return c.JSONPretty(http.StatusOK, rowData, "  ")
		}
		return c.JSON(http.StatusOK, rowData)
	})

	return s.Echo.Start(bind)
}

// SetupRateLimit sets up rate limit for all routes
func (s *Server) SetupRateLimit(limit float64) {
	s.Echo.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(limit))))
}

// EnableCORS enables CORS for all routes
func (s *Server) EnableCORS() {
	s.Echo.Use(middleware.CORS())
}

// EnableCaching enables caching for all routes
func (s *Server) EnableCaching(ttl time.Duration) {
	c := freecache.NewCache(1024 * 1024 * 5)
	s.Echo.Use(cache.New(&cache.Config{
		TTL: ttl,
	}, c))
}
