package scopes

// #include <stdlib.h>
// #include "shim.h"
import "C"
import (
	"encoding/json"
	"runtime"
	"unsafe"
)

// Result represents a result from the scope
type Result struct {
	result unsafe.Pointer
}

func finalizeResult(res *Result) {
	if res.result != nil {
		C.destroy_result(res.result)
	}
	res.result = nil
}

// Get returns the named result attribute.
//
// If the attribute does not exist, an error is returned.
func (res *Result) Get(attr string) (interface{}, error) {
	var errorString *C.char = nil
	cData := C.result_get_attr(res.result, unsafe.Pointer(&attr), &errorString)
	if err := checkError(errorString); err != nil {
		return nil, err
	}
	data := C.GoString(cData)
	C.free(unsafe.Pointer(cData))
	var value interface{}
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return nil, err
	}
	return value, nil
}

// Set sets the named result attribute.
//
// An error may be returned if the value can not be stored, or if
// there is any other problems updating the result.
func (res *Result) Set(attr string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	stringValue := string(data)
	var errorString *C.char = nil
	C.result_set_attr(res.result, unsafe.Pointer(&attr), unsafe.Pointer(&stringValue), &errorString)
	return checkError(errorString)
}

// SetInterceptActivation marks this result as needing custom activation handling.
//
// By default, results are activated by the client directly (e.g. by
// running the application associated with the result URI).  For
// results with this flag set though, the scope will be asked to
// perform activation.
func (res *Result) SetInterceptActivation() {
	C.result_set_intercept_activation(res.result)
}

// SetURI sets the "uri" attribute of the result.
func (res *Result) SetURI(uri string) error {
	return res.Set("uri", uri)
}

// SetTitle sets the "title" attribute of the result.
func (res *Result) SetTitle(title string) error {
	return res.Set("title", title)
}

// SetArt sets the "art" attribute of the result.
func (res *Result) SetArt(art string) error {
	return res.Set("art", art)
}

// SetDndURI sets the "dnd_uri" attribute of the result.
func (res *Result) SetDndURI(uri string) error {
	return res.Set("dnd_uri", uri)
}

func (res *Result) getString(attr string) string {
	val, err := res.Get(attr)
	if err != nil {
		return ""
	}
	// If val is not a string, then s will be set to the zero value
	s, _ := val.(string)
	return s
}

// URI returns the "uri" attribute of the result if set, or an empty string.
func (res *Result) URI() string {
	return res.getString("uri")
}

// Title returns the "title" attribute of the result if set, or an empty string.
func (res *Result) Title() string {
	return res.getString("title")
}

// Art returns the "art" attribute of the result if set, or an empty string.
func (res *Result) Art() string {
	return res.getString("art")
}

// DndURI returns the "dnd_uri" attribute of the result if set, or an
// empty string.
func (res *Result) DndURI() string {
	return res.getString("dnd_uri")
}

// CategorisedResult represents a result linked to a particular category.
//
// CategorisedResult embeds Result, so all of its attribute
// manipulation methods can be used on variables of this type.
type CategorisedResult struct {
	Result
}

// NewCategorisedResult creates a new empty result linked to the given
// category.
func NewCategorisedResult(category *Category) *CategorisedResult {
	res := new(CategorisedResult)
	runtime.SetFinalizer(res, finalizeCategorisedResult)
	res.result = C.new_categorised_result(&category.c[0])
	return res
}

func finalizeCategorisedResult(res *CategorisedResult) {
	finalizeResult(&res.Result)
}
