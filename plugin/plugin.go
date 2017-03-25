package plugin

import (
	"github.com/henrylee2cn/rpc2/common"
)

type (
	//IPlugin represents a plugin.
	IPlugin interface {
		Name() string
	}
	// IPluginContainer represents a plugin container that defines all methods to manage plugins.
	IPluginContainer interface {
		Add(plugins ...IPlugin) error
		Remove(pluginName string) error
		GetName(plugin IPlugin) string
		GetByName(pluginName string) IPlugin
		GetAll() []IPlugin
	}
)

// PluginContainer implements IPluginContainer interface.
type PluginContainer struct {
	Plugins []IPlugin
}

// Add adds a plugin.
func (p *PluginContainer) Add(plugins ...IPlugin) error {
	if p.Plugins == nil {
		p.Plugins = make([]IPlugin, 0)
	}
	for _, plugin := range plugins {
		if plugin == nil {
			return common.ErrPluginIsNil
		}
		pName := p.GetName(plugin)
		if pName != "" && p.GetByName(pName) != nil {
			return common.ErrPluginAlreadyExists.Format(pName)
		}
		p.Plugins = append(p.Plugins, plugin)
	}
	return nil
}

// Remove removes a plugin by it's name.
func (p *PluginContainer) Remove(pluginName string) error {
	if p.Plugins == nil {
		return common.ErrPluginRemoveNoPlugins.Return()
	}

	if pluginName == "" {
		//return error: cannot delete an unamed plugin
		return common.ErrPluginRemoveEmptyName.Return()
	}

	indexToRemove := -1
	for i := range p.Plugins {
		if p.GetName(p.Plugins[i]) == pluginName {
			indexToRemove = i
			break
		}
	}
	if indexToRemove == -1 {
		return common.ErrPluginRemoveNotFound.Return()
	}

	p.Plugins = append(p.Plugins[:indexToRemove], p.Plugins[indexToRemove+1:]...)

	return nil
}

// GetName returns the name of a plugin, if no GetName() implemented it returns an empty string ""
func (p *PluginContainer) GetName(plugin IPlugin) string {
	return plugin.Name()
}

// GetByName returns a plugin instance by it's name
func (p *PluginContainer) GetByName(pluginName string) IPlugin {
	if p.Plugins == nil {
		return nil
	}

	for _, plugin := range p.Plugins {
		if plugin.Name() == pluginName {
			return plugin
		}
	}

	return nil
}

// GetAll returns all activated plugins
func (p *PluginContainer) GetAll() []IPlugin {
	return p.Plugins
}
