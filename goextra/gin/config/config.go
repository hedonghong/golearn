package config

type Server struct {
	Jwt Jwt `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Zap   Zap   `mapstructure:"zap" json:"zap" yaml:"zap"`
	Redis Redis `mapstructure:"redis" json:"redis" yaml:"redis"`
	Mysql Mysql `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
}

