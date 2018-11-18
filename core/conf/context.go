package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Context structure handles the parsing of app.yaml
// It has a "preferred" section that is checked first for option queries.
// If the preferred section does not have the option, the DEFAULT section is
// checked fallback.
type Context struct {
	vipers  []*viper.Viper
	section string
}

func LoadContext(confName string, confPaths []string) (*Context, error) {
	ctx := &Context{vipers: []*viper.Viper{}}
	for _, confPath := range confPaths {
		path := filepath.Join(confPath, confName+".yaml")
		if _, err := os.Stat(path); err == nil {
			v := viper.New()
			v.SetConfigType("yaml")
			v.SetConfigName(confName)
			v.AddConfigPath(confPath)
			err := v.ReadInConfig() // Find and read the config file
			if err != nil {         // Handle errors reading the config file
				panic(fmt.Errorf("Fatal error config file: %s \n", err))
			} else {
				fmt.Println("Load conf from path: " + confPath)
			}
			ctx.vipers = append(ctx.vipers, v)
		}
	}

	return ctx, nil
}

func (c *Context) SetSection(section string) {
	c.section = "[" + section + "]"
}

func (c *Context) Get(key string) interface{} {
	for _, v := range c.vipers {
		if v.IsSet(c.section + "." + key) {
			return v.Get(c.section + "." + key)
		}
		if v.IsSet(key) {
			return v.Get(key)
		}
	}
	return nil
}
func (c *Context) GetBool(key string) bool {
	return c.GetBoolDefault(key, false)
}
func (c *Context) GetBoolDefault(key string, value bool) bool {
	for _, v := range c.vipers {
		if v.IsSet(c.section + "." + key) {
			return v.GetBool(c.section + "." + key)
		}
		if v.IsSet(key) {
			return v.GetBool(key)
		}
	}
	return value
}
func (c *Context) GetFloat64(key string) float64 {
	return c.GetFloat64Default(key, 0)
}
func (c *Context) GetFloat64Default(key string, value float64) float64 {
	for _, v := range c.vipers {
		if v.IsSet(c.section + "." + key) {
			return v.GetFloat64(c.section + "." + key)
		}
		if v.IsSet(key) {
			return v.GetFloat64(key)
		}
	}
	return value
}
func (c *Context) GetInt(key string) int {
	return c.GetIntDefault(key, 0)
}
func (c *Context) GetIntDefault(key string, value int) int {
	for _, v := range c.vipers {
		if v.IsSet(c.section + "." + key) {
			return v.GetInt(c.section + "." + key)
		}
		if v.IsSet(key) {
			return v.GetInt(key)
		}
	}
	return value
}
func (c *Context) GetString(key string) string {
	return c.GetStringDefault(key, "")
}
func (c *Context) GetStringDefault(key string, value string) string {
	for _, v := range c.vipers {
		if v.IsSet(c.section + "." + key) {
			return v.GetString(c.section + "." + key)
		}
		if v.IsSet(key) {
			return v.GetString(key)
		}
	}
	return value
}
func (c *Context) GetStringMap(key string) map[string]interface{} {
	return c.GetStringMapDefault(key, map[string]interface{}{})
}
func (c *Context) GetStringMapDefault(key string, value map[string]interface{}) map[string]interface{} {
	for _, v := range c.vipers {
		if v.IsSet(c.section + "." + key) {
			return v.GetStringMap(c.section + "." + key)
		}
		if v.IsSet(key) {
			return v.GetStringMap(key)
		}
	}
	return value
}
func (c *Context) GetStringMapString(key string) map[string]string {
	return c.GetStringMapStringDefault(key, map[string]string{})
}
func (c *Context) GetStringMapStringDefault(key string, value map[string]string) map[string]string {
	for _, v := range c.vipers {
		if v.IsSet(c.section + "." + key) {
			return v.GetStringMapString(c.section + "." + key)
		}
		if v.IsSet(key) {
			return v.GetStringMapString(key)
		}
	}
	return value
}
func (c *Context) GetStringSlice(key string) []string {
	return c.GetStringSliceDefault(key, []string{})
}
func (c *Context) GetStringSliceDefault(key string, value []string) []string {
	for _, v := range c.vipers {
		if v.IsSet(c.section + "." + key) {
			return v.GetStringSlice(c.section + "." + key)
		}
		if v.IsSet(key) {
			return v.GetStringSlice(key)
		}
	}
	return value
}
func (c *Context) GetTime(key string) time.Time {
	return c.GetTimeDefault(key, time.Now())
}
func (c *Context) GetTimeDefault(key string, value time.Time) time.Time {
	for _, v := range c.vipers {
		if v.IsSet(c.section + "." + key) {
			return v.GetTime(c.section + "." + key)
		}
		if v.IsSet(key) {
			return v.GetTime(key)
		}
	}
	return value
}
func (c *Context) GetDuration(key string) time.Duration {
	return c.GetDurationDefault(key, time.Duration(0))
}
func (c *Context) GetDurationDefault(key string, value time.Duration) time.Duration {
	for _, v := range c.vipers {
		if v.IsSet(c.section + "." + key) {
			return v.GetDuration(c.section + "." + key)
		}
		if v.IsSet(key) {
			return v.GetDuration(key)
		}
	}
	return value
}
func (c *Context) IsSet(key string) bool {
	for _, v := range c.vipers {
		if v.IsSet(c.section + "." + key) {
			return true
		}
		if v.IsSet(key) {
			return true
		}
	}
	return false
}
