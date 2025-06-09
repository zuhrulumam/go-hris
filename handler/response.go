package handler

type CheckInResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CheckOutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success    bool   `json:"success"`
	HumanError string `json:"human_error"`
	DebugError string `json:"debug_error"`
}

type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
