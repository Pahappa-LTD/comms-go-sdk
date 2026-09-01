package models

type ApiRequest struct {
    Method      string         `json:"method"`
    Userdata    UserData       `json:"userdata"`
    MessageData []MessageModel `json:"msgdata"`
    WalletType  WalletType     `json:"walletType"`
}
