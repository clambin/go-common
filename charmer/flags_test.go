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

func TestSetPersistentFlagsWithDefaults(t *testing.T) {
	var cmd cobra.Command
	v := viper.New()

	if err := charmer.SetPersistentFlagsWithDefaults(&cmd, v, args); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	for name := range args {
		if flag := cmd.PersistentFlags().Lookup(name); flag == nil {
			t.Errorf("Flag %s not found", name)
		}
	}

	for name := range args {
		want := args[name].Default
		got := v.Get(name)
		if want != got {
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

func TestSetDefaults_Invalid_Default(t *testing.T) {
	v := viper.New()
	badArgs := charmer.Arguments{"null": {Default: nil}}
	if err := charmer.SetDefaults(v, badArgs); err == nil {
		t.Error("Expected error, got none")
	}
}
