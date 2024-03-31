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

// SetPersistentFlags adds the arguments and their default values to a cobra.Command as PersistentFlags and binds them to viper.
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

// SetDefaults sets up default values in viper.
func SetDefaults(v *viper.Viper, args Arguments) error {
	for name, arg := range args {
		if arg.Default == nil {
			return fmt.Errorf("no default for %s", name)
		}
		v.SetDefault(name, arg.Default)
	}
	return nil
}
