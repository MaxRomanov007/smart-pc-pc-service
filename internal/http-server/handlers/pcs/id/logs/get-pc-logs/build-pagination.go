package getPcLogs

import (
	"github.com/MaxRomanov007/smart-pc-go-lib/api/response/pagination"
	"github.com/MaxRomanov007/smart-pc-go-lib/domain/models"
)

type PaginationResult struct {
	Logs []models.PcLog
	Pag  *pagination.Pagination
}

func buildPagination(logs []models.PcLog, params *LogsQueryParams, total int64) PaginationResult {
	hasExtra := len(logs) > int(params.Limit)
	if hasExtra {
		logs = logs[:params.Limit]
	}

	var hasPrev, hasNext bool
	switch params.Type {
	case "after":
		hasPrev, hasNext = true, hasExtra
	case "before":
		hasPrev, hasNext = hasExtra, true
	default:
		hasPrev, hasNext = false, hasExtra
	}

	pag := &pagination.Pagination{
		Total:   total,
		HasPrev: hasPrev,
		HasNext: hasNext,
	}

	if len(logs) > 0 {
		if hasPrev {
			pag.PrevCursor = logs[0].ID.String()
		}
		if hasNext {
			pag.NextCursor = logs[len(logs)-1].ID.String()
		}
	}

	return PaginationResult{Logs: logs, Pag: pag}
}
