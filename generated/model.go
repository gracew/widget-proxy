// A sample object struct definition. This is expected to be replaced by code generated based on user API definitions.

package generated

type Object struct {
	ID        string `json:"id" sql:"type:uuid,default:gen_random_uuid()"`
	CreatedBy string `json:"createdBy"`
	Test      string `json:"test"`
	CreatedAt string `json:"createdAt" sql:"default:now()"`
	// UpdatedAt string `json:"updatedAt"`
}
