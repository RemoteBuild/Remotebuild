package models

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/DataManager-Go/DataManagerServer/constants"

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
	Profiling            bool
	MaxHeaderLength      uint  `default:"8000" required:"true"`
	MaxRequestBodyLength int64 `default:"10000" required:"true"`
	MaxUploadFileLength  int64 `default:"1000000000" required:"true"`
	DownloadFileBuffer   int   `default:"100000" required:"true"`
	UserAgentsRawfile    []string
	MaxPreviewFilesize   int64  `default:"50000"`
	HTMLFiles            string `default:"./html/" required:"true"`
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	HTTP                 configHTTPstruct
	HTTPS                configTLSStruct
}

type configServer struct {
	Database                  configDBstruct
	PathConfig                pathConfig
	Roles                     roleConfig
	AllowRegistration         bool          `default:"false"`
	DeleteUnusedSessionsAfter time.Duration `default:"10m"`
}

type roleConfig struct {
	DefaultRole uint `required:"true"`
	Roles       []Role
}

type pathConfig struct {
	FileStore string `required:"true"`
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
					DatabasePort: 3306,
					SSLMode:      "require",
				},
				PathConfig: pathConfig{
					FileStore: "./files",
				},
				AllowRegistration:         false,
				DeleteUnusedSessionsAfter: 10 * time.Minute,
				Roles: roleConfig{
					DefaultRole: 1,
					Roles: []Role{
						{
							ID:                     1,
							RoleName:               "user",
							IsAdmin:                false,
							CreateNamespaces:       true,
							AccesForeignNamespaces: 0,
							MaxURLcontentSize:      5000000,
							MaxUploadFileSize:      10000000000,
						},
						{
							ID:                     2,
							RoleName:               "admin",
							IsAdmin:                true,
							CreateNamespaces:       true,
							AccesForeignNamespaces: 3,
							MaxURLcontentSize:      -1,
							MaxUploadFileSize:      10000000,
						},
					},
				},
			},
			Webserver: webserverConf{
				Profiling: false,
				UserAgentsRawfile: []string{
					"curl",
					"wget",
					"telegrambot",
				},
				MaxPreviewFilesize:   50000,
				HTMLFiles:            "./html",
				MaxRequestBodyLength: 100000,
				MaxUploadFileLength:  10000000000,
				MaxHeaderLength:      8000,
				DownloadFileBuffer:   100000,
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

	if err = configService.Load(&config, confFile); err != nil {
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

	//Check file exists file storage dir
	if !DirExists(config.Server.PathConfig.FileStore) {
		err := os.Mkdir(config.Server.PathConfig.FileStore, 0700)
		if err != nil {
			log.Fatal(err)
			return false
		}
		log.Infof("Filestorage path '%s' created", config.Server.PathConfig.FileStore)
	}

	//Check default role
	if config.GetDefaultRole() == nil {
		log.Fatalln("Can't find default role. You need to specify the ID of the role to use as default")
		return false
	}

	// Check if role can upload more than servers max filesize
	for _, role := range config.Server.Roles.Roles {
		if role.MaxUploadFileSize > config.Webserver.MaxUploadFileLength {
			log.Fatalln("Role has bigger uploadfilesize than server will allow")
			return false
		}
	}

	return true
}

//IsRawUseragent return true if file should be raw depending on useragent
func (config Config) IsRawUseragent(agent string) bool {
	agent = strings.ToLower(agent)
	return gaw.IsInStringArrayContains(agent, config.Webserver.UserAgentsRawfile)
}

//GetStorageFile return the path and file for an uploaded file
func (config Config) GetStorageFile(fileName string) string {
	return path.Join(config.Server.PathConfig.FileStore, fileName)
}

//GetHTMLFile return path of html file
func (config Config) GetHTMLFile(fileName string) string {
	return path.Join(config.Webserver.HTMLFiles, fileName)
}

//GetStaticFile return path of html file
func (config Config) GetStaticFile(fileName string) string {
	return path.Join(config.Webserver.HTMLFiles, "static", fileName)
}

//GetTemplateFile return path of html file
func (config Config) GetTemplateFile(fileName string) string {
	return path.Join(config.Webserver.HTMLFiles, "templates", fileName)
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
