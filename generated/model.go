package generated

type Object struct {
	ID        string `json:"id" sql:"type:uuid,default:gen_random_uuid()"`
	CreatedBy string `json:"createdBy"`
	Test	  string `json:"test"`
	// CreatedAt string `json:"createdAt" `
	// UpdatedAt string `json:"updatedAt"`
}