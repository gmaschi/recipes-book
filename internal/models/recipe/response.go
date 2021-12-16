package recipe

type CreateResponse struct {
	Author      string   `json:"author" binding:"required,alphanum"`
	Ingredients []string `json:"ingredients" binding:"required"`
	Steps       []string `json:"steps" binding:"required"`
}
