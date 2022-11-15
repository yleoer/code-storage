package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/locales/zh_Hans"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	val "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

type CountRequest struct {
	Name  string `json:"name" form:"name" binding:"max=100"`
	State uint8  `json:"state" form:"state,default=1" binding:"oneof=0 1"`
}

var u *ut.UniversalTranslator

func Translations() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := c.GetHeader("locale")
		trans, _ := u.GetTranslator(locale)
		c.Set("trans", trans)

		c.Next()
	}
}

type ValidError struct {
	Key     string
	Message string
}

type ValidErrors []*ValidError

func (v *ValidError) Error() string {
	return v.Message
}

func (v ValidErrors) Errors() []string {
	var errs []string
	for _, err := range v {
		errs = append(errs, err.Error())
	}

	return errs
}

func (v ValidErrors) Error() string {
	return strings.Join(v.Errors(), ",")
}

func BindAndValid(c *gin.Context, v interface{}) (bool, ValidErrors) {
	var errs ValidErrors
	err := c.ShouldBind(v)
	if err != nil {
		v := c.Value("trans")
		trans, _ := v.(ut.Translator)
		log.Println("===||trans", trans.Locale())
		verrs, ok := err.(val.ValidationErrors)
		if !ok {
			log.Printf("no ok: %#v\n", err)
			return false, errs
		}

		for key, value := range verrs.Translate(trans) {
			errs = append(errs, &ValidError{
				Key:     key,
				Message: value,
			})
		}

		return false, errs
	}

	return true, nil
}

func init() {
	u = ut.New(en.New(), zh.New(), zh_Hans.New())
	zhTrans, _ := u.GetTranslator("zh")
	enTrans, _ := u.GetTranslator("en")
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		_ = zh_translations.RegisterDefaultTranslations(v, zhTrans)
		_ = en_translations.RegisterDefaultTranslations(v, enTrans)
	}
}

func main() {
	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	r.Use(Translations())
	r.POST("/", func(c *gin.Context) {
		var param CountRequest
		valid, errs := BindAndValid(c, &param)
		if !valid {
			fmt.Printf("errs: %#v\n", errs)
			c.JSON(400, errs)
			return
		}
		c.JSON(200, nil)
		return
	})

	log.Fatal(r.Run())
}
