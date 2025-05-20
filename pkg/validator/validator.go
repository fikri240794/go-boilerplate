package validator

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/fikri240794/gocerr"
	"github.com/go-playground/locales"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	validate   *validator.Validate
	translator ut.Translator
)

func fieldFromJSONTag(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

	if name == "-" {
		return ""
	}

	return name
}

func translateErrorToEnglish() {
	var (
		localeEnglish       locales.Translator
		universalTranslator *ut.UniversalTranslator
	)

	localeEnglish = en.New()
	universalTranslator = ut.New(localeEnglish)
	translator, _ = universalTranslator.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, translator)
}

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(fieldFromJSONTag)
	translateErrorToEnglish()
}

func ValidateStruct(s interface{}) error {
	var err error = validate.Struct(s)
	if err != nil {
		var errFields []gocerr.ErrorField = []gocerr.ErrorField{}

		for _, errField := range err.(validator.ValidationErrors) {
			errFields = append(
				errFields,
				gocerr.NewErrorField(
					errField.Field(),
					errField.Translate(translator),
				),
			)
		}

		return gocerr.New(
			http.StatusBadRequest,
			http.StatusText(http.StatusBadRequest),
			errFields...,
		)
	}

	return nil
}
