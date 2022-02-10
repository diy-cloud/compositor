package multipass

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"

	"github.com/snowmerak/compositor/vm"
	"gopkg.in/yaml.v3"
)

var instance *Multipass = nil

type Multipass struct {
	sync.Mutex
}

func New() vm.VM {
	if instance == nil {
		instance = &Multipass{}
		runtime.SetFinalizer(instance, func(_ *Multipass) {
			if err := commandWithStd("multipass", "purge"); err != nil {
				panic(err)
			}
		})
	}
	return instance
}

func commandWithStd(args ...string) error {
	name := args[0]
	args = args[1:]
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

type data struct {
	Name    string   `json:"name"`
	State   string   `json:"state"`
	IPv4    []string `json:"ipv4"`
	Release string   `json:"release"`
}

func getList() ([]data, error) {
	output, err := commandWithOutput("multipass", "list", "--format", "json")
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	d := struct {
		Datas []data `json:"list"`
	}{}
	if err := decoder.Decode(&d); err != nil {
		return nil, err
	}
	return d.Datas, nil
}

func commandWithOutput(args ...string) ([]byte, error) {
	name := args[0]
	args = args[1:]
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

func (m *Multipass) Create(name string, config *vm.Config) error {
	m.Lock()
	defer m.Unlock()
	err := commandWithStd("multipass", "launch", "--name", name, "--cpus", strconv.FormatInt(config.CPUs, 10), "--mem", config.Memory, "--disk", config.Disk)
	return err
}

func (m *Multipass) Mount(name string, src string, vmName string, dst string) error {
	if err := commandWithStd("multipass", "mount", src, vmName+":"+dst); err != nil {
		return err
	}
	return nil
}

func (m *Multipass) Unmount(name string, dst string) error {
	if err := commandWithStd("multipass", "unmount", dst); err != nil {
		return err
	}
	return nil
}

func (m *Multipass) Start(name string) error {
	if err := commandWithStd("multipass", "start", name); err != nil {
		return err
	}
	return nil
}

func (m *Multipass) Stop(name string) error {
	if err := commandWithStd("multipass", "stop", name); err != nil {
		return err
	}
	return nil
}

func (m *Multipass) Delete(name string) error {
	if err := commandWithStd("multipass", "delete", name); err != nil {
		return err
	}
	return nil
}

func (m *Multipass) List() ([]string, error) {
	d, err := getList()
	if err != nil {
		return nil, err
	}
	l := make([]string, len(d))
	for i, v := range d {
		l[i] = v.Name
	}
	return l, nil
}

func (m *Multipass) IsRunning(name string) (bool, error) {
	d, err := getList()
	if err != nil {
		return false, err
	}
	for _, v := range d {
		if v.Name == name {
			if v.State == "running" {
				return true, nil
			}
		}
	}
	return true, nil
}

func (m *Multipass) IsExist(name string) (bool, error) {
	d, err := getList()
	if err != nil {
		return false, err
	}
	for _, v := range d {
		if v.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (m *Multipass) Info(name string) (vm.Info, error) {
	info := vm.Info{}
	cmd := exec.Command("multipass", "info", name, "--format", "yaml")
	cmd.Stdin = os.Stdin
	output, err := cmd.Output()
	if err != nil {
		return info, err
	}
	if err := yaml.Unmarshal(output, &info); err != nil {
		return info, err
	}
	return info, nil
}

func (m *Multipass) InstanceOf() string {
	return vm.Multipass
}
