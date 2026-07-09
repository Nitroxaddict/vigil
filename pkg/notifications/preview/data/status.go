package data

import wt "github.com/Nitroxaddict/vigil/pkg/types"

type containerStatus struct {
	containerID   wt.ContainerID
	oldImage      wt.ImageID
	newImage      wt.ImageID
	containerName string
	imageName     string
	error
	state State
}

func (u *containerStatus) ID() wt.ContainerID {
	return u.containerID
}

func (u *containerStatus) Name() string {
	return u.containerName
}

func (u *containerStatus) CurrentImageID() wt.ImageID {
	return u.oldImage
}

func (u *containerStatus) LatestImageID() wt.ImageID {
	return u.newImage
}

// LatestImageVersion returns the OCI version label for the latest image.
// The preview data generator does not set version labels, so this always
// returns "" (versionOrID will fall back to ShortID).
func (u *containerStatus) LatestImageVersion() string {
	return ""
}

func (u *containerStatus) ImageName() string {
	return u.imageName
}

func (u *containerStatus) Error() string {
	if u.error == nil {
		return ""
	}
	return u.error.Error()
}

func (u *containerStatus) State() string {
	return string(u.state)
}
