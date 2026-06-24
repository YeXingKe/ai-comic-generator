package common

const (
	SuccessCode = 0
	ErrorCode   = 50000

	UserRole  = "user"
	AdminRole = "admin"

	SessionUserKey = "loginUser"
)

type BaseResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func Success(data interface{}) BaseResponse {
	return BaseResponse{Code: SuccessCode, Data: data, Message: "ok"}
}

func Error(code int, message string) BaseResponse {
	if code == 0 {
		code = ErrorCode
	}
	return BaseResponse{Code: code, Data: nil, Message: message}
}

type PageResult struct {
	Records  interface{} `json:"records"`
	TotalRow int64       `json:"totalRow"`
	PageNum  int         `json:"pageNum"`
	PageSize int         `json:"pageSize"`
}
