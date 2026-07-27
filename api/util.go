package api

import (
	"strconv"

	fiber "github.com/zap-proto/fiber/v3"
	"github.com/zap-proto/zip"
)

const (
	// DefaultPage is the default page to use if none is specified or an invalid value is provided
	DefaultPage = 1

	// DefaultPageSize is the default page size to use if none is specified or an invalid value is provided
	DefaultPageSize = 50
)

// writeJSON is the one way this API answers with JSON. zip.Ctx.JSON would tag
// the body "application/json; charset=utf-8"; every published response has
// always carried bare "application/json", and clients match on it exactly.
func writeJSON(c *zip.Ctx, statusCode int, value any) error {
	c.Status(statusCode)
	return c.Fiber().JSON(value, fiber.MIMEApplicationJSON)
}

func extractPageAndPageSizeFromRequest(c *zip.Ctx, maximumNumberOfResults int) (page, pageSize int) {
	var err error
	if pageParameter := c.Query("page"); len(pageParameter) == 0 {
		page = DefaultPage
	} else {
		page, err = strconv.Atoi(pageParameter)
		if err != nil {
			page = DefaultPage
		}
		if page < 1 {
			page = DefaultPage
		}
	}
	if pageSizeParameter := c.Query("pageSize"); len(pageSizeParameter) == 0 {
		pageSize = DefaultPageSize
	} else {
		pageSize, err = strconv.Atoi(pageSizeParameter)
		if err != nil {
			pageSize = DefaultPageSize
		}
	}
	if page == 1 && pageSize > maximumNumberOfResults {
		// If the page is 1 and the page size is greater than the maximum number of results, return
		// no more than the maximum number of results
		pageSize = maximumNumberOfResults
	} else if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	return
}
