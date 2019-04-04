package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/locales/ja_JP"
	"github.com/go-playground/universal-translator"
	"github.com/labstack/echo"
	_ "github.com/lib/pq"
	"gopkg.in/go-playground/validator.v9"
	ja "gopkg.in/go-playground/validator.v9/translations/ja"
	"gopkg.in/gorp.v2"
)

var dbDriver = "postgres"

// Validator is implementation of validation of rquest values.
type Validator struct {
	trans     ut.Translator
	validator *validator.Validate
}

// Validate do validation for request value.
func (v *Validator) Validate(i interface{}) error {
	err := v.validator.Struct(i)
	if err == nil {
		return nil
	}
	errs := err.(validator.ValidationErrors)
	msg := ""
	for _, v := range errs.Translate(v.trans) {
		if msg != "" {
			msg += ", "
		}
		msg += v
	}
	return errors.New(msg)
}

// Error indicate response erorr
type Error struct {
	Error string `json:"error"`
}

// Comment is a struct to hold unit of request and response.
type Comment struct {
	Id      int64     `json:"id" db:"id,primarykey,autoincrement"`
	Name    string    `json:"name" form:"name" db:"name,notnull,default:'名無し',size:200"`
	Text    string    `json:"text" form:"text" validate:"required,max=20" db:"text,notnull,size:399"`
	Created time.Time `json:"created" db:"created,notnull"`
	Updated time.Time `json:"updated" db:"updated,notnull"`
}

// PreInsert update fields Created and Updated.
func (c *Comment) PreInsert(s gorp.SqlExecutor) error {
	c.Created = time.Now()
	c.Updated = c.Created
	return nil
}

// PreInsert update field Updated.
func (c *Comment) PreUpdate(s gorp.SqlExecutor) error {
	c.Updated = time.Now()
	return nil
}

func setupDB() (*gorp.DbMap, error) {
	db, err := sql.Open(dbDriver, os.Getenv("DSN"))
	if err != nil {
		return nil, err
	}

	var diarect gorp.Dialect = gorp.PostgresDialect{}

	// for testing
	if dbDriver == "sqlite3" {
		diarect = gorp.SqliteDialect{}
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: diarect}
	dbmap.AddTableWithName(Comment{}, "comments").SetKeys(true, "id")
	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		return nil, err
	}
	return dbmap, nil
}

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Debug = true
	e.Logger.SetOutput(os.Stderr)

	// setup japanese translation
	japanese := ja_JP.New()
	uni := ut.New(japanese, japanese)
	trans, _ := uni.GetTranslator("ja")
	validate := validator.New()
	err := ja.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		log.Fatal(err)
	}

	// register japanese translation for input field
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		switch name {
		case "name":
			return "お名前"
		case "text":
			return "コメント"
		case "-":
			return ""
		}
		return name
	})

	e.Validator = &Validator{validator: validate, trans: trans}
	return e
}

// Controller is a controller for this application.
type Controller struct {
	dbmap *gorp.DbMap
}

// InsertComment is GET handler to return record.
func (controller *Controller) GetComment(c echo.Context) error {
	var comment Comment
	// fetch record specified by parameter id
	err := controller.dbmap.SelectOne(&comment,
		"SELECT * FROM comments WHERE id = $1", c.Param("id"))
	if err != nil {
		if err != sql.ErrNoRows {
			c.Logger().Error("SelectOne: ", err)
			return c.String(http.StatusBadRequest, "SelectOne: "+err.Error())
		}
		return c.String(http.StatusNotFound, "Not Found")
	}
	return c.JSON(http.StatusOK, comment)
}

// InsertComment is GET handler to return records.
func (controller *Controller) ListComments(c echo.Context) error {
	var comments []Comment
	// fetch last 10 records
	_, err := controller.dbmap.Select(&comments,
		"SELECT * FROM comments ORDER BY created desc LIMIT 10")
	if err != nil {
		c.Logger().Error("Select: ", err)
		return c.String(http.StatusBadRequest, "Select: "+err.Error())
	}
	return c.JSON(http.StatusOK, comments)
}

// InsertComment is POST handler to insert record.
func (controller *Controller) InsertComment(c echo.Context) error {
	var comment Comment
	// bind request to comment struct
	if err := c.Bind(&comment); err != nil {
		c.Logger().Error("Bind: ", err)
		return c.String(http.StatusBadRequest, "Bind: "+err.Error())
	}
	// validate request
	if err := c.Validate(&comment); err != nil {
		c.Logger().Error("Validate: ", err)
		return c.JSON(http.StatusBadRequest, &Error{Error: err.Error()})
	}
	// insert record
	if err := controller.dbmap.Insert(&comment); err != nil {
		c.Logger().Error("Insert: ", err)
		return c.String(http.StatusBadRequest, "Insert: "+err.Error())
	}
	c.Logger().Infof("inserted comment: %v", comment.Id)
	return c.NoContent(http.StatusCreated)
}

func main() {
	dbmap, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}
	controller := &Controller{dbmap: dbmap}

	e := setupEcho()

	e.GET("/api/comments/:id", controller.GetComment)
	e.GET("/api/comments", controller.ListComments)
	e.POST("/api/comments", controller.InsertComment)
	e.Static("/", "static/")
	e.Logger.Fatal(e.Start(":8989"))
}
