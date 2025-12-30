package app

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const DEFAULT_PAGE_SIZE = 10
const MAX_PAGE_SIZE = 100

type PageRequest struct {
	Page     int64 `json:"page" form:"page"`
	PageSize int64 `json:"page_size" form:"page_size"`
}

func NewPageRequest(c *gin.Context) *PageRequest {
	page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 64)
	return &PageRequest{
		Page:     page,
		PageSize: pageSize,
	}
}

func (pr *PageRequest) GetSkip() int64 {
	if pr.Page < 1 {
		pr.Page = 1
	}
	if pr.PageSize < 1 {
		pr.PageSize = DEFAULT_PAGE_SIZE
	}
	if pr.PageSize > MAX_PAGE_SIZE {
		pr.PageSize = MAX_PAGE_SIZE
	}
	return (pr.Page - 1) * pr.PageSize
}

func (pr *PageRequest) GetLimit() int64 {
	if pr.PageSize < 1 {
		pr.PageSize = DEFAULT_PAGE_SIZE
	}
	if pr.PageSize > MAX_PAGE_SIZE {
		pr.PageSize = MAX_PAGE_SIZE
	}
	return pr.PageSize
}

func (pr *PageRequest) IsValid() bool {
	return pr.Page >= 1
}

func (pr *PageRequest) ToFindOptions() *options.FindOptionsBuilder {
	opts := options.Find()
	if pr.IsValid() {
		opts.SetLimit(pr.GetLimit())
		opts.SetSkip(pr.GetSkip())
	}
	return opts
}

type PageResponse[T any] struct {
	TotalCount int64 `json:"total_count"`
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	Items      []T   `json:"items"`
}

func NewPageResponse[T any](totalCount, page, pageSize int64, items []T) *PageResponse[T] {
	return &PageResponse[T]{
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		Items:      items,
	}
}

func NewEmptyPageResponse[T any]() *PageResponse[T] {
	return &PageResponse[T]{
		TotalCount: 0,
		Page:       1,
		PageSize:   0,
		Items:      []T{},
	}
}
