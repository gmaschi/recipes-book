package recipeModel

type (
	CreateRequest struct {
		Author      string   `json:"author" binding:"required,alphanum"`
		Ingredients []string `json:"ingredients" binding:"required"`
		Steps       []string `json:"steps" binding:"required"`
	}

	GetRequest struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}

	UpdateRequest struct {
		ID          int64    `json:"id" binding:"required"`
		Ingredients []string `json:"ingredients"`
		Steps       []string `json:"steps"`
	}

	DeleteRequest struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}

	ListRequest struct {
		PageID   int32 `form:"page_id" binding:"required,min=1"`
		PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
	}
)
