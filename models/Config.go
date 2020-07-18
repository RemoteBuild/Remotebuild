package models

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
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
	Jobs                      jobconfig
	AllowRegistration         bool `default:"false"`
	KeepBuildContainer        bool
	KeepBuildFiles            bool
	DeleteUnusedSessionsAfter time.Duration `default:"10m"`
	Ccache                    ccacheConfig
}

type ccacheConfig struct {
	Dir     string
	MaxSize int
}

type jobconfig struct {
	Images map[string]string
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
						libremotebuild.JobAUR.String(): "jojii/buildaur:v1.2",
					},
				},
				DeleteUnusedSessionsAfter: 10 * time.Minute,
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

	// Print Warning if ccache is not set up properly
	if !config.IsCcacheDirValid() {
		log.Warn("Ccache directory is not valid")
	} else {
		log.Info("Ccache set up correctly! Using ", config.Server.Ccache.MaxSize, "G of diskspace for ccache")
	}

	return true
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

// GetImage get DockerImage for buildType
func (config Config) GetImage(buildType libremotebuild.JobType) (string, bool) {
	v, ok := config.Server.Jobs.Images[buildType.String()]
	return v, ok
}

// IsCcacheDirValid return true if cache is valid
func (config Config) IsCcacheDirValid() bool {
	if config.Server.Ccache.MaxSize == 0 {
		return false
	}

	// Try to create the folder if dir is set
	if len(config.Server.Ccache.Dir) > 0 && !gaw.FileExists(config.Server.Ccache.Dir) {
		err := os.MkdirAll(config.Server.Ccache.Dir, 0700)
		if err != nil {
			log.Warn("Can't create cacche dir:", err)
			return false
		}
	}

	return len(config.Server.Ccache.Dir) > 0 && gaw.FileExists(config.Server.Ccache.Dir)
}
