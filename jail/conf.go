package jail

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"go.sbk.wtf/runj/state"
)

const (
	confName       = "jail.conf"
	configTemplate = `{{ .Name }} {
  path = "{{ .Root }}";
  devfs_ruleset = 4;
  mount.devfs;
  persist;
}
`
)

func CreateConfig(id, root string) (string, error) {
	config, err := renderConfig(id, root)
	if err != nil {
		return "", err
	}
	fmt.Println(config)
	confPath := ConfPath(id)
	confFile, err := os.OpenFile(confPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("jail: config should not already exist: %w", err)
	}
	defer func() {
		confFile.Close()
		if err != nil {
			os.Remove(confFile.Name())
		}
	}()
	_, err = confFile.Write([]byte(config))
	if err != nil {
		return "", err
	}
	return confFile.Name(), nil
}

func ConfPath(id string) string {
	return filepath.Join(state.Dir(id), confName)
}

func renderConfig(id, root string) (string, error) {
	config, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return "", err
	}
	buf := bytes.Buffer{}
	config.Execute(&buf, struct {
		Name string
		Root string
	}{
		Name: id,
		Root: root,
	})
	return buf.String(), nil
}