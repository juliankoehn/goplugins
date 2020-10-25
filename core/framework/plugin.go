package framework

import (
	"errors"
	"goplugins/core/routing"
	"os"
	"path/filepath"
	"plugin"
)

type (
	// PluginInitializer returns a Plugin
	PluginInitializer interface {
		Initialize() Plugin
	}
	// Plugin is a extension for our core
	Plugin interface {
		// The Installation Hook
		Install()
		// PostInstall is called after the Installation of the plugin
		PostInstall()
		// Update
		Update()
		// PostUpdate after Update is called
		PostUpdate()
		// Activate enabled the Plugin
		Activate()
		// Deactivate disabled the plugin
		Deactivate()
		// ConfigureRoutes adds routes to our Route Handler
		ConfigureRoutes(*routing.Mux)
	}
)

// ListAvailablePlugins parses the plugins folder
// for available plugins
func ListAvailablePlugins() ([]string, error) {
	var matches []string
	err := filepath.Walk("./plugins", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if matched, err := filepath.Match("*.so", filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return matches, err
}

// InitializePlugin runs all plugin related calls like install and postinstall
func InitializePlugin(pluginPath string, mux *routing.Mux) error {
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return err
	}

	// find func Install
	rawInitializer, err := plug.Lookup("Plugin")
	if err != nil {
		return err
	}

	var extension Plugin
	extension, ok := rawInitializer.(Plugin)
	if !ok {
		return errors.New("could not map initializer to PluginInitializer")
	}

	extension.Install()
	extension.ConfigureRoutes(mux)

	return nil
}
