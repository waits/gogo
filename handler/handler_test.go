package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/waits/gogo/model"
)

var tmpls = LoadTemplates("../template/")
var env = &Env{tmpls, "../template/"}

// Flushes the database and sets up test data
func init() {
	pool := model.InitPool(1)
	conn := pool.Get()
	defer conn.Close()

	conn.Send("FLUSHDB")
	games := []model.Game{
		{Key: "1", Size: 9, Turn: 1, Ko: -1, Handicap: 0, Black: "Aaron", White: "Job"},
		{Key: "2", Size: 9, Turn: 1, Ko: -1, Handicap: 0, Black: "Frank"},
	}
	for _, g := range games {
		args := redis.Args{}.Add("game:" + g.Key).AddFlat(g)
		conn.Send("HMSET", args...)
	}
	conn.Flush()
}

func recordRequest(t *testing.T, fn func(*Env, http.ResponseWriter, *http.Request) (int, error),
	method string, path string, data string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	body := strings.NewReader(data)
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) > 0 {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	handler := http.Handler(Handler{env, fn})
	handler.ServeHTTP(rec, req)

	return rec
}

func testStatusCode(t *testing.T, rec *httptest.ResponseRecorder, expected int) {
	if status := rec.Code; status != expected {
		t.Errorf("handler returned wrong status code: got %v want %v", status, expected)
	}
}

func testBody(t *testing.T, rec *httptest.ResponseRecorder, expected string) {
	if body := rec.Body.String(); !strings.Contains(body, expected) {
		t.Errorf("handler returned unexpected body: did not contain %v", expected)
	}
}

func testRedirect(t *testing.T, rec *httptest.ResponseRecorder, expected string) {
	if location := rec.Header().Get("Location"); !strings.HasPrefix(location, expected) {
		t.Errorf("handler redirected to wrong url: got %v want %v", location, expected)
	}
}

func TestStaticHandler(t *testing.T) {
	rec := recordRequest(t, StaticHandler, "GET", "/", "")
	testStatusCode(t, rec, http.StatusOK)
	testBody(t, rec, "No games are in progress.")

	rec = recordRequest(t, StaticHandler, "GET", "/help", "")
	testStatusCode(t, rec, http.StatusOK)
	testBody(t, rec, "Rules")

	rec = recordRequest(t, StaticHandler, "GET", "/new", "")
	testStatusCode(t, rec, http.StatusOK)
	testBody(t, rec, "New Game")
}

func TestCreateGame(t *testing.T) {
	rec := recordRequest(t, GameHandler, "POST", "/game", "name=Marco&color=white&size=19&handicap=3")
	testStatusCode(t, rec, http.StatusSeeOther)
	testRedirect(t, rec, "/game/")
}

func TestJoinGame(t *testing.T) {
	rec := recordRequest(t, GameHandler, "PUT", "/game/2", "name=Guy")
	testStatusCode(t, rec, http.StatusSeeOther)
	testRedirect(t, rec, "/game/2")
}

func TestShowGame(t *testing.T) {
	c := &http.Cookie{Name: "1", Value: "black"}
	rec := recordRequest(t, GameHandler, "GET", "/game/1", "", c)
	testStatusCode(t, rec, http.StatusOK)
	testBody(t, rec, "Aaron vs. Job")
}

func TestUpdateGame(t *testing.T) {
	c := &http.Cookie{Name: "1", Value: "black"}
	rec := recordRequest(t, GameHandler, "PUT", "/game/1", "x=1&y=2", c)
	testStatusCode(t, rec, http.StatusOK)
}
