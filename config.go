package transbank

const IntegrationBaseURL = "https://webpay3gint.transbank.cl"
const ProductionBaseURL = "https://webpay3g.transbank.cl"

type Config struct {
	CommerceCode string
	APIKey       string
	BaseURL      string
	TimeoutMs    int
}

func Open(s string) (*Webpay, error) {
	if len(s) < 9 || s[:9] != "webpay://" {
		return nil, ErrConfigScheme
	}
	s = s[9:]

	envIdx := -1
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '@' {
			envIdx = i
			break
		}
	}
	if envIdx == -1 {
		return nil, ErrConfigFormat
	}

	creds := s[:envIdx]
	env := s[envIdx+1:]

	colonIdx := -1
	for i := 0; i < len(creds); i++ {
		if creds[i] == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx == -1 {
		return nil, ErrConfigFormat
	}

	code := creds[:colonIdx]
	key := creds[colonIdx+1:]

	if code == "" || key == "" {
		return nil, ErrConfigCredentials
	}

	baseURL := ""
	if env == "integration" {
		baseURL = IntegrationBaseURL
	} else if env == "production" {
		baseURL = ProductionBaseURL
	} else {
		return nil, ErrConfigEnvironment
	}

	return New(Config{
		CommerceCode: code,
		APIKey:       key,
		BaseURL:      baseURL,
	}), nil
}

func New(c Config) *Webpay {
	if c.TimeoutMs == 0 {
		c.TimeoutMs = 15000
	}
	return &Webpay{cfg: c}
}
