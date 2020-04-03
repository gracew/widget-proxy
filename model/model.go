// TODO(gracew): generate this

package model

type API struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Operations *OperationDefinition `json:"operations"`
}

type OperationDefinition struct {
	List   *ListDefinition   `json:"list"`
	Update *UpdateDefinition `json:"update"`
}

type ListDefinition struct {
	Enabled bool             `json:"enabled"`
	Sort    []SortDefinition `json:"sort"`
	Filter  []string         `json:"filter"`
}

type SortDefinition struct {
	Field string    `json:"field"`
	Order SortOrder `json:"order"`
}

type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

var AllSortOrder = []SortOrder{
	SortOrderAsc,
	SortOrderDesc,
}

func (e SortOrder) IsValid() bool {
	switch e {
	case SortOrderAsc, SortOrderDesc:
		return true
	}
	return false
}

func (e SortOrder) String() string {
	return string(e)
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
