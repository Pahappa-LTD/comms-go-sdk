package models

type ApiResponse struct {
    Status              ApiResponseCode `json:"Status"`
    Message             string          `json:"Message"`
    Cost                *float64        `json:"Cost"`
    MessageFollowUpCode string          `json:"MsgFollowUpUniqueCode"`
    Balance             *float64        `json:"Balance"`
}

