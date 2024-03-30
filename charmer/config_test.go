package charmer_test

import (
	"github.com/clambin/go-common/charmer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math"
	"testing"
	"time"
)

var args = charmer.Arguments{
	"int":      {Default: 42},
	"float64":  {Default: math.Pi},
	"string":   {Default: "foo"},
	"bool":     {Default: true},
	"duration": {Default: time.Second},
}

func TestCharmer_Config(t *testing.T) {
	var cmd cobra.Command
	v := viper.New()

	if err := charmer.SetDefaults(v, args); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if err := charmer.SetPersistentFlags(&cmd, v, args); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	for name := range args {
		if flag := cmd.PersistentFlags().Lookup(name); flag == nil {
			t.Errorf("Flag %s not found", name)
		}
	}

	if args["int"].Default.(int) != v.Get("int") {
		t.Errorf("Int flag has wrong value: %s", v.Get("int"))
	}
	if args["float64"].Default.(float64) != v.Get("float64") {
		t.Errorf("Float64 flag has wrong value: %s", v.Get("float64"))
	}
	if args["string"].Default.(string) != v.Get("string") {
		t.Errorf("String flag has wrong value: %s", v.Get("string"))
	}
	if args["bool"].Default.(bool) != v.Get("bool") {
		t.Errorf("Bool flag has wrong value: %s", v.Get("bool"))
	}
	if args["duration"].Default.(time.Duration) != v.Get("duration") {
		t.Errorf("Duration flag has wrong value: %s", v.Get("duration"))
	}
}

func TestSetPersistentFlags_Invalid_Type(t *testing.T) {
	var cmd cobra.Command
	v := viper.New()

	badArgs1 := charmer.Arguments{"time": {Default: time.Now()}}
	if err := charmer.SetPersistentFlags(&cmd, v, badArgs1); err == nil {
		t.Error("Expected error, got none")
	}
}

func TestSetDefaults(t *testing.T) {
	v := viper.New()
	badArgs2 := charmer.Arguments{"null": {Default: nil}}
	if err := charmer.SetDefaults(v, badArgs2); err == nil {
		t.Error("Expected error, got none")
	}
}
