package resource

import ()

type Image struct {
	Host    string
	Created int64
}

type ImageArray []*Image

func (this ImageArray) Len() int {
	return len(this)
}

func (this ImageArray) Less(i, j int) bool {
	return this[i].Created > this[j].Created
}

func (this ImageArray) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
