package models

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/JojiiOfficial/Remotebuild/constants"
	"github.com/JojiiOfficial/configService"
	"github.com/JojiiOfficial/gaw"
	log "github.com/sirupsen/logrus"
)

//Config config for the server
type Config struct {
	Server    configServer
	Webserver webserverConf
}

type webserverConf struct {
	MaxHeaderLength      uint  `default:"8000" required:"true"`
	MaxRequestBodyLength int64 `default:"10000" required:"true"`
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	HTTP                 configHTTPstruct
	HTTPS                configTLSStruct
}

type configServer struct {
	Database                  configDBstruct
	Roles                     roleConfig
	Jobs                      jobconfig
	AllowRegistration         bool          `default:"false"`
	DeleteUnusedSessionsAfter time.Duration `default:"10m"`
}

type jobconfig struct {
	Images map[string]string
}

type roleConfig struct {
	DefaultRole uint `required:"true"`
	Roles       []Role
}

type configDBstruct struct {
	Host         string
	Username     string
	Database     string
	Pass         string
	DatabasePort int
	SSLMode      string
}

//Config for HTTPS
type configTLSStruct struct {
	Enabled       bool   `default:"false"`
	ListenAddress string `default:":443"`
	CertFile      string
	KeyFile       string
}

//Config for HTTP
type configHTTPstruct struct {
	Enabled       bool   `default:"false"`
	ListenAddress string `default:":80"`
}

//GetDefaultConfig gets the default config path
func GetDefaultConfig() string {
	return path.Join(constants.DataDir, constants.DefaultConfigFile)
}

//InitConfig inits the config
//Returns true if system should exit
func InitConfig(confFile string, createMode bool) (*Config, bool) {
	var config Config
	if len(confFile) == 0 {
		confFile = GetDefaultConfig()
	}

	s, err := os.Stat(confFile)
	if createMode || err != nil {
		if createMode {
			if s != nil && s.IsDir() {
				log.Fatalln("This name is already taken by a folder")
				return nil, true
			}
			if !strings.HasSuffix(confFile, ".yml") {
				log.Fatalln("The configFile must end with .yml")
				return nil, true
			}
		}

		//Autocreate folder
		path, _ := filepath.Split(confFile)
		_, err := os.Stat(path)
		if err != nil {
			err = os.MkdirAll(path, 0700)
			if err != nil {
				log.Fatalln(err)
				return nil, true
			}
			log.Info("Creating new directory")
		}

		config = Config{
			Server: configServer{
				Database: configDBstruct{
					Host:         "localhost",
					DatabasePort: 5432,
					SSLMode:      "require",
				},
				AllowRegistration: false,
				Jobs: jobconfig{
					Images: map[string]string{
						JobAUR.String(): "jojii/buildaur:v1.0",
					},
				},
				DeleteUnusedSessionsAfter: 10 * time.Minute,
				Roles: roleConfig{
					DefaultRole: 1,
					Roles: []Role{
						{
							ID:       1,
							RoleName: "user",
							IsAdmin:  false,
						},
						{
							ID:       2,
							RoleName: "admin",
							IsAdmin:  true,
						},
					},
				},
			},
			Webserver: webserverConf{
				HTTP: configHTTPstruct{
					Enabled:       true,
					ListenAddress: ":80",
				},
				HTTPS: configTLSStruct{
					Enabled:       false,
					ListenAddress: ":443",
				},
			},
		}
	}

	isDefault, err := configService.SetupConfig(&config, confFile, configService.NoChange)
	if err != nil {
		log.Fatalln(err.Error())
		return nil, true
	}
	if isDefault {
		log.Println("New config created.")
		if createMode {
			log.Println("Exiting")
			return nil, true
		}
	}

	if err = configService.New(&configService.Config{
		AutoReloadCallback: func(config interface{}) {
			log.Info("Config changed")
		},
		AutoReload: true,
	}).Load(&config, confFile); err != nil {
		log.Fatalln(err.Error())
		return nil, true
	}

	return &config, false
}

//Check check the config file of logical errors
func (config *Config) Check() bool {
	if !config.Webserver.HTTP.Enabled && !config.Webserver.HTTPS.Enabled {
		log.Error("You must at least enable one of the server protocols!")
		return false
	}

	if config.Webserver.HTTPS.Enabled {
		if len(config.Webserver.HTTPS.CertFile) == 0 || len(config.Webserver.HTTPS.KeyFile) == 0 {
			log.Error("If you enable TLS you need to set CertFile and KeyFile!")
			return false
		}
		//Check SSL files
		if !gaw.FileExists(config.Webserver.HTTPS.CertFile) {
			log.Error("Can't find the SSL certificate. File not found")
			return false
		}
		if !gaw.FileExists(config.Webserver.HTTPS.KeyFile) {
			log.Error("Can't find the SSL key. File not found")
			return false
		}
	}

	//Check DB port
	if config.Server.Database.DatabasePort < 1 || config.Server.Database.DatabasePort > 65535 {
		log.Errorf("Invalid port for database %d\n", config.Server.Database.DatabasePort)
		return false
	}

	//Check default role
	if config.GetDefaultRole() == nil {
		log.Fatalln("Can't find default role. You need to specify the ID of the role to use as default")
		return false
	}

	return true
}

//GetDefaultRole return the path and file for an uploaded file
func (config Config) GetDefaultRole() *Role {
	for rI, role := range config.Server.Roles.Roles {
		if role.ID == config.Server.Roles.DefaultRole {
			return &config.Server.Roles.Roles[rI]
		}
	}

	return nil
}

// DirExists return true if dir exists
func DirExists(path string) bool {
	s, err := os.Stat(path)
	if err == nil {
		return s.IsDir()
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
