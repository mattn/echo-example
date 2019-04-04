package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gorp.v2"
)

func init() {
	dbDriver = "sqlite3"
}

func testSetupDB() (*gorp.DbMap, error) {
	os.Remove("test.sqlite")

	old := os.Getenv("DSN")
	defer os.Setenv("DSN", old)
	os.Setenv("DSN", "test.sqlite")
	return setupDB()
}

func TestInsertCommentWithoutComment(t *testing.T) {
	dbmap, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}

	controller := &Controller{dbmap: dbmap}

	req := httptest.NewRequest(http.MethodPost, "/api/comments", strings.NewReader(`
	{
		"name": "job"
	}
	`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = controller.InsertComment(c)
	if err != nil {
		t.Fatal(err)
	}
	var errMsg Error
	err = json.NewDecoder(rec.Body).Decode(&errMsg)
	if err != nil {
		t.Fatal(err)
	}
	got := errMsg.Error
	want := "コメントは必須フィールドです"
	if got != want {
		log.Fatalf("want %v but got %v", want, got)
	}
}

func TestInsertCommentWithComment(t *testing.T) {
	dbmap, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}

	controller := &Controller{dbmap: dbmap}

	req := httptest.NewRequest(http.MethodPost, "/api/comments", strings.NewReader(`
	{
		"name": "job",
		"text": "hello"
	}
	`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = controller.InsertComment(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != 201 {
		t.Fatal("should be succeeded")
	}
	if rec.Body.Len() > 0 {
		log.Fatal("response body should be empty")
	}
}

func TestGetComment(t *testing.T) {
	dbmap, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}

	controller := &Controller{dbmap: dbmap}

	req := httptest.NewRequest(http.MethodGet, "/api/comments/1", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/comments/:id")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(1))

	err = controller.GetComment(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != 404 {
		t.Fatal("should be 404")
	}

	comment := Comment{
		Text: "hello",
	}
	if dbmap.Insert(&comment); err != nil {
		t.Fatal(err)
	}

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/api/comments/:id")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(1))

	err = controller.GetComment(c)
	if err != nil {
		t.Fatal(err)
	}

	err = json.NewDecoder(rec.Body).Decode(&comment)
	if err != nil {
		t.Fatal(err)
	}
	want := "hello"
	got := comment.Text
	if got != want {
		log.Fatalf("want %v but got %v", want, got)
	}
}

func TestListComment(t *testing.T) {
	dbmap, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}

	controller := &Controller{dbmap: dbmap}

	req := httptest.NewRequest(http.MethodGet, "/api/comments", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/comments/:id")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(1))

	err = controller.ListComments(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 {
		t.Fatal("should be 200")
	}
	var comments []Comment
	err = json.NewDecoder(rec.Body).Decode(&comments)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) > 0 {
		t.Fatal("should be empty")
	}

	comment := Comment{
		Text: "hello",
	}
	if dbmap.Insert(&comment); err != nil {
		t.Fatal(err)
	}

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = controller.ListComments(c)
	if err != nil {
		t.Fatal(err)
	}

	err = json.NewDecoder(rec.Body).Decode(&comments)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) == 0 {
		t.Fatal("should not be empty")
	}
}
