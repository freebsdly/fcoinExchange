package conf

import (
	"fcoinExchange/model"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

//

//
var (
	defaultPath string = "fcoin.yaml"
	cfg         *config
)

func Init() {
	cfgpath := flag.String("config", defaultPath, "configuration file")
	flag.Parse()

	cfg = new(config)
	cfg.init()
	cfg.SetPath(*cfgpath)

	err := cfg.parse()
	if err != nil {
		fmt.Printf("parse configuration file failed. %s\n", err)
		os.Exit(1)
	}

	go cfg.AutoUpdate()
}

//
func NewConfiguration() *model.Configuration {
	return &model.Configuration{}
}

//
func GetConfiguration() *model.Configuration {
	return cfg.configuration()
}

//
func SetConfigurationFilePath(path string) {
	cfg.SetPath(path)
}

func SetReloadInterval(i int) {
	cfg.intervalChan <- i
}

//

// 定义config用于操作配置文件
type config struct {
	intervalChan chan int
	errorChan    chan error

	interval int
	path     string
	cfg      *model.Configuration
	sync.RWMutex
}

//
func (p *config) init() {
	p.intervalChan = make(chan int, 1)
	p.errorChan = make(chan error, 1000)
	p.cfg = NewConfiguration()
}

//
func (p *config) configuration() *model.Configuration {
	p.RLock()
	defer p.RUnlock()
	return p.cfg
}

//
func (p *config) SetPath(path string) {
	p.Lock()
	defer p.Unlock()
	p.path = path
}

//
func (p *config) SetReloadInterval(i int) {
	p.intervalChan <- i
}

//
func (p *config) parse() error {
	p.RLock()
	path := p.path
	p.RUnlock()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	p.Lock()
	err = yaml.Unmarshal(data, p.cfg)
	p.Unlock()

	return err
}

// 定时读取配置文件
func (p *config) AutoUpdate() {
	if p.cfg.ReloadInterval == 0 {
		return
	}

	p.interval = p.cfg.ReloadInterval

	var (
		i   int
		err error
		t   *time.Ticker = time.NewTicker(time.Duration(p.interval) * time.Millisecond)
	)
	for {
		select {
		case err = <-p.errorChan:
			fmt.Printf("%s", err)
		case i = <-p.intervalChan:
			if i == 0 {
				fmt.Printf("reload interval set to 0, close auto reload\n")
				return
			}

			if i != p.interval {
				fmt.Printf("reload interval set to %d from %d\n", i, p.interval)
				p.interval = i
				t.Stop()
				t = time.NewTicker(time.Duration(p.interval) * time.Millisecond)
			}
		default:
			<-t.C
			err = p.parse()
			if err != nil {
				p.errorChan <- err
			} else {
				p.SetReloadInterval(p.cfg.ReloadInterval)
			}

		}
	}
}
