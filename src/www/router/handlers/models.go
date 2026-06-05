package handlers

type errorResponse struct {
	Code   int32  `json:"code"`
	Status string `json:"status"`
	Msg    string `json:"msg"`
}