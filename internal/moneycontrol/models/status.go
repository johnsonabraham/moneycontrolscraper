package models

type HC struct {
	App string `json:"app"`
}

type Status struct {
	AppVersion    string `json:"app_version"`
	IsDBConnected bool   `json:"is_db_connected"`
}

type Response struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Vid    string `json:"vid"`
}

type FailedResponse struct {
	Status   int    `json:"status"`
	ErrorMsg string `json:"error_msg"`
}
