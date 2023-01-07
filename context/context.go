package context

import (
	"strconv"

	session "github.com/ipfans/echo-session"
	"github.com/labstack/echo"
)

type GlobalContext struct {
	echo.Context
	Session session.Session
}

type ShopContext struct {
	echo.Context
	ResponseContext Response
}

func InitContext(g *GlobalContext) *ShopContext {
	wg := &ShopContext{Context: g}

	wg.ResponseContext = Response{
		Data: echo.Map{},
	}

	return wg
}

type Response struct {
	SuccessMessage string      `json:"success_message"`
	ErrorMessage   string      `json:"error_message"`
	StatusCode     int         `json:"status_code"`
	Data           interface{} `json:"data"`
	Metadata       Metadata    `json:"metadata"`
}

type Metadata struct {
	Limit       int    `json:"limit"`
	CurrentPage int    `json:"current_page"`
	TotalPages  int    `json:"total_pages"`
	TotalCounts int    `json:"total_counts"`
	Sort        string `json:"sort"`
}

type PublicFilter struct {
	Limit int    `json:"limit"`
	Page  int    `json:"page"`
	Sort  string `json:"sort"`
}

func SetFilter(limit, page, sort string) PublicFilter {
	l, _ := strconv.Atoi(limit)
	if l < 1 || l > 100 {
		l = 10
	}

	p, _ := strconv.Atoi(page)
	if p < 1 {
		p = 1
	}

	return PublicFilter{
		Limit: l,
		Page:  p,
		Sort:  sort,
	}
}
