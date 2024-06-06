package commafeed_util

import (
	"fmt"
	"github.com/dayzerosec/zerodayfans/pkg/commafeed"
	"github.com/dayzerosec/zerodayfans/pkg/config"
	"log"
	"reflect"
	"strings"
)

func doFilter(entry commafeed.Entry, filter config.FeedFilter) bool {
	var result bool
	var err error
	switch filter.Type {
	case config.FilterTypeMatchOne:
		result, err = doFilterMatchOne(entry, filter)
	case config.FilterTypeMatchAll:
		result, err = doFilterMatchAll(entry, filter)
	default:
		log.Printf("Unknown filter type: %s", filter.Type)
		return false
	}

	if err != nil {
		log.Printf("Error processing filter: %v", err)
		return false
	}

	if filter.Negate {
		return !result
	}
	return result
}

func doFilterMatchOne(entry commafeed.Entry, filter config.FeedFilter) (bool, error) {
	fieldVal, ok := fieldValue(filter.Field, entry)
	if !ok {
		return false, fmt.Errorf("field '%s' not found", filter.Field)
	}

	if !filter.CaseSensitive {
		fieldVal = strings.ToLower(fieldVal)
	}

	for _, value := range filter.Values {
		if !filter.CaseSensitive {
			value = strings.ToLower(value)
		}

		if strings.Contains(fieldVal, value) {
			return true, nil
		}
	}

	return false, nil
}

func doFilterMatchAll(entry commafeed.Entry, filter config.FeedFilter) (bool, error) {
	fieldVal, ok := fieldValue(filter.Field, entry)
	if !ok {
		return false, fmt.Errorf("field '%s' not found", filter.Field)
	}

	if !filter.CaseSensitive {
		fieldVal = strings.ToLower(fieldVal)
	}

	for _, value := range filter.Values {
		if !filter.CaseSensitive {
			value = strings.ToLower(value)
		}

		if strings.Contains(fieldVal, value) {
			return false, nil
		}
	}

	return true, nil

}

func fieldValue(fieldname string, entry interface{}) (string, bool) {
	obj := reflect.ValueOf(entry)
	for obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}

	field := obj.FieldByName(fieldname)
	if !field.IsValid() {
		return "", false
	}

	for field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return "", false
		}
		field = field.Elem()
	}

	return field.String(), true
}
