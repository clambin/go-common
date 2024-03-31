package charmer

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
)

type Arguments map[string]Argument

type Argument struct {
	Default any
	Help    string
}

// SetPersistentFlagsWithDefaults adds the arguments to cmd as PersistentFlags, binds them to viper and sets up the default values.
func SetPersistentFlagsWithDefaults(cmd *cobra.Command, viper *viper.Viper, args Arguments) error {
	var err error
	if err = SetPersistentFlags(cmd, viper, args); err == nil {
		err = SetDefaults(viper, args)
	}
	return err
}

// SetDefaults sets up the default values in viper.
func SetDefaults(v *viper.Viper, args Arguments) error {
	for name, arg := range args {
		if arg.Default == nil {
			return fmt.Errorf("no default for %s", name)
		}
		v.SetDefault(name, arg.Default)
	}
	return nil
}

// SetPersistentFlags adds the arguments to cmd as PersistentFlags and binds them to viper.
func SetPersistentFlags(cmd *cobra.Command, v *viper.Viper, args Arguments) error {
	for name, arg := range args {
		switch val := arg.Default.(type) {
		case int:
			cmd.PersistentFlags().Int(name, val, arg.Help)
		case float64:
			cmd.PersistentFlags().Float64(name, val, arg.Help)
		case string:
			cmd.PersistentFlags().String(name, val, arg.Help)
		case bool:
			cmd.PersistentFlags().Bool(name, val, arg.Help)
		case time.Duration:
			cmd.PersistentFlags().Duration(name, val, arg.Help)
		default:
			return fmt.Errorf("unsupported type for flag '%s'", name)
		}
		_ = v.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
	}
	return nil
}
