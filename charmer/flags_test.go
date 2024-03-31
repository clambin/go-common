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

func TestSetPersistentFlags(t *testing.T) {
	var cmd cobra.Command
	v := viper.New()

	if err := charmer.SetPersistentFlags(&cmd, v, args); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	for name := range args {
		if flag := cmd.PersistentFlags().Lookup(name); flag == nil {
			t.Errorf("Flag %s not found", name)
		}
	}

	for name := range args {
		var got any
		switch args[name].Default.(type) {
		case int:
			got = v.GetInt(name)
		case float64:
			got = v.GetFloat64(name)
		case bool:
			got = v.GetBool(name)
		case string:
			got = v.GetString(name)
		case time.Duration:
			got = v.GetDuration(name)
		}
		if want := args[name].Default; want != got {
			t.Errorf("Flag %s had incorrect default value. want %s, got %s", name, want, got)
		}
	}
}

func TestSetPersistentFlags_Invalid_Type(t *testing.T) {
	var cmd cobra.Command
	v := viper.New()
	badArgs := charmer.Arguments{"time": {Default: time.Now()}}
	if err := charmer.SetPersistentFlags(&cmd, v, badArgs); err == nil {
		t.Error("Expected error, got none")
	}
}

func TestSetDefaults(t *testing.T) {
	v := viper.New()
	if got := v.GetString("string"); got != "" {
		t.Fatalf("Unexpected default value: %v", got)
	}

	if err := charmer.SetDefaults(v, args); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if got := v.GetString("string"); got != "foo" {
		t.Errorf("Unexpected default value: %v", got)
	}
}

func TestSetDefaults_Invalid_Default(t *testing.T) {
	v := viper.New()
	badArgs := charmer.Arguments{"null": {Default: nil}}
	if err := charmer.SetDefaults(v, badArgs); err == nil {
		t.Error("Expected error, got none")
	}
}
