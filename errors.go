package transbank

import "webtyp.com/fmt"

// Errores de configuración (Open) y de datos; los de la API los arma api.go.
var (
	ErrConfigScheme      = fmt.Err("transbank: config must start with webpay://")
	ErrConfigFormat      = fmt.Err("transbank: config must be webpay://<code>:<key>@<environment>")
	ErrConfigCredentials = fmt.Err("transbank: commerce code and key are required")
	ErrConfigEnvironment = fmt.Err("transbank: environment must be integration or production")
	ErrCurrency          = fmt.Err("transbank: unsupported currency")
	ErrFieldTooLong      = fmt.Err("transbank: field exceeds Webpay length")
	ErrUnknownStatus     = fmt.Err("transbank: unknown status")
	ErrRefundRejected    = fmt.Err("transbank: refund rejected")
	ErrUnknownRefundType = fmt.Err("transbank: unknown refund type")
	ErrBadResponse       = fmt.Err("transbank: bad response")
)
