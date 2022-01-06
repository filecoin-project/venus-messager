package controller

var AuthMap = map[string]string{
	"WaitMessage":              "read",
	"GetMessageByUid":          "read",
	"RepublishMessage":         "admin",
	"RecoverFailedMsg":         "admin",
	"GetNode":                  "admin",
	"PushMessage":              "write",
	"GetMessageBySignedCid":    "read",
	"ListMessage":              "admin",
	"ListFailedMessage":        "admin",
	"SaveNode":                 "admin",
	"DeleteNode":               "admin",
	"ReplaceMessage":           "admin",
	"DeleteAddress":            "admin",
	"ForbiddenAddress":         "admin",
	"ClearUnFillMessage":       "admin",
	"GetSharedParams":          "admin",
	"HasMessageByUid":          "read",
	"ListMessageByFromState":   "admin",
	"ListBlockedMessage":       "admin",
	"WalletHas":                "read",
	"ListAddress":              "admin",
	"SetSelectMsgNum":          "admin",
	"ListNode":                 "admin",
	"PushMessageWithId":        "write",
	"GetMessageByUnsignedCid":  "read",
	"ListMessageByAddress":     "admin",
	"GetAddress":               "admin",
	"UpdateNonce":              "admin",
	"ActiveAddress":            "admin",
	"HasNode":                  "admin",
	"GetMessageByFromAndNonce": "read",
	"UpdateMessageStateByID":   "admin",
	"UpdateFilledMessageByID":  "admin",
	"MarkBadMessage":           "admin",
	"SetSharedParams":          "admin",
	"Send":                     "admin",
	"UpdateAllFilledMessage":   "admin",
	"HasAddress":               "read",
	"SetFeeParams":             "admin",
	"SetLogLevel":              "admin",
	"ForcePushMessage":         "admin",
	"ForcePushMessageWithId":   "write",
	"RefreshSharedParams":      "admin",
}
