package container

type Image interface {
	Pull()
}

type Container interface {
	Start()
	Stop()
	Remove()
}
