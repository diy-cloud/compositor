package vm

type VM interface {
	Create(name string, config *Config) error
	Mount(name string, src string, vmName string, dst string) error
	Unmount(name string, dst string) error
	Start(name string) error
	Stop(name string) error
	Delete(name string) error
	List() ([]string, error)
	IsRunning(name string) (bool, error)
	IsExist(name string) (bool, error)
	Info(name string) (Info, error)
	InstanceOf() string
}

type Config struct {
	Name   string
	CPUs   int64
	Memory string
	Disk   string
}

type Info struct {
	Errors    []string  `yaml:"errors"`
	ImageHash string    `yaml:"image_hash"`
	Release   string    `yaml:"release"`
	Load      []float64 `yaml:"load"`
	Disks     map[string]struct {
		Usage int64 `yaml:"usage"`
		Total int64 `yaml:"total"`
	} `yaml:"disks"`
	IPv4   []string `yaml:"ipv4"`
	Mounts map[string]struct {
		UidMappings []string `yaml:"uid_mappings"`
		GidMappings []string `yaml:"gid_mappings"`
		SourcePath  string   `yaml:"source_path"`
	}
}
