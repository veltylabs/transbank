package transbank

// EnvWebpay es la variable de entorno con la configuración:
// webpay://<código de comercio>:<llave secreta>@integration|production
const EnvWebpay = "TRANSBANK_WEBPAY"

const (
	EnvironmentIntegration = "integration"
	EnvironmentProduction  = "production"

	IntegrationBaseURL = "https://webpay3gint.transbank.cl"
	ProductionBaseURL  = "https://webpay3g.transbank.cl"

	// Credenciales públicas del ambiente de integración de Transbank Developers.
	IntegrationCommerceCode = "597055555532"
	IntegrationAPIKey       = "579B532A7440BB0C9079DED94D31EA1615BACEB56610332264630D42D0A36B1C"

	// DefaultTimeoutMs es el tiempo máximo de una llamada a la API.
	DefaultTimeoutMs = 30000
)

const configScheme = "webpay://"

// Config es la configuración de Webpay; Open la arma desde EnvWebpay.
type Config struct {
	CommerceCode string
	APIKey       string
	BaseURL      string // IntegrationBaseURL o ProductionBaseURL (otra solo en pruebas)
	TimeoutMs    int    // 0 = DefaultTimeoutMs
	ReturnURL    string // URL absoluta y pública de ReturnPath (https://<sitio>/payments/webpay/return)
}

// Webpay implementa paymentmanager.Gateway sobre la API REST v1.2 de Webpay Plus.
type Webpay struct {
	cfg Config
}

// Open interpreta el valor de EnvWebpay. No llama a Transbank. Ningún error
// incluye la llave.
// Ejemplo: wp, err := transbank.Open(env, "https://example.com/payments/webpay/return")
func Open(raw, returnURL string) (*Webpay, error) {
	if len(raw) < len(configScheme) || raw[:len(configScheme)] != configScheme {
		return nil, ErrConfigScheme
	}
	rest := raw[len(configScheme):]
	at := lastIndex(rest, '@')
	colon := firstIndex(rest, ':')
	if at < 0 || colon < 0 || colon > at {
		return nil, ErrConfigFormat
	}
	code, key, env := rest[:colon], rest[colon+1:at], rest[at+1:]
	if code == "" || key == "" {
		return nil, ErrConfigCredentials
	}
	var base string
	switch env {
	case EnvironmentIntegration:
		base = IntegrationBaseURL
	case EnvironmentProduction:
		base = ProductionBaseURL
	default:
		return nil, ErrConfigEnvironment
	}
	return New(Config{CommerceCode: code, APIKey: key, BaseURL: base, ReturnURL: returnURL}), nil
}

// New crea el Gateway con la configuración dada.
// Ejemplo: wp := transbank.New(transbank.Config{ReturnURL: "https://example.com/payments/webpay/return", ...})
func New(cfg Config) *Webpay {
	if cfg.TimeoutMs == 0 {
		cfg.TimeoutMs = DefaultTimeoutMs
	}
	return &Webpay{cfg: cfg}
}

func firstIndex(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func lastIndex(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}
