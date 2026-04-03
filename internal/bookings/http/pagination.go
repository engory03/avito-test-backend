package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseBookingListPagination: page по умолчанию 1, pageSize — 20; при ошибке валидации ok == false.
func parseBookingListPagination(c *gin.Context) (page, pageSize int, ok bool) {
	page = 1
	pageSize = 20
	if s := c.Query("page"); s != "" {
		p, err := strconv.Atoi(s)
		if err != nil || p < 1 {
			return 0, 0, false
		}
		page = p
	}
	if s := c.Query("pageSize"); s != "" {
		ps, err := strconv.Atoi(s)
		if err != nil || ps < 1 || ps > 100 {
			return 0, 0, false
		}
		pageSize = ps
	}
	return page, pageSize, true
}
