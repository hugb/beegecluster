package resource

import ()

type Container struct {
	Host    string
	Created int64
}

type ContainerArray []*Container

func (this ContainerArray) Len() int {
	return len(this)
}

func (this ContainerArray) Less(i, j int) bool {
	return this[i].Created > this[j].Created
}

func (this ContainerArray) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
