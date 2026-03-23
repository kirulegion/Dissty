package domain

// Provider represents a supported authentication provider on Dissty
type Provider string

const (
    ProviderGoogle Provider = "google"
    ProviderGitHub Provider = "github"
    ProviderEmail  Provider = "email"
    ProviderPhone  Provider = "phone"
)

// IsValid checks if a provider is supported
func (p Provider) IsValid() bool {
    switch p {
    case ProviderGoogle, ProviderGitHub, ProviderEmail, ProviderPhone:
        return true
    }
    return false
}

// IsOTP returns true if this provider uses OTP flow
func (p Provider) IsOTP() bool {
    return p == ProviderEmail || p == ProviderPhone
}

// IsOAuth returns true if this provider uses OAuth flow
func (p Provider) IsOAuth() bool {
    return p == ProviderGoogle || p == ProviderGitHub
}
