package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/labstack/gommon/log"
	govaluate "gopkg.in/Knetic/govaluate.v2"
)

// Validator to store all validator data
type Validator struct {
	Channel            string   `json:"channel,omitempty"`
	Product            string   `json:"product,omitempty"`
	TransactionType    string   `json:"transactionType,omitempty"`
	Unit               string   `json:"unit,omitempty"`
	Multiplier         *float64 `json:"multiplier,omitempty"`
	Value              *int64   `json:"value,omitempty"`
	Formula            string   `json:"formula,omitempty"`
	MinimalTransaction string   `json:"minimalTransaction,omitempty"`
	Source             string   `json:"source,omitempty"` // device name that user used
}

// PayloadValidator to store a payload to validate a request
type PayloadValidator struct {
	PromoCode         string     `json:"promoCode,omitempty"`
	VoucherID         string     `json:"voucherId,omitempty"`
	CampaignID        string     `json:"campaignId,omitempty"`
	UserID            string     `json:"userId,omitempty"`
	TransactionAmount float64    `json:"transactionAmount,omitempty"`
	Validators        *Validator `json:"validators,omitempty"`
}

var skippedValidator = []string{"multiplier", "value", "formula"}
var compareEqual = []string{"channel", "product", "transactionType", "unit", "source"}

// Validate to validate client input with admin input
func (v *Validator) Validate(payloadValidator *PayloadValidator) error {
	var reqValidator map[string]interface{}

	if v == nil {
		log.Error(ErrValidatorUnavailable)

		return ErrValidatorUnavailable
	}

	vReflector := reflect.ValueOf(v).Elem()
	tempJSON, _ := json.Marshal(payloadValidator.Validators)
	json.Unmarshal(tempJSON, &reqValidator)

	for i := 0; i < vReflector.NumField(); i++ {
		interfaceValue := vReflector.Field(i).Interface()

		fieldName := strcase.ToLowerCamel(vReflector.Type().Field(i).Name)
		fieldValue := getInterfaceValue(interfaceValue)

		if contains(skippedValidator, fieldName) {
			continue
		}

		if fieldValue == "" {
			continue
		}

		switch {
		case contains(compareEqual, fieldName):
			reqValidatorVal := fmt.Sprintf("%v", reqValidator[fieldName])

			if !strings.Contains(fieldValue, reqValidatorVal) {
				customErr := fmt.Errorf("%s on this transaction is not valid to use this", fieldName)

				return customErr
			}
		case fieldName == "minimalTransaction":
			minTrx, _ := strconv.ParseFloat(fieldValue, 64)

			if minTrx > payloadValidator.TransactionAmount {
				return ErrValidationTrxAmt
			}
		}
	}

	return nil
}

// GetFormulaResult to proccess the formula then get the result
func (v *Validator) GetFormulaResult(payloadValidator *PayloadValidator) (float64, error) {
	expression, err := govaluate.NewEvaluableExpression(v.Formula)

	if err != nil {
		log.Error(err)

		return 0, err
	}

	// Make a Regex to say we only want letters and numbers
	regex, err := regexp.Compile("[^a-zA-Z0-9]+")

	if err != nil {
		log.Fatal(err)
	}

	formulaVarStr := regex.ReplaceAllString(v.Formula, " ")
	formulaVarStr = strings.TrimLeft(formulaVarStr, " ")
	formulaVarStr = strings.TrimRight(formulaVarStr, " ")
	formulaVar := strings.Split(formulaVarStr, " ")
	remove(formulaVar, "transactionAmount")

	// get formula parameters
	parameters := make(map[string]interface{}, 8)
	parameters["transactionAmount"] = payloadValidator.TransactionAmount

	for _, fVar := range formulaVar {
		parameters[fVar] = v.getField(fVar)
	}

	result, err := expression.Evaluate(parameters)

	if err != nil {
		log.Error(err)

		return 0, err
	}

	// Parse interface to float
	return getFloat(result)
}

func (v *Validator) getField(field string) interface{} {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(strings.Title(field)).Interface()

	if f == nil {
		return 0
	}

	switch f.(type) {
	case *float64:
		v := getInterfaceValue(f)
		result, _ := strconv.ParseFloat(v, 64)

		return result
	case *int64:
		v := getInterfaceValue(f)
		result, _ := strconv.ParseInt(v, 10, 64)

		return result
	}

	return getInterfaceValue(f)
}

func contains(strings []string, str string) bool {
	for _, n := range strings {
		if str == n {
			return true
		}
	}

	return false
}

func getInterfaceValue(intfc interface{}) string {
	switch v := intfc.(type) {
	case *float64:
		if v == nil {
			return ""
		}

		return fmt.Sprintf("%v", *v)
	case *int64:
		if v == nil {
			return ""
		}

		return fmt.Sprintf("%v", *v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getFloat(unk interface{}) (float64, error) {
	floatType := reflect.TypeOf(float64(0))
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)

	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}

	fv := v.Convert(floatType)
	return fv.Float(), nil
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
