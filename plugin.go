package rpc2

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
	plugins []IPlugin
}

// Add adds a plugin.
func (p *PluginContainer) Add(plugins ...IPlugin) error {
	if p.plugins == nil {
		p.plugins = make([]IPlugin, 0)
	}
	for _, plugin := range plugins {
		if plugin == nil {
			return ErrPluginIsNil
		}
		pName := p.GetName(plugin)
		if pName != "" && p.GetByName(pName) != nil {
			return ErrPluginAlreadyExists.Format(pName)
		}
		p.plugins = append(p.plugins, plugin)
	}
	return nil
}

// Remove removes a plugin by it's name.
func (p *PluginContainer) Remove(pluginName string) error {
	if p.plugins == nil {
		return ErrPluginRemoveNoPlugins.Return()
	}

	if pluginName == "" {
		//return error: cannot delete an unamed plugin
		return ErrPluginRemoveEmptyName.Return()
	}

	indexToRemove := -1
	for i := range p.plugins {
		if p.GetName(p.plugins[i]) == pluginName {
			indexToRemove = i
			break
		}
	}
	if indexToRemove == -1 {
		return ErrPluginRemoveNotFound.Return()
	}

	p.plugins = append(p.plugins[:indexToRemove], p.plugins[indexToRemove+1:]...)

	return nil
}

// GetName returns the name of a plugin, if no GetName() implemented it returns an empty string ""
func (p *PluginContainer) GetName(plugin IPlugin) string {
	return plugin.Name()
}

// GetByName returns a plugin instance by it's name
func (p *PluginContainer) GetByName(pluginName string) IPlugin {
	if p.plugins == nil {
		return nil
	}

	for _, plugin := range p.plugins {
		if plugin.Name() == pluginName {
			return plugin
		}
	}

	return nil
}

// GetAll returns all activated plugins
func (p *PluginContainer) GetAll() []IPlugin {
	return p.plugins
}
