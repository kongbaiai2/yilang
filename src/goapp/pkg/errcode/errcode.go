package errcode

type Err struct {
	Code int
	Msg  string
}

func (e *Err) Error() string { return e.Msg }

var (
	StatusSuccess                 = &Err{Code: 200, Msg: "successful"}
	ErrorDBSelect                 = &Err{Code: -500001, Msg: "cidripexist"}
	ErrorCidrFormat               = &Err{Code: -500002, Msg: "cidrformaterrors"}
	ErrorDBDelete                 = &Err{Code: -500003, Msg: "ErrorDBDelete"}
	ErrorParameters               = &Err{Code: -500004, Msg: "ErrorParameters"}
	ErrorOpenApiError             = &Err{Code: -500005, Msg: "ErrorOpenApiError"}
	ErrorAuthentication           = &Err{Code: -500006, Msg: "ErrorAuthentication"}
	ErrorMethodNotSupport         = &Err{Code: -500007, Msg: "ErrorMethodNotSupport"}
	ErrorUserCredentials          = &Err{Code: -500008, Msg: "ErrorUserCredentials"}
	ErrorUnmarshalJSON            = &Err{Code: -500009, Msg: "ErrorUnmarshalJSON"}
	ErrorTxCommit                 = &Err{Code: -500010, Msg: "ErrorTxCommit"}
	ErrorTenantNotFound           = &Err{Code: -500011, Msg: "ErrorTenantNotFound"}
	ErrorSingletunnelExist        = &Err{Code: -500012, Msg: "ErrorSingletunnelExist"}
	ErrorRsPoolNotFound           = &Err{Code: -500013, Msg: "ErrorRsPoolNotFound"}
	ErrorMarshalJSON              = &Err{Code: -500014, Msg: "ErrorMarshalJSON"}
	ErrorLogstoreInUse            = &Err{Code: -500015, Msg: "ErrorLogstoreInUse"}
	ErrorInternalError            = &Err{Code: -500016, Msg: "ErrorInternalError"}
	ErrorDBUpdate                 = &Err{Code: -500017, Msg: "ErrorDBUpdate"}
	ErrorDBNoRow                  = &Err{Code: -500018, Msg: "ErrorDBNoRow"}
	ErrorDBInsert                 = &Err{Code: -500019, Msg: "ErrorDBInsert"}
	ErrorClusterWitchSameResGroup = &Err{Code: -500020, Msg: "ErrorClusterWitchSameResGroup"}
	ErrorClusterNotFound          = &Err{Code: -500021, Msg: "ErrorClusterNotFound"}
	ErrorFirewallAlreadyEnabled   = &Err{Code: -500022, Msg: "ErrorFirewallAlreadyEnabled"}
	ErrorDefaultRsPoolNotFound    = &Err{Code: -500023, Msg: "ErrorDefaultRsPoolNotFound"}
	ErrorFirewallNotEnabled       = &Err{Code: -500024, Msg: "ErrorFirewallNotEnabled"}
	ErrorUserRsPoolExist          = &Err{Code: -500025, Msg: "ErrorUserRsPoolExist"}
	ErrorParametersRsPoolBiz      = &Err{Code: -500026, Msg: "ErrorParametersRsPoolBiz"}
	ErrorParametersFirewallType   = &Err{Code: -500027, Msg: "ErrorParametersFirewallType"}
	ErrorRsPoolHasAssignments     = &Err{Code: -500028, Msg: "ErrorRsPoolHasAssignments"}
)
var ErrMsg = map[int]*Err{
	StatusSuccess.Code:                 StatusSuccess,
	ErrorDBSelect.Code:                 ErrorDBSelect,
	ErrorCidrFormat.Code:               ErrorCidrFormat,
	ErrorDBDelete.Code:                 ErrorDBDelete,
	ErrorParameters.Code:               ErrorParameters,
	ErrorOpenApiError.Code:             ErrorOpenApiError,
	ErrorAuthentication.Code:           ErrorAuthentication,
	ErrorMethodNotSupport.Code:         ErrorMethodNotSupport,
	ErrorUserCredentials.Code:          ErrorUserCredentials,
	ErrorUnmarshalJSON.Code:            ErrorUnmarshalJSON,
	ErrorTxCommit.Code:                 ErrorTxCommit,
	ErrorTenantNotFound.Code:           ErrorTenantNotFound,
	ErrorSingletunnelExist.Code:        ErrorSingletunnelExist,
	ErrorRsPoolNotFound.Code:           ErrorRsPoolNotFound,
	ErrorMarshalJSON.Code:              ErrorMarshalJSON,
	ErrorLogstoreInUse.Code:            ErrorLogstoreInUse,
	ErrorInternalError.Code:            ErrorInternalError,
	ErrorDBUpdate.Code:                 ErrorDBUpdate,
	ErrorDBNoRow.Code:                  ErrorDBNoRow,
	ErrorDBInsert.Code:                 ErrorDBInsert,
	ErrorClusterWitchSameResGroup.Code: ErrorClusterWitchSameResGroup,
	ErrorClusterNotFound.Code:          ErrorClusterNotFound,
	ErrorFirewallAlreadyEnabled.Code:   ErrorFirewallAlreadyEnabled,
	ErrorDefaultRsPoolNotFound.Code:    ErrorDefaultRsPoolNotFound,
	ErrorFirewallNotEnabled.Code:       ErrorFirewallNotEnabled,
	ErrorUserRsPoolExist.Code:          ErrorUserRsPoolExist,
	ErrorParametersRsPoolBiz.Code:      ErrorParametersRsPoolBiz,
	ErrorParametersFirewallType.Code:   ErrorParametersFirewallType,
	ErrorRsPoolHasAssignments.Code:     ErrorRsPoolHasAssignments,
	ErrorRsPoolHasAssignments.Code:     ErrorRsPoolHasAssignments,
	ErrorParametersRsPoolBiz.Code:      ErrorParametersRsPoolBiz,
}
