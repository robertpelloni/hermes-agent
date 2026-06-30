package agent

import "sync"

type Persona struct {
	Name         string
	SystemPrompt string
	Temperature  float32
}

type PersonaManager struct {
	mu       sync.RWMutex
	personas map[string]Persona
	active   string
}

func NewPersonaManager() *PersonaManager {
	pm := &PersonaManager{
		personas: make(map[string]Persona),
	}
	pm.registerDefaults()
	return pm
}

func (pm *PersonaManager) registerDefaults() {
	pm.personas["default"] = Persona{
		Name:         "Default",
		SystemPrompt: "You are a helpful AI assistant.",
		Temperature:  0.7,
	}
	pm.personas["fun_mode"] = Persona{
		Name:         "Fun Mode",
		SystemPrompt: "You are an extremely enthusiastic and creative AI assistant! You love using emojis and making jokes! 🚀🎉",
		Temperature:  0.9,
	}
	pm.personas["security_architect"] = Persona{
		Name:         "Security Architect",
		SystemPrompt: "You are a strict security architect. You analyze all code for vulnerabilities, always prioritizing safety and robust defensive programming practices.",
		Temperature:  0.2,
	}
	pm.active = "default"
}

func (pm *PersonaManager) SetActive(name string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if _, ok := pm.personas[name]; ok {
		pm.active = name
		return true
	}
	return false
}

func (pm *PersonaManager) GetActive() Persona {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.personas[pm.active]
}

func (pm *PersonaManager) List() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	var names []string
	for name := range pm.personas {
		names = append(names, name)
	}
	return names
}
