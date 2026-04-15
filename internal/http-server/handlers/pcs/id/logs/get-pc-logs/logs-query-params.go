package getPcLogs

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type LogsQueryParams struct {
	Order  string
	Limit  int32
	Cursor uuid.UUID
	Type   string
}

func parseLogsQueryParams(r *http.Request) (*LogsQueryParams, error) {
	order := r.URL.Query().Get("order")
	if order == "" {
		order = "asc"
	}
	if order != "asc" && order != "desc" {
		return nil, fmt.Errorf(`order does not match "asc" or "desc"`)
	}

	limit := int32(20)
	if l := r.URL.Query().Get("limit"); l != "" {
		val, err := strconv.ParseInt(l, 10, 32)
		if err != nil || val <= 0 {
			return nil, fmt.Errorf("limit must be positive integer number")
		}
		limit = int32(val)
	}

	before := r.URL.Query().Get("before")
	after := r.URL.Query().Get("after")
	if before != "" && after != "" {
		return nil, fmt.Errorf(`request can not contain "before" and "after" in one query`)
	}

	var cursorType string
	var cursorID uuid.UUID
	var err error

	switch {
	case after != "":
		cursorType = "after"
		cursorID, err = uuid.Parse(after)
	case before != "":
		cursorType = "before"
		cursorID, err = uuid.Parse(before)
	default:
		cursorType = "first"
	}
	if err != nil {
		return nil, fmt.Errorf("invalid cursor UUID")
	}

	return &LogsQueryParams{
		Order:  order,
		Limit:  limit,
		Cursor: cursorID,
		Type:   cursorType,
	}, nil
}
