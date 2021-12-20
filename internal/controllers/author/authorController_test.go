package authorController_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gmaschi/go-recipes-book/internal/factories/book-recipe-factory"
	mockedstore "github.com/gmaschi/go-recipes-book/internal/mocks/datastore/postgresql/recipes"
	authorModel "github.com/gmaschi/go-recipes-book/internal/models/author"
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

type eqCreateAuthorParamsMatcher struct {
	arg      db.CreateAuthorParams
	password string
}

func (e eqCreateAuthorParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateAuthorParams)
	if !ok {
		return false
	}

	err := password.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword

	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateAuthorParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateAuthorParams(arg db.CreateAuthorParams, password string) gomock.Matcher {
	return eqCreateAuthorParamsMatcher{arg, password}
}

type eqUpdateAuthorParamsMatcher struct {
	arg             db.UpdateAuthorParams
	updatedPassword string
}

func (e eqUpdateAuthorParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.UpdateAuthorParams)
	if !ok {
		return false
	}

	err := password.CheckPassword(e.updatedPassword, arg.HashedPassword)
	if err != nil {
		return false
	}
	e.arg.HashedPassword = arg.HashedPassword
	e.arg.UpdatedAt = arg.UpdatedAt
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqUpdateAuthorParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.updatedPassword)
}

func EqUpdateAuthorParams(arg db.UpdateAuthorParams, updatedPassword string) gomock.Matcher {
	return eqUpdateAuthorParamsMatcher{arg, updatedPassword}
}

func TestCreate(t *testing.T) {
	author, password := randomAuthor(t)

	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockedstore.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"username": author.Username,
				"password": password,
				"email":    author.Email,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				arg := db.CreateAuthorParams{
					Username: author.Username,
					Email:    author.Email,
				}
				store.EXPECT().
					CreateAuthor(gomock.Any(), EqCreateAuthorParams(arg, password)).
					Times(1).
					Return(author, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCreate(t, recorder.Body, author)
			},
		},
		{
			name: "InvalidEmail",
			body: map[string]interface{}{
				"username": author.Username,
				"password": password,
				"email":    "invalid-email",
			},
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					CreateAuthor(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: map[string]interface{}{
				"username": "inavlid-username#",
				"password": password,
				"email":    author.Email,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					CreateAuthor(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPassword",
			body: map[string]interface{}{
				"username": author.Username,
				"password": "abc",
				"email":    author.Email,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					CreateAuthor(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: map[string]interface{}{
				"username": author.Username,
				"password": password,
				"email":    author.Email,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					CreateAuthor(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Author{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "ExistingUsername",
			body: map[string]interface{}{
				"username": author.Username,
				"password": password,
				"email":    author.Email,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					CreateAuthor(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Author{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockedstore.NewMockStore(ctrl)
			// build stubs
			tc.buildStubs(store)

			// start test server and send request
			//server := newTestServer(t, store)
			server := book_recipe_factory.New(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/authors"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			// check response
			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestAuthor(t *testing.T) {
	author, _ := randomAuthor(t)

	testCases := []struct {
		name           string
		authorUsername string
		buildStubs     func(store *mockedstore.MockStore)
		checkResponse  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:           "OK",
			authorUsername: author.Username,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(author, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAuthor(t, recorder.Body, author)
			},
		},
		{
			name:           "BadRequest",
			authorUsername: "inavlid-username#",
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:           "InternalError",
			authorUsername: author.Username,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(db.Author{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:           "NotFound",
			authorUsername: author.Username,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(db.Author{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
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

			url := fmt.Sprintf("/authors/%s", tc.authorUsername)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestUpdate(t *testing.T) {
	author, _ := randomAuthor(t)
	updatedEmail := random.Email()
	updatedPassword := random.String(10)
	updatedHashedPassword, err := password.HashPassword(updatedPassword)
	require.NoError(t, err)
	updatedTime := time.Now().UTC()

	updatedAuthor := db.Author{
		Username:       author.Username,
		HashedPassword: updatedHashedPassword,
		Email:          updatedEmail,
		CreatedAt:      author.CreatedAt,
		UpdatedAt:      updatedTime,
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
				"username": author.Username,
				"email":    updatedEmail,
				"password": updatedPassword,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				updateArgs := db.UpdateAuthorParams{
					Username:  author.Username,
					Email:     updatedEmail,
					UpdatedAt: updatedTime,
				}
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(author, nil)
				store.EXPECT().
					UpdateAuthor(gomock.Any(), EqUpdateAuthorParams(updateArgs, updatedPassword)).
					Times(1).
					Return(updatedAuthor, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUpdateAuthor(t, recorder.Body, updatedAuthor)
			},
		},
		{
			name: "InvalidUsername",
			body: map[string]interface{}{
				"username": "invalid-username#",
				"email":    updatedEmail,
				"password": updatedPassword,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidUpdatedEmail",
			body: map[string]interface{}{
				"username": author.Username,
				"email":    "invalid-email",
				"password": updatedPassword,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				updateArgs := db.UpdateAuthorParams{
					Username:  author.Username,
					Email:     updatedEmail,
					UpdatedAt: updatedTime,
				}
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(author, nil)
				store.EXPECT().
					UpdateAuthor(gomock.Any(), EqUpdateAuthorParams(updateArgs, updatedPassword)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidUpdatedPassword",
			body: map[string]interface{}{
				"username": author.Username,
				"email":    updatedEmail,
				"password": "367",
			},
			buildStubs: func(store *mockedstore.MockStore) {
				updateArgs := db.UpdateAuthorParams{
					Username:  author.Username,
					Email:     updatedEmail,
					UpdatedAt: updatedTime,
				}
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(author, nil)
				store.EXPECT().
					UpdateAuthor(gomock.Any(), EqUpdateAuthorParams(updateArgs, updatedPassword)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NotFound",
			body: map[string]interface{}{
				"username": author.Username,
				"email":    updatedEmail,
				"password": updatedPassword,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				updateArgs := db.UpdateAuthorParams{
					Username:  author.Username,
					Email:     updatedEmail,
					UpdatedAt: updatedTime,
				}
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(db.Author{}, sql.ErrNoRows)
				store.EXPECT().
					UpdateAuthor(gomock.Any(), EqUpdateAuthorParams(updateArgs, updatedPassword)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "GetAuthorInternalError",
			body: map[string]interface{}{
				"username": author.Username,
				"email":    updatedEmail,
				"password": updatedPassword,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				updateArgs := db.UpdateAuthorParams{
					Username:  author.Username,
					Email:     updatedEmail,
					UpdatedAt: updatedTime,
				}
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(db.Author{}, sql.ErrConnDone)
				store.EXPECT().
					UpdateAuthor(gomock.Any(), EqUpdateAuthorParams(updateArgs, updatedPassword)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "UpdateAuthorInternalError",
			body: map[string]interface{}{
				"username": author.Username,
				"email":    updatedEmail,
				"password": updatedPassword,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				updateArgs := db.UpdateAuthorParams{
					Username:  author.Username,
					Email:     updatedEmail,
					UpdatedAt: updatedTime,
				}
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(author, nil)
				store.EXPECT().
					UpdateAuthor(gomock.Any(), EqUpdateAuthorParams(updateArgs, updatedPassword)).
					Times(1).
					Return(db.Author{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "UniqueEmailViolation",
			body: map[string]interface{}{
				"username": author.Username,
				"email":    updatedEmail,
				"password": updatedPassword,
			},
			buildStubs: func(store *mockedstore.MockStore) {
				updateArgs := db.UpdateAuthorParams{
					Username:  author.Username,
					Email:     updatedEmail,
					UpdatedAt: updatedTime,
				}
				store.EXPECT().
					GetAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(author, nil)
				store.EXPECT().
					UpdateAuthor(gomock.Any(), EqUpdateAuthorParams(updateArgs, updatedPassword)).
					Times(1).
					Return(db.Author{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
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

			url := "/authors"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDelete(t *testing.T) {
	author, _ := randomAuthor(t)

	testCases := []struct {
		name           string
		authorUsername string
		buildStubs     func(store *mockedstore.MockStore)
		checkResponse  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:           "OK",
			authorUsername: author.Username,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					DeleteAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:           "InvalidUsername",
			authorUsername: "inavlid-username#",
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					DeleteAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:           "ForeignKeyViolation",
			authorUsername: author.Username,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					DeleteAuthor(gomock.Any(), gomock.Eq(author.Username)).
					Times(1).
					Return(&pq.Error{Code: "23503"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:           "InternalServerError",
			authorUsername: author.Username,
			buildStubs: func(store *mockedstore.MockStore) {
				store.EXPECT().
					DeleteAuthor(gomock.Any(), gomock.Eq(author.Username)).
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

			url := fmt.Sprintf("/authors/%s", tc.authorUsername)
			fmt.Println(url)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestList(t *testing.T) {
	n := 10
	authorsSlice := make([]db.Author, 0, 10)
	var author db.Author
	for i := 0; i < n; i++ {
		author, _ = randomAuthor(t)
		authorsSlice = append(authorsSlice, author)
	}

	pageID := 1
	pageSize := 5

	testCases := []struct {
		name           string
		paginationData struct {
			pageID   int
			pageSize int
		}
		buildStubs    func(store *mockedstore.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			paginationData: struct {
				pageID   int
				pageSize int
			}{pageID: pageID, pageSize: pageSize},
			buildStubs: func(store *mockedstore.MockStore) {
				listArgs := db.ListAuthorsParams{
					Limit:  int32(pageSize),
					Offset: int32(pageSize * (pageID - 1)),
				}
				store.EXPECT().
					ListAuthors(gomock.Any(), gomock.Eq(listArgs)).
					Times(1).
					Return(authorsSlice, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchListAuthors(t, recorder.Body, authorsSlice)
			},
		},
		{
			name: "InvalidPageID",
			paginationData: struct {
				pageID   int
				pageSize int
			}{pageID: 0, pageSize: pageSize},
			buildStubs: func(store *mockedstore.MockStore) {
				listArgs := db.ListAuthorsParams{
					Limit:  int32(pageSize),
					Offset: int32(pageSize * (pageID - 1)),
				}
				store.EXPECT().
					ListAuthors(gomock.Any(), gomock.Eq(listArgs)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPageSize",
			paginationData: struct {
				pageID   int
				pageSize int
			}{pageID: pageID, pageSize: 20},
			buildStubs: func(store *mockedstore.MockStore) {
				listArgs := db.ListAuthorsParams{
					Limit:  int32(pageSize),
					Offset: int32(pageSize * (pageID - 1)),
				}
				store.EXPECT().
					ListAuthors(gomock.Any(), gomock.Eq(listArgs)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NotFound",
			paginationData: struct {
				pageID   int
				pageSize int
			}{pageID: pageID, pageSize: pageSize},
			buildStubs: func(store *mockedstore.MockStore) {
				listArgs := db.ListAuthorsParams{
					Limit:  int32(pageSize),
					Offset: int32(pageSize * (pageID - 1)),
				}
				store.EXPECT().
					ListAuthors(gomock.Any(), gomock.Eq(listArgs)).
					Times(1).
					Return([]db.Author{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			paginationData: struct {
				pageID   int
				pageSize int
			}{pageID: pageID, pageSize: pageSize},
			buildStubs: func(store *mockedstore.MockStore) {
				listArgs := db.ListAuthorsParams{
					Limit:  int32(pageSize),
					Offset: int32(pageSize * (pageID - 1)),
				}
				store.EXPECT().
					ListAuthors(gomock.Any(), gomock.Eq(listArgs)).
					Times(1).
					Return([]db.Author{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/authors?page_id=%v&page_size=%v", tc.paginationData.pageID, tc.paginationData.pageSize)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAuthor(t *testing.T) (db.Author, string) {
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

	return author, randomPassword
}

func requireBodyMatchCreate(t *testing.T, body *bytes.Buffer, author db.Author) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var expectedAuthorModel, gotAuthor authorModel.CreateResponse
	jsonModelAuthor, err := json.Marshal(&author)
	require.NoError(t, err)
	err = json.Unmarshal(jsonModelAuthor, &expectedAuthorModel)
	require.NoError(t, err)

	err = json.Unmarshal(data, &gotAuthor)
	require.NoError(t, err)
	require.Equal(t, expectedAuthorModel, gotAuthor)

	require.Equal(t, expectedAuthorModel.Username, gotAuthor.Username)
	require.Equal(t, expectedAuthorModel.CreatedAt, gotAuthor.CreatedAt)
	require.Empty(t, gotAuthor.Email)
	require.Empty(t, gotAuthor.UpdatedAt)
	require.Empty(t, gotAuthor.HashedPassword)
}

func requireBodyMatchAuthor(t *testing.T, body *bytes.Buffer, author db.Author) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var expectedAuthorModel, gotAuthor authorModel.GetResponse
	jsonAuthor, err := json.Marshal(author)
	require.NoError(t, err)
	err = json.Unmarshal(jsonAuthor, &expectedAuthorModel)
	require.NoError(t, err)
	require.NotEmpty(t, expectedAuthorModel)

	err = json.Unmarshal(data, &gotAuthor)
	require.NoError(t, err)
	require.NotEmpty(t, gotAuthor)

	require.Equal(t, expectedAuthorModel.Username, gotAuthor.Username)
	require.Equal(t, expectedAuthorModel.Email, gotAuthor.Email)
	require.Equal(t, expectedAuthorModel.CreatedAt, gotAuthor.CreatedAt)
	require.Equal(t, expectedAuthorModel.UpdatedAt, gotAuthor.UpdatedAt)
}

func requireBodyMatchUpdateAuthor(t *testing.T, body *bytes.Buffer, author db.Author) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var expectedAuthorModel, gotAuthor authorModel.UpdateResponse

	jsonAuthor, err := json.Marshal(author)
	require.NoError(t, err)
	err = json.Unmarshal(jsonAuthor, &expectedAuthorModel)
	require.NoError(t, err)
	require.NotEmpty(t, expectedAuthorModel)

	err = json.Unmarshal(data, &gotAuthor)
	require.NoError(t, err)
	require.NotEmpty(t, gotAuthor)

	require.Equal(t, expectedAuthorModel.Username, gotAuthor.Username)
	require.Equal(t, expectedAuthorModel.Email, gotAuthor.Email)
	require.Equal(t, expectedAuthorModel.CreatedAt, gotAuthor.CreatedAt)
	require.Equal(t, expectedAuthorModel.UpdatedAt, gotAuthor.UpdatedAt)
}

func requireBodyMatchListAuthors(t *testing.T, body *bytes.Buffer, authors []db.Author) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var expectedAuthorsModel, gotAuthors []authorModel.ListResponse

	jsonAuthors, err := json.Marshal(authors)
	require.NoError(t, err)
	err = json.Unmarshal(jsonAuthors, &expectedAuthorsModel)
	require.NoError(t, err)

	err = json.Unmarshal(data, &gotAuthors)
	require.NoError(t, err)

	require.Equal(t, len(expectedAuthorsModel), len(gotAuthors))

	for i, author := range gotAuthors {
		require.NotEmpty(t, author)
		require.Empty(t, author.HashedPassword)
		require.Equal(t, expectedAuthorsModel[i].Username, author.Username)
		require.Equal(t, expectedAuthorsModel[i].Email, author.Email)
		require.Equal(t, expectedAuthorsModel[i].CreatedAt, author.CreatedAt)
		require.Equal(t, expectedAuthorsModel[i].UpdatedAt, author.UpdatedAt)
	}
}
