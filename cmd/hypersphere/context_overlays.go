// Path: cmd/hypersphere/context_overlays.go
// Description: Load endpoint-specific overlays for aliases, plugins, and hotkeys.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const hotkeyRegistryEnvPath = "HYPERSPHERE_HOTKEYS_FILE"

type endpointOverlays struct {
	aliases commandAliasRegistry
	plugins pluginRegistry
	hotkeys map[string]string
}

func loadEndpointOverlays(endpoint string) endpointOverlays {
	overlays := endpointOverlays{
		aliases: commandAliasRegistry{aliases: map[string]string{}},
		plugins: pluginRegistry{entries: []pluginEntry{}},
		hotkeys: map[string]string{},
	}
	overlays.aliases = mergeAliasRegistries(loadAliasRegistryWithOverlay(endpoint))
	overlays.plugins = mergePluginRegistries(loadPluginRegistryWithOverlay(endpoint))
	overlays.hotkeys = mergeHotkeyBindings(loadHotkeyBindingsWithOverlay(endpoint))
	return overlays
}

func loadAliasRegistryWithOverlay(endpoint string) (commandAliasRegistry, commandAliasRegistry) {
	global := commandAliasRegistry{aliases: map[string]string{}}
	overlay := commandAliasRegistry{aliases: map[string]string{}}
	basePath, err := defaultAliasRegistryPath()
	if err != nil {
		return global, overlay
	}
	global, _ = loadCommandAliasRegistry(basePath)
	overlay, _ = loadCommandAliasRegistry(endpointOverlayPath(basePath, endpoint))
	return global, overlay
}

func mergeAliasRegistries(
	global commandAliasRegistry,
	overlay commandAliasRegistry,
) commandAliasRegistry {
	merged := commandAliasRegistry{aliases: map[string]string{}}
	for key, value := range global.aliases {
		merged.aliases[key] = value
	}
	for key, value := range overlay.aliases {
		merged.aliases[key] = value
	}
	return merged
}

func loadPluginRegistryWithOverlay(endpoint string) (pluginRegistry, pluginRegistry) {
	global := pluginRegistry{entries: []pluginEntry{}}
	overlay := pluginRegistry{entries: []pluginEntry{}}
	basePath, err := defaultPluginRegistryPath()
	if err != nil {
		return global, overlay
	}
	global, _ = loadPluginRegistry(basePath)
	overlay, _ = loadPluginRegistry(endpointOverlayPath(basePath, endpoint))
	return global, overlay
}

func mergePluginRegistries(global pluginRegistry, overlay pluginRegistry) pluginRegistry {
	mergedEntries := append([]pluginEntry{}, global.entries...)
	for _, overlayEntry := range overlay.entries {
		mergedEntries = append(mergedEntries, overlayEntry)
	}
	return pluginRegistry{entries: mergedEntries}
}

func loadHotkeyBindingsWithOverlay(endpoint string) (map[string]string, map[string]string) {
	global := map[string]string{}
	overlay := map[string]string{}
	basePath, err := defaultHotkeysPath()
	if err != nil {
		return global, overlay
	}
	global, _ = loadHotkeyBindings(basePath)
	overlay, _ = loadHotkeyBindings(endpointOverlayPath(basePath, endpoint))
	return global, overlay
}

func mergeHotkeyBindings(global map[string]string, overlay map[string]string) map[string]string {
	merged := map[string]string{}
	for key, value := range global {
		merged[key] = value
	}
	for key, value := range overlay {
		merged[key] = value
	}
	return merged
}

func defaultHotkeysPath() (string, error) {
	if override := strings.TrimSpace(os.Getenv(hotkeyRegistryEnvPath)); override != "" {
		return override, nil
	}
	paths, err := infoPaths()
	if err != nil {
		return "", err
	}
	return paths["hotkeys"], nil
}

func loadHotkeyBindings(path string) (map[string]string, error) {
	bindings := map[string]string{}
	if strings.TrimSpace(path) == "" {
		return bindings, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return bindings, nil
		}
		return nil, err
	}
	if strings.TrimSpace(string(content)) == "" {
		return bindings, nil
	}
	if err := json.Unmarshal(content, &bindings); err != nil {
		return nil, err
	}
	return bindings, nil
}

func endpointOverlayPath(basePath string, endpoint string) string {
	ext := filepath.Ext(basePath)
	base := strings.TrimSuffix(basePath, ext)
	normalizedEndpoint := strings.ToLower(strings.TrimSpace(endpoint))
	if normalizedEndpoint == "" {
		return basePath
	}
	return base + "." + normalizedEndpoint + ext
}
