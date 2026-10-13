package transbank

import "webtyp.com/fmt"

var (
	ErrConfigScheme      = fmt.Err("transbank: config must start with webpay://")
	ErrConfigFormat      = fmt.Err("transbank: config format must be webpay://CODE:KEY@ENV")
	ErrConfigCredentials = fmt.Err("transbank: config CODE and KEY cannot be empty")
	ErrConfigEnvironment = fmt.Err("transbank: environment must be integration or production")
)
