---
title: What you can't do with reflect
tags: ["Golang", "Reflect"]
date: 2018-06-15
description : So ... you can't do it with reflect? Actually, you can!
---

### What's the problem?

A while ago, I was working on backend application in which we were trying to increase readability of the code, while defining a standard. In my morning explorations I've found a technique (sorry, can't remember where : I'll edit this article when I'll remember) that allowed us to use reflection inside a `before` type handler, thus decoding JSON payloads would be agnostic to controllers.

A typical controller signature looks like this :
```go
func PostAddress(c *context.Context, entity *Address) error 
```

Basically, the `before` handler would both construct the context (in which, for instance, to deposit the currently logged in user) and also decode the JSON payload and provide it as parameter when everything fine. Of course, in the above declaration, a convention was already made : the first parameter would always be context, the second one always a pointer to the desired decoded payload and the function will always have to return an error.

Now, the problem increase in complexity once you would like to receive as URL path parameters, route parameters or HTTP headers. Taking the above example (context aside), a desired controller signature should look like this:
```go
func GetAddressesForUser(HeaderUserAgent string, URLParamUserId, URLVarPage, URLVarPerPage int) (entity []*Address,error)
```

The first parameter would come from HTTP headers, the second one from the route parsing and Page, PerPage should be taken from URL parameters. Nothing special, once you could make a convention that the parameter names would provide the name of what you require, in our example (yes, you've guessed it), it's the start of the parameter name : Header - I need this from headers, URLParam - take this from route and URLVar to be taken from the URL after the question mark.

Well, Go does not provide access to parameter names, even via reflection. Looking for the reason why, I've found this [argument](https://github.com/golang/go/issues/12384#issuecomment-137321421) of Russ Cox: "Reflect provides information about types. Function parameter names are not part of the func type. In contrast, struct field names are. [...] I think it would likely be a mistake to expand the scope of package reflect beyond the clearly defined line of "types". It's not obvious to me where else to draw a line between func parameter names and the entire source files.".

Despite the fact that I don't agree with him, I can't argue that this won't be extremly complicated. Now, how about finding a workaround?

### Solution

The closest you can get to what you want is passing a struct. Create a struct type that wraps your current parameters, and change your function to accept a value of that struct (or a pointer to it). Combining all the required parameters in a struct, for each controller function would imply - even for a small project - a lot of struct declarations. To avoid such I wish that Go would support local struct declarations (that is just syntactic sugar) like this:
```go
func (r *State) MethodA(args struct { Arg1 int; Arg2 string }) error {
   // use args.Arg1, args.Arg2, etc.
}
```
But it doesn't. By this time, I know you are eager to see the code. I will asume that you know your way around with [Gorilla muxer](https://github.com/gorilla/mux), so here is the commented code:

```go
package routers

import (
    "github.com/gorilla/mux"
    "fmt"
    "net/http"
    "reflect"
    "encoding/json"
    "io/ioutil"
    "context"
    "runtime"
    "strings"
    "strconv"
)

type IValidator interface {
	Validate() error
}

type ParamAndIndex struct {
    tag   string
    index int
    isVar bool
}

func collectRequirements(fnValue reflect.Value) (reflect.Type, reflect.Type, reflect.Type, []ParamAndIndex, []string) {

	// checking if we're registering a function, not something else
	functionType := fnValue.Type()
	if functionType.Kind() != reflect.Func {
		panic("Can only register functions.")
	}
	
	// getting the function name (for debugging purposes)
	fnCallerName := runtime.FuncForPC(fnValue.Pointer()).Name()
	parts := strings.Split(fnCallerName, "/")
	callerName := parts[len(parts)-1]

	// collecting injected parameters
	var payloadType reflect.Type
	var paramType reflect.Type
	var headersType reflect.Type
	var paramFields []ParamAndIndex
	var headerFields []string

	if functionType.NumIn() == 0 {
		panic("Handler must have at least one argument : context.Context")
	}
	// convention : first param is always context.Context
	contextParam := functionType.In(0)
	if "context.Context" != contextParam.String() {
		panic("bad handler func : first param should be context.Context")
	}

	for p := 1; p < functionType.NumIn(); p++ {
		param := functionType.In(p)
		paramName := param.Name()
		// param types should have the name starting with "Param" (e.g. "ParamPageAndSomethingElse")
		if strings.HasPrefix(paramName, "Param") {
			paramType = param
			for j := 0; j < param.NumField(); j++ {
				field := param.Field(j)
				// if a field is read from muxer vars, it should have a tag set to the name of the required parameter
				varTag := field.Tag.Get("var")
				// if a field is read from muxer form, it should have a tag set to the name of the required parameter
				formTag := field.Tag.Get("form")
				if len(varTag) > 0 {
					paramFields = append(paramFields, ParamAndIndex{tag: varTag, index: j, isVar: true})
				}

				if len(formTag) > 0 {
					paramFields = append(paramFields, ParamAndIndex{tag: formTag, index: j})
				}
			}
			// convention : Headers mark headers struct (e.g. "HeadersForMe")
		} else if strings.HasPrefix(paramName, "Headers") {
			headersType = param
			// forced add of the "User-Agent" - more can be added here, of course...
			headerFields = append(headerFields, "User-Agent")
			for j := 0; j < param.NumField(); j++ {
				field := param.Field(j)
				// all headers should have hdr tag
				hdrTag := field.Tag.Get("hdr")
				if len(hdrTag) > 0 {
					headerFields = append(headerFields, hdrTag)
				}
			}
		} else {
			if payloadType != nil {
				panic("Seems you are expecting two payloads on " + callerName + ". You should take only one.")
			}
			// convention : second param is always the json payload (which gets automatically decoded)
			switch functionType.In(p).Kind() {
			case reflect.Ptr, reflect.Map, reflect.Slice:
				payloadType = functionType.In(p)
			default:
				fmt.Printf("Second argument must be an *object, map, or slice and it's %q on %s [will be ignored].\n", functionType.In(p).String(), callerName)
			}
		}
	}

	// the function must always return 2 params
	if functionType.NumOut() != 2 {
		panic("Handler has " + strconv.Itoa(functionType.NumOut()) + " returns. Must have two : pointer to something and error. (while registering " + callerName + ")")
	}

	// last param returned must be error
	errorParam := functionType.Out(1)
	if "error" != errorParam.String() {
		panic("bad handler func : last parameter is supposed to be error")
	}

	return payloadType, paramType, headersType, paramFields, headerFields
}

func HandleFunc(router *mux.Router, route string, fn interface{}) *mux.Route {
	// reflect on the provided handler (controller with signature above)
	fnValue := reflect.ValueOf(fn)

	// get payload, parameters and headers that will be injected
	payloadType, paramType, headersType, paramFields, headerFields := collectRequirements(fnValue)

	// sometimes controller expects the request itself - we're providing it
	isRequestInjected := false
	if payloadType != nil && payloadType.Kind() == reflect.Ptr && payloadType.Elem().Name() == "Request" {
		isRequestInjected = true
	}
	// the actual before handler, which collects and build all the informations expected
	return router.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		
		// checking if client has sent us content type
		if len(r.Header["Content-Type"]) == 0{
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("{\"error\":\"missing Content-Type\"}"))
            return
		}
		
		// content type "negociation" - in our case we're dealing with json, but you can extend the functionality after your needs (being crazy like GOB over HTTP :))
		if r.Header["Content-Type"][0] != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("{\"error\":\"unknown Content-Type\"}"))
			return
		}
		
		var err error
        var reqBody []byte
        // only if the controller is not expecting Request itself, we're reading the body
        if !isRequestInjected {
        	// now we read the request body
        	reqBody, err = ioutil.ReadAll(r.Body)
        	if err != nil {
        		w.WriteHeader(http.StatusInternalServerError)
        		w.Write([]byte("{\"error\":\"" + err.Error()+ "\"}"))
        		return
        	}
        	// always defering close
        	defer r.Body.Close()
        }
        
        // starting to build the arguments of calling our handler. First one, the context
        in := []reflect.Value{ reflect.ValueOf(context.Background())}
        
        if payloadType != nil {
        	// Building the deserialize value
        	var deserializeTo reflect.Value
        	switch payloadType.Kind() {
        	case reflect.Slice, reflect.Map:
        		deserializeTo = reflect.New(payloadType)
        		in = append(in, deserializeTo.Elem())
        	case reflect.Ptr:
        		if !isRequestInjected {
        			// the most common scenario - expecting a struct
        			deserializeTo = reflect.New(payloadType.Elem())
        			in = append(in, deserializeTo)
        		    }
            }
            if !isRequestInjected {
        		// json decode the payload
        		if err = json.Unmarshal(reqBody, deserializeTo.Interface()); err != nil {
        			w.WriteHeader(http.StatusBadRequest)
        			w.Write([]byte("{\"error\":\""+fmt.Sprintf("Unmarshal error: %v. Received from client : `%s`", err, string(reqBody))+"\"}"))
                    return
        		}
        		// checking if value is implementing Validate() error
        		iVal, isValidator := deserializeTo.Interface().(IValidator)
        		if isValidator {
        			// it does - we call validate
        			err = iVal.Validate()
        			if err != nil {        				
        				w.WriteHeader(http.StatusBadRequest)
        				w.Write([]byte("{\"error\":\""+fmt.Sprintf("Validation error : %v", err)+"\"}"))
                        return
        			}
        		}
        	} else {
        			// append request as it is, since body is going to be read in controller.
        			in = append(in, reflect.ValueOf(r))
        	}
        }
        
        // we have parameters that need to be injected
        if paramType != nil {
        	vars := mux.Vars(r)
        	p := reflect.New(paramType).Elem()
        	for _, pf := range paramFields {
        		// if the parameter is in muxer vars
        		if pf.isVar {
        			p.Field(pf.index).Set(reflect.ValueOf(vars[pf.tag]))
        		} else {
        			// otherwise it must come from muxer form
        			fv := r.FormValue(pf.tag)
        			p.Field(pf.index).Set(reflect.ValueOf(fv))
        		}
        	}
        	// adding the injected
        	in = append(in, p)
        }
        
        // we have headers that need to be injected
        if headersType != nil {
        	h := reflect.New(headersType).Elem()
        	for idx, hf := range headerFields {
        		switch hf {
        		case "User-Agent":
        			h.Field(idx).Set(reflect.ValueOf(r.UserAgent()))
        		default:
        			h.Field(idx).Set(reflect.ValueOf(r.Header.Get(hf)))
        		}
        	}
        	in = append(in, h)
        }
        
        // finally, we're calling the handler with all the params
        out := fnValue.Call(in)
        
        // processing return of the handler (should be payload, error)
        isError := out[0].IsNil()
        // preparing the json encoder
        enc := json.NewEncoder(w)
        // we have error
        if isError {
        	// header
        	w.Header().Set("Content-Type", "application/json")
            problem, ok := out[1].Interface().(error)
        	if !ok {
        		// should never happen, since the check is done in the collect function. But better safe than sorry.
        		w.WriteHeader(http.StatusInternalServerError)
        		w.Write([]byte("{\"error\":\"not error - second param should be error.\"}"))
                return
        	}
        	w.WriteHeader(http.StatusInternalServerError)
        	w.Write([]byte("{\"error\":\""+problem.Error()+"\"}"))
            return
        } else {
        	
        	// bytes are delivered as they are (since they help you for downloads)
        	if byts, ok := out[0].Interface().([]byte); ok {
        		w.Write(byts)
        		return
        	}
        	// only now we're seting header, so download can work correctly
        	w.Header().Set("Content-Type", "application/json")
                        
        	// no error has occured - serializing payload
        	err := enc.Encode(out[0].Interface())
        	if err != nil {
        		w.WriteHeader(http.StatusInternalServerError)
        		w.Write([]byte("{\"error\":\"encoding payload error : "+err.Error()+"\"}"))
        		return
        	}
        }
	}
}
```

First question you might ask, is where the controller would require the Request itself. Well, this is the case for uploading files, in which the controller itself would have the responsability of reading the request `Body`. Last thing : example of Headers and Params structs:

```go
    type Headers struct {
		UserAgent      string // this doesn't require tag to be `hdr:"User-Agent"`
		AcceptLanguage string `hdr:"Accept-Language"`
	}
	
	type ParamKeyAndDevice struct {
        Key        string `var:"key"` // taken form url e.g. /api/v1/contents/123, where 123 will be the key
        DeviceID   string `form:"id"` // taken from path vars e.g. ?id=something 
        DeviceName string `form:"name"`
        End        string `form:"end"`
    }
```

### End note

Another big limitation on reflection is that while you can use reflection to create new functions (take a look at this [beautiful usage example](https://github.com/golang/go/blob/master/src/net/http/httptrace/trace.go#L197) of `MakeFunc`), there’s no way to create new methods at runtime. 

Unfortunatelly, this means you cannot use reflection to implement an interface at runtime - believe me, I've tried. In Java, this functionality is called a dynamic proxy. 

There is an workaround for this too, a bit ugly if you ask me, but it can be done. I'll write about it in a later post. 