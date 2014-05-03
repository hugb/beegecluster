package registry

import (
	"sort"
	"sync"

	"github.com/hugb/beegecontroller/config"
	"github.com/hugb/beegecontroller/resource"
)

type Registry struct {
	sync.RWMutex

	config *config.Config

	images     map[string]*resource.Image
	containers map[string]*resource.Container
}

func NewRegistry(c *config.Config) (*Registry, error) {
	r := &Registry{
		config:     c,
		images:     make(map[string]*resource.Image),
		containers: make(map[string]*resource.Container),
	}
	return r, nil
}

func (this *Registry) RegisterImage(id string, image *resource.Image) {
	this.Lock()
	defer this.Unlock()

	this.images[id] = image
}

func (this *Registry) UnregisterImage(id string) {
	this.Lock()
	defer this.Unlock()

	delete(this.images, id)
}

func (this *Registry) GetAllImages() resource.ImageArray {
	this.RLock()
	defer this.RUnlock()

	var images resource.ImageArray
	for _, value := range this.images {
		images = append(images, value)
	}

	sort.Sort(images)

	return images
}

func (this *Registry) LookupImage(id string) (*resource.Image, bool) {
	this.RLock()
	defer this.RUnlock()

	image, ok := this.images[id]
	return image, ok
}

func (this *Registry) GetHostByImageId(id string) string {
	if image, ok := this.LookupImage(id); ok {
		return image.Host
	} else {
		return ""
	}
}

func (this *Registry) RegisterContainer(id string, container *resource.Container) {
	this.Lock()
	defer this.Unlock()

	this.containers[id] = container
	this.containers[id[0:12]] = container
}

func (this *Registry) UnregisterContainer(id string) {
	this.Lock()
	defer this.Unlock()

	delete(this.containers, id)
}

func (this *Registry) GetAllContainers() resource.ContainerArray {
	this.RLock()
	defer this.RUnlock()

	var containers resource.ContainerArray
	for index, value := range this.containers {
		if len(index) == 12 {
			containers = append(containers, value)
		}
	}

	sort.Sort(containers)

	return containers
}

// 根据容器ID得到容器信息
func (this *Registry) LookupContainer(id string) (*resource.Container, bool) {
	this.RLock()
	defer this.RUnlock()

	container, ok := this.containers[id]
	return container, ok
}

// 获取容器所在的主机IP:PORT
func (this *Registry) GetHostByContainerId(id string) string {
	if container, ok := this.LookupContainer(id); ok {
		return container.Host
	} else {
		return ""
	}
}
