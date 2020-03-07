// TODO(gracew): generate this

package model

type Auth struct {
	ID                 string             `json:"id"`
	APIID              string             `json:"apiID"`
	AuthenticationType AuthenticationType `json:"authenticationType"`
	ReadPolicy         *AuthPolicy        `json:"readPolicy"`
	WritePolicy        *AuthPolicy        `json:"writePolicy"`
}

type AuthPolicy struct {
	Type            AuthPolicyType `json:"type"`
	UserAttribute   *string        `json:"userAttribute"`
	ObjectAttribute *string        `json:"objectAttribute"`
}

type AuthPolicyType string

const (
	AuthPolicyTypeCreatedBy      AuthPolicyType = "CREATED_BY"
	AuthPolicyTypeAttributeMatch AuthPolicyType = "ATTRIBUTE_MATCH"
	AuthPolicyTypeCustom         AuthPolicyType = "CUSTOM"
)

var AllAuthPolicyType = []AuthPolicyType{
	AuthPolicyTypeCreatedBy,
	AuthPolicyTypeAttributeMatch,
	AuthPolicyTypeCustom,
}

func (e AuthPolicyType) IsValid() bool {
	switch e {
	case AuthPolicyTypeCreatedBy, AuthPolicyTypeAttributeMatch, AuthPolicyTypeCustom:
		return true
	}
	return false
}

func (e AuthPolicyType) String() string {
	return string(e)
}

type AuthenticationType string

const (
	AuthenticationTypeBuiltIn AuthenticationType = "BUILT_IN"
)

var AllAuthenticationType = []AuthenticationType{
	AuthenticationTypeBuiltIn,
}

func (e AuthenticationType) IsValid() bool {
	switch e {
	case AuthenticationTypeBuiltIn:
		return true
	}
	return false
}

func (e AuthenticationType) String() string {
	return string(e)
}
