package authorModel

type (
	CreateRequest struct {
		Username string `json:"username" binding:"required,alphanum"`
		Password string `json:"password" binding:"required,min=6"`
		Email    string `json:"email" binding:"required,email"`
	}

	GetRequest struct {
		Username string `uri:"username" binding:"required,alphanum"`
	}

	UpdateRequest struct {
		Username string `json:"username" binding:"required,alphanum"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	DeleteRequest struct {
		Username string `uri:"username" binding:"required,alphanum"`
	}

	ListRequest struct {
		PageID   int32 `form:"page_id" binding:"required,min=1"`
		PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
	}

	LoginRequest struct {
		Username string `json:"username" binding:"required,alphanum"`
		Password string `json:"password" binding:"required,min=6"`
	}
)
