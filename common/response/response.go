package response

import (
	"net/http"
)

type BaseResponsePaging struct {
	Data  interface{} `json:"data"`
	Limit int64       `json:"limit"`
	Page  int64       `json:"page"`
	Total int64       `json:"total"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewBaseResponsePaging(data interface{}, limit int64, page int64, total int64) BaseResponsePaging {
	return BaseResponsePaging{
		Data:  data,
		Limit: limit,
		Page:  page,
		Total: total,
	}
}

func NewBaseResponsePagination(data, limit, offset, total interface{}) (int, interface{}) {
	return http.StatusOK, map[string]interface{}{
		"data":   data,
		"limit":  limit,
		"offset": offset,
		"total":  total,
	}
}

func NewBaseResponseScroll(data interface{}, scrollId string) BaseResponseScroll {
	return BaseResponseScroll{
		Items:    data,
		ScrollId: scrollId,
	}
}

func NewResponse(code int, data interface{}) (int, interface{}) {
	return code, data
}

func NewOKResponse(data interface{}) (int, interface{}) {
	return http.StatusOK, map[string]interface{}{
		"data": data,
	}
}

func NewCreatedResponse(data map[string]interface{}) (int, interface{}) {
	result := map[string]interface{}{
		"created": true,
	}
	for key, value := range data {
		result[key] = value
	}
	return http.StatusCreated, result
}
func NewErrorResponse(code int, msg interface{}) (int, interface{}) {
	return code, map[string]interface{}{
		"error": msg,
	}
}

func ServiceUnavailable() (int, interface{}) {
	return http.StatusServiceUnavailable, map[string]interface{}{
		"error": http.StatusText(http.StatusServiceUnavailable),
	}
}

func ServiceUnavailableMsg(msg interface{}) (int, interface{}) {
	return http.StatusServiceUnavailable, map[string]interface{}{
		"error": msg,
	}
}

func BadRequest() (int, interface{}) {
	return http.StatusBadRequest, map[string]interface{}{
		"error": http.StatusText(http.StatusBadRequest),
	}
}

func BadRequestMsg(msg interface{}) (int, interface{}) {
	return http.StatusBadRequest, map[string]interface{}{
		"error": msg,
	}
}

func NotFound() (int, interface{}) {
	return http.StatusNotFound, map[string]interface{}{
		"error": http.StatusText(http.StatusNotFound),
	}
}

func NotFoundMsg(msg interface{}) (int, interface{}) {
	return http.StatusNotFound, map[string]interface{}{
		"error": msg,
	}
}

func Forbidden() (int, interface{}) {
	return http.StatusForbidden, map[string]interface{}{
		"error": "Do not have permission for the request.",
	}
}

func Unauthorized() (int, interface{}) {
	return http.StatusUnauthorized, map[string]interface{}{
		"error": http.StatusText(http.StatusUnauthorized),
	}
}

type BaseResponseScroll struct {
	Items    interface{} `json:"data"`
	ScrollId string      `json:"scroll_id"`
}
