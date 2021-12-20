package recipeController_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gmaschi/go-recipes-book/internal/factories/book-recipe-factory"
	mockedstore "github.com/gmaschi/go-recipes-book/internal/mocks/datastore/postgresql/recipes"
	recipeModel "github.com/gmaschi/go-recipes-book/internal/models/recipe"
	db "github.com/gmaschi/go-recipes-book/internal/services/datastore/postgresql/recipes/sqlc"
	"github.com/gmaschi/go-recipes-book/pkg/tools/password"
	"github.com/gmaschi/go-recipes-book/pkg/tools/random"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type eqUpdateRecipesMatcher struct {
	arg db.UpdateRecipeParams
}

func (e eqUpdateRecipesMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.UpdateRecipeParams)
	if !ok {
		return false
	}
	if time.Since(arg.UpdatedAt) < time.Second {
		e.arg.UpdatedAt = arg.UpdatedAt
	}
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqUpdateRecipesMatcher) String() string {
	return fmt.Sprintf("arg: %v", e.arg)
}

func EqUpdateRecipesParams(arg db.UpdateRecipeParams) gomock.Matcher {
	return eqUpdateRecipesMatcher{arg: arg}
}

func TestCreate(t *testing.T) {
	recipe := randomRecipe(t)
	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockedstore.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"author":      recipe.Author,
				"ingredients": recipe.Ingredients,
				"steps":       recipe.Steps,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				createArg := db.CreateRecipeParams{
					Author:      recipe.Author,
					Ingredients: recipe.Ingredients,
					Steps:       recipe.Steps,
				}
				store.EXPECT().
					CreateRecipe(gomock.Any(), gomock.Eq(createArg)).
					Times(1).
					Return(recipe, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCreate(t, recorder.Body, recipe)
			},
		},
		{
			name: "InvalidAuthorName",
			body: map[string]interface{}{
				"author":      "invalid-author#",
				"ingredients": recipe.Ingredients,
				"steps":       recipe.Steps,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				createArg := db.CreateRecipeParams{
					Author:      recipe.Author,
					Ingredients: recipe.Ingredients,
					Steps:       recipe.Steps,
				}
				store.EXPECT().
					CreateRecipe(gomock.Any(), gomock.Eq(createArg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidIngredients",
			body: map[string]interface{}{
				"author":      recipe.Author,
				"ingredients": []string{},
				"steps":       recipe.Steps,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				createArg := db.CreateRecipeParams{
					Author:      recipe.Author,
					Ingredients: recipe.Ingredients,
					Steps:       recipe.Steps,
				}
				store.EXPECT().
					CreateRecipe(gomock.Any(), gomock.Eq(createArg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidSteps",
			body: map[string]interface{}{
				"author":      recipe.Author,
				"ingredients": recipe.Ingredients,
				"steps":       []string{},
			},
			buildStubs: func(store *mockedstore.MockStore) {
				createArg := db.CreateRecipeParams{
					Author:      recipe.Author,
					Ingredients: recipe.Ingredients,
					Steps:       recipe.Steps,
				}
				store.EXPECT().
					CreateRecipe(gomock.Any(), gomock.Eq(createArg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NonexistentAuthor",
			body: map[string]interface{}{
				"author":      recipe.Author,
				"ingredients": recipe.Ingredients,
				"steps":       recipe.Steps,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				createArg := db.CreateRecipeParams{
					Author:      recipe.Author,
					Ingredients: recipe.Ingredients,
					Steps:       recipe.Steps,
				}
				store.EXPECT().
					CreateRecipe(gomock.Any(), gomock.Eq(createArg)).
					Times(1).
					Return(db.Recipe{}, &pq.Error{Code: "23503"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: map[string]interface{}{
				"author":      recipe.Author,
				"ingredients": recipe.Ingredients,
				"steps":       recipe.Steps,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				createArg := db.CreateRecipeParams{
					Author:      recipe.Author,
					Ingredients: recipe.Ingredients,
					Steps:       recipe.Steps,
				}
				store.EXPECT().
					CreateRecipe(gomock.Any(), gomock.Eq(createArg)).
					Times(1).
					Return(db.Recipe{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockedstore.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := book_recipe_factory.New(store)
			recorder := httptest.NewRecorder()

			url := "/recipes"

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestRecipe(t *testing.T) {
	recipe := randomRecipe(t)

	testCases := []struct {
		name          string
		ID            int64
		buildStubs    func(store *mockedstore.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			ID:   recipe.ID,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(1).
					Return(recipe, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchRecipe(t, recorder.Body, recipe)
			},
		},
		{
			name: "InvalidID",
			ID:   0,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NotFound",
			ID:   recipe.ID,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(1).
					Return(db.Recipe{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			ID:   recipe.ID,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(1).
					Return(db.Recipe{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockedstore.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := book_recipe_factory.New(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/recipes/%v", tc.ID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestUpdate(t *testing.T) {
	recipe := randomRecipe(t)
	updatedSteps := random.StringSlice(8)
	updatedIngredients := random.StringSlice(6)
	updatedTime := time.Now().UTC()

	updatedRecipe := db.Recipe{
		ID:          recipe.ID,
		Author:      recipe.Author,
		Ingredients: updatedIngredients,
		Steps:       updatedSteps,
		CreatedAt:   recipe.CreatedAt,
		UpdatedAt:   updatedTime,
	}

	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockedstore.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"id":          recipe.ID,
				"ingredients": updatedIngredients,
				"steps":       updatedSteps,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				arg := db.UpdateRecipeParams{
					ID:          recipe.ID,
					Steps:       updatedSteps,
					Ingredients: updatedIngredients,
					UpdatedAt:   updatedTime,
				}
				store.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(1).
					Return(recipe, nil)
				store.EXPECT().
					UpdateRecipe(gomock.Any(), EqUpdateRecipesParams(arg)).
					Times(1).
					Return(updatedRecipe, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			body: map[string]interface{}{
				"id":          0,
				"ingredients": updatedIngredients,
				"steps":       updatedSteps,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NotFound",
			body: map[string]interface{}{
				"id":          recipe.ID,
				"ingredients": updatedIngredients,
				"steps":       updatedSteps,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(1).
					Return(db.Recipe{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "GetStepInternalError",
			body: map[string]interface{}{
				"id":          recipe.ID,
				"ingredients": updatedIngredients,
				"steps":       updatedSteps,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(1).
					Return(db.Recipe{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "UpdateStepInternalError",
			body: map[string]interface{}{
				"id":          recipe.ID,
				"ingredients": updatedIngredients,
				"steps":       updatedSteps,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				arg := db.UpdateRecipeParams{
					ID:          recipe.ID,
					Steps:       updatedSteps,
					Ingredients: updatedIngredients,
					UpdatedAt:   updatedTime,
				}
				store.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(1).
					Return(recipe, nil)
				store.EXPECT().
					UpdateRecipe(gomock.Any(), EqUpdateRecipesParams(arg)).
					Times(1).
					Return(db.Recipe{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockedstore.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := book_recipe_factory.New(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/recipes"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDelete(t *testing.T) {
	recipe := randomRecipe(t)

	testCases := []struct {
		name          string
		ID            int64
		buildStubs    func(store *mockedstore.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			ID:   recipe.ID,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					DeleteRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			ID:   -2,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					DeleteRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			ID:   recipe.ID,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					DeleteRecipe(gomock.Any(), gomock.Eq(recipe.ID)).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockedstore.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := book_recipe_factory.New(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/recipes/%v", tc.ID)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestList(t *testing.T) {
	n := 10
	var recipe db.Recipe
	recipes := make([]db.Recipe, 0, n)
	for i := 0; i < n; i++ {
		recipe = randomRecipe(t)
		recipes = append(recipes, recipe)
	}

	pageID := 1
	pageSize := 10

	testCases := []struct {
		name           string
		paginationData struct {
			pageID   int32
			pageSize int32
		}
		buildStubs    func(store *mockedstore.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			paginationData: struct {
				pageID   int32
				pageSize int32
			}{pageID: int32(pageID), pageSize: int32(pageSize)},
			buildStubs: func(store *mockedstore.MockStore) {
				arg := db.ListRecipesParams{
					Limit:  int32(pageSize),
					Offset: int32(pageSize * (pageID - 1)),
				}
				store.EXPECT().
					ListRecipes(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(recipes, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchList(t, recorder.Body, recipes)
			},
		},
		{
			name: "InvalidPageID",
			paginationData: struct {
				pageID   int32
				pageSize int32
			}{pageID: int32(0), pageSize: int32(pageSize)},
			buildStubs: func(store *mockedstore.MockStore) {
				arg := db.ListRecipesParams{
					Limit:  int32(pageSize),
					Offset: int32(pageSize * (pageID - 1)),
				}
				store.EXPECT().
					ListRecipes(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPageSize",
			paginationData: struct {
				pageID   int32
				pageSize int32
			}{pageID: int32(pageID), pageSize: int32(15)},
			buildStubs: func(store *mockedstore.MockStore) {
				arg := db.ListRecipesParams{
					Limit:  int32(pageSize),
					Offset: int32(pageSize * (pageID - 1)),
				}
				store.EXPECT().
					ListRecipes(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NotFound",
			paginationData: struct {
				pageID   int32
				pageSize int32
			}{pageID: int32(pageID), pageSize: int32(pageSize)},
			buildStubs: func(store *mockedstore.MockStore) {
				arg := db.ListRecipesParams{
					Limit:  int32(pageSize),
					Offset: int32(pageSize * (pageID - 1)),
				}
				store.EXPECT().
					ListRecipes(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.Recipe{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			paginationData: struct {
				pageID   int32
				pageSize int32
			}{pageID: int32(pageID), pageSize: int32(pageSize)},
			buildStubs: func(store *mockedstore.MockStore) {
				arg := db.ListRecipesParams{
					Limit:  int32(pageSize),
					Offset: int32(pageSize * (pageID - 1)),
				}
				store.EXPECT().
					ListRecipes(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.Recipe{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockedstore.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := book_recipe_factory.New(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/recipes?page_id=%v&page_size=%v", tc.paginationData.pageID, tc.paginationData.pageSize)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAuthor(t *testing.T) db.Author {
	randomPassword := random.String(8)
	hashedPassword, err := password.HashPassword(randomPassword)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	now := time.Now().UTC()
	author := db.Author{
		Username:       random.String(10),
		Email:          random.Email(),
		HashedPassword: hashedPassword,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return author
}

func randomRecipe(t *testing.T) db.Recipe {
	author := randomAuthor(t)
	now := time.Now().UTC()
	recipe := db.Recipe{
		ID:          1,
		Author:      author.Username,
		Steps:       random.StringSlice(5),
		Ingredients: random.StringSlice(5),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return recipe
}

func requireBodyMatchCreate(t *testing.T, body *bytes.Buffer, recipe db.Recipe) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var expectedRecipeModel, gotRecipe recipeModel.CreateResponse
	jsonExpectedRecipe, err := json.Marshal(recipe)
	require.NoError(t, err)
	err = json.Unmarshal(jsonExpectedRecipe, &expectedRecipeModel)
	require.NoError(t, err)
	require.NotEmpty(t, expectedRecipeModel)

	err = json.Unmarshal(data, &gotRecipe)
	require.NoError(t, err)
	require.NotEmpty(t, gotRecipe)

	require.Equal(t, expectedRecipeModel.Author, gotRecipe.Author)
	require.Equal(t, expectedRecipeModel.Ingredients, gotRecipe.Ingredients)
	require.Equal(t, expectedRecipeModel.Steps, gotRecipe.Steps)
	require.Equal(t, expectedRecipeModel.CreatedAt, gotRecipe.CreatedAt)
	require.Empty(t, gotRecipe.ID)
	require.Empty(t, gotRecipe.UpdatedAt)
}

func requireBodyMatchRecipe(t *testing.T, body *bytes.Buffer, recipe db.Recipe) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var expectedRecipeModel, gotRecipe recipeModel.GetResponse
	jsonExpectedRecipe, err := json.Marshal(recipe)
	require.NoError(t, err)
	err = json.Unmarshal(jsonExpectedRecipe, &expectedRecipeModel)
	require.NoError(t, err)

	err = json.Unmarshal(data, &gotRecipe)
	require.NoError(t, err)

	require.Equal(t, expectedRecipeModel.Author, gotRecipe.Author)
	require.Equal(t, expectedRecipeModel.Ingredients, gotRecipe.Ingredients)
	require.Equal(t, expectedRecipeModel.Steps, gotRecipe.Steps)
	require.Equal(t, expectedRecipeModel.CreatedAt, gotRecipe.CreatedAt)
	require.Equal(t, expectedRecipeModel.UpdatedAt, gotRecipe.UpdatedAt)
	require.Empty(t, gotRecipe.ID)
}

func requireBodyMatchList(t *testing.T, body *bytes.Buffer, recipes []db.Recipe) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var expectedListRecipesModel, gotRecipes []recipeModel.ListResponse
	jsonListExpectedModel, err := json.Marshal(recipes)
	require.NoError(t, err)
	err = json.Unmarshal(jsonListExpectedModel, &expectedListRecipesModel)
	require.NoError(t, err)

	err = json.Unmarshal(data, &gotRecipes)
	require.NoError(t, err)

	require.Equal(t, len(expectedListRecipesModel), len(gotRecipes))

	for i, recipe := range gotRecipes {
		require.NotEmpty(t, recipe)
		require.Empty(t, recipe.ID)

		require.Equal(t, expectedListRecipesModel[i].Author, recipe.Author)
		require.Equal(t, expectedListRecipesModel[i].Steps, recipe.Steps)
		require.Equal(t, expectedListRecipesModel[i].Ingredients, recipe.Ingredients)
		require.Equal(t, expectedListRecipesModel[i].CreatedAt, recipe.CreatedAt)
		require.Equal(t, expectedListRecipesModel[i].UpdatedAt, recipe.UpdatedAt)
	}
}
