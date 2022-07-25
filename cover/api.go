package cover

import (
	"bytes"
	"github.com/fvk113/go-tkt-convenios/util"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type Coverage struct {
	baseUrl *string
}

func (o *Coverage) Cover(suite interface{}) {
	value := reflect.ValueOf(suite)
	casesFunc := value.MethodByName("Cases")
	casesValue := casesFunc.Call([]reflect.Value{})[0]
	cases := casesValue.Interface().([]func())
	for _, f := range cases {
		o.executeMethod(f)
	}
}

func (o *Coverage) executeMethod(f func()) {
	def := reflect.ValueOf(f)
	name := runtime.FuncForPC(def.Pointer()).Name()
	if strings.HasSuffix(name, "-fm") {
		name = name[:len(name)-3]
	}
	util.Logger("info").Println(name)
	defer func() {
		if r := recover(); r != nil {
			util.Logger("error").Panicf("Error executing %s", name)
			util.ProcessPanic(r)
		}
	}()
	f()
}

func PostJson(url string, out interface{}, in interface{}) {
	client := http.Client{}
	data := util.Marshal(out)
	response, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	util.CheckErr(err)
	data, err = ioutil.ReadAll(response.Body)
	util.CheckErr(err)
	if response.StatusCode != 200 {
		if len(data) > 0 {
			panic(string(data))
		} else {
			panic(response.Status)
		}
	}
	if in != nil {
		util.JsonDecode(in, bytes.NewBuffer(data))
	}
}
