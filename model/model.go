// TODO(gracew): generate this

package model

type API struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Operations *OperationDefinition `json:"operations"`
}

type OperationDefinition struct {
	Update *UpdateDefinition `json:"update"`
}

type UpdateDefinition struct {
	Actions []ActionDefinition `json:"actions"`
}

type ActionDefinition struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
}

type Auth struct {
	APIID  string                 `json:"apiID"`
	Read   *AuthPolicy            `json:"read"`
	Update map[string]*AuthPolicy `json:"update"`
	Delete *AuthPolicy            `json:"delete"`
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

type CustomLogic struct {
	Before *string `json:"before"`
	After  *string `json:"after"`
}

type AllCustomLogic struct {
	APIID  string                  `json:"apiID" sql:",pk"`
	Create *CustomLogic            `json:"create"`
	Update map[string]*CustomLogic `json:"update"`
	Delete *CustomLogic            `json:"delete"`
}
